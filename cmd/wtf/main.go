// wtf — терминальный sysadmin-агент.
//
// Запуск:
//
//	wtf <вопрос>             запустить диагностику
//	cat err.log | wtf <q>    тоже самое + содержимое stdin как контекст
//	wtf config               настройка провайдера/ключа/языка
//	wtf version              версия
//
// Никаких shell-хуков, обёрток вокруг команд, захвата stdout. Юзер описывает
// проблему словами, агент сам выполняет диагностические команды (read-only)
// и предлагает действия. Destructive-команды (sudo/install/restart) только
// показываются — выполняет их юзер вручную.
package main

import (
	"bufio"
	stdctx "context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"

	"github.com/bitcoff/wtf/internal/agent"
	"github.com/bitcoff/wtf/internal/config"
	wctx "github.com/bitcoff/wtf/internal/context"
	"github.com/bitcoff/wtf/internal/memory"
	"github.com/bitcoff/wtf/internal/provider"
	"github.com/bitcoff/wtf/internal/render"
	"github.com/bitcoff/wtf/internal/ui"
)

// version подменяется через -ldflags при release-сборке.
var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "config":
			runConfig(os.Args[2:])
			return
		case "memory":
			runMemoryCmd(os.Args[2:])
			return
		case "version", "--version", "-v":
			fmt.Println("wtf", version)
			return
		case "help", "--help", "-h":
			printHelp()
			return
		}
	}
	runAgent(os.Args[1:])
}

func printHelp() {
	fmt.Print(`wtf — терминальный sysadmin-агент

Использование:
  wtf <вопрос>                 запустить диагностику
  cat error.log | wtf <вопрос>  диагностика + содержимое stdin как контекст
  wtf                          интерактив: запросит вопрос

Опции:
  --provider <name>            claude | openai | gemini
  --lang <ru|en>               язык ответа (по умолчанию из config)
  -v, --verbose                показывать полный вывод диагностических команд

Команды:
  wtf config                   настройка провайдера, ключей, моделей
  wtf config show              показать текущий конфиг
  wtf config set k=v           задать значение
  wtf memory show              показать что агент о тебе помнит
  wtf memory clear             стереть память
  wtf version                  версия

Примеры:
  wtf nginx не стартует
  wtf почему медленно работает диск
  journalctl -u nginx | wtf что не так
`)
}

// runAgent — основной режим. Парсит позиционный промпт, запускает agent.Run.
func runAgent(args []string) {
	// Флаги парсим вручную чтобы не съесть слова из промпта.
	// Правило: если токен — известный флаг (с --), он снимается; иначе уходит в промпт.
	var providerFlag, langFlag string
	var verbose bool
	var passthrough []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--provider" && i+1 < len(args):
			providerFlag = args[i+1]
			i++
		case strings.HasPrefix(a, "--provider="):
			providerFlag = strings.TrimPrefix(a, "--provider=")
		case a == "--lang" && i+1 < len(args):
			langFlag = args[i+1]
			i++
		case strings.HasPrefix(a, "--lang="):
			langFlag = strings.TrimPrefix(a, "--lang=")
		case a == "--verbose" || a == "-v":
			verbose = true
		default:
			passthrough = append(passthrough, a)
		}
	}
	question := strings.TrimSpace(strings.Join(passthrough, " "))

	cfg, err := config.Load()
	if err != nil {
		ui.Err(fmt.Sprintf("config: %v", err))
		os.Exit(1)
	}

	if !anyKeyConfigured(cfg) {
		ui.Banner("wtf — sysadmin-агент", "первый запуск — настроим провайдера за минуту")
		runWizard(cfg)
		fmt.Fprintln(os.Stderr)
	}

	prov := cfg.DefaultProvider
	if providerFlag != "" {
		prov = config.Provider(providerFlag)
	}
	lang := cfg.Language
	if langFlag != "" {
		lang = langFlag
	}

	piped := readPipedStdin()

	// Если нет ни вопроса, ни pipe — спрашиваем интерактивно.
	if question == "" && piped == "" {
		question = promptForQuestion()
		if question == "" {
			ui.Err("пустой вопрос — выход")
			os.Exit(1)
		}
	}

	cli, err := provider.New(cfg, prov)
	if err != nil {
		ui.Err(fmt.Sprintf("провайдер %s недоступен: %v", prov, err))
		ui.Info("настрой ключ: wtf config")
		os.Exit(2)
	}

	store, err := memory.Load()
	if err != nil {
		ui.Warn(fmt.Sprintf("память не загрузилась: %v (продолжаю без неё)", err))
		store = &memory.Store{}
	}

	envInfo := wctx.Collect()

	// NotifyContext — отменяет ctx по SIGINT/SIGTERM; agent видит ctx.Done()
	// и аккуратно завершает текущий HTTP-запрос. После этого мы успеем сохранить
	// notes (если уже были) и память. Жёсткий таймаут 5 минут как backup —
	// чтобы случайно не зависнуть в бесконечном retry.
	ctx, cancel := stdctx.WithTimeout(stdctx.Background(), 5*time.Minute)
	defer cancel()
	ctx, stopSignal := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stopSignal()

	io := &consoleIO{verbose: verbose}

	res, runErr := agent.Run(ctx, cli, envInfo, store, io, question, piped, lang)

	// Сохраняем что есть, даже если агент прерван — это и есть смысл graceful shutdown.
	if res != nil {
		// Fallback: модели часто игнорируют notes несмотря на инструкции.
		// Если ответ дан, но notes пусты — записываем минимум сами,
		// чтобы память не оставалась пустой после успешной сессии.
		if res.Stopped == "finish" && len(res.Notes) == 0 && res.Summary != "" && question != "" {
			res.Notes = append(res.Notes, memory.Entry{
				Type:    memory.TypeResolvedIssue,
				Key:     fallbackIssueKey(question),
				Content: truncate(question+" → "+firstLine(res.Summary), 200),
				TTLDays: 30,
			})
		}
		for _, n := range res.Notes {
			store.Add(n)
		}
	}
	needConsolidation := store.MarkSession()

	if err := store.Save(); err != nil {
		ui.Warn(fmt.Sprintf("не удалось сохранить память: %v", err))
	}

	if runErr != nil {
		// Прерывание Ctrl+C — нормальный выход кодом 130 (SIGINT convention).
		if errors.Is(runErr, stdctx.Canceled) {
			ui.Info("прервано пользователем — память сохранена")
			os.Exit(130)
		}
		ui.Err(fmt.Sprintf("агент: %v", runErr))
		os.Exit(1)
	}

	// Консолидация — best-effort. Если упала — не страшно, в следующий раз попробуем.
	if needConsolidation {
		ui.Info("сжимаю память...")
		consCtx, ccancel := stdctx.WithTimeout(stdctx.Background(), 60*time.Second)
		if err := agent.Consolidate(consCtx, cli, store); err != nil {
			ui.Warn(fmt.Sprintf("консолидация: %v", err))
		} else {
			_ = store.Save()
		}
		ccancel()
	}

	if res.Stopped == "max_iterations" {
		ui.Warn(fmt.Sprintf("достигнут лимит раундов диагностики (%d)", agent.MaxIterations))
	}
}

// readPipedStdin — если stdin не TTY, читаем всё содержимое.
// Используется для `cat err.log | wtf что не так`.
func readPipedStdin() string {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		return ""
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return ""
	}
	s := strings.TrimSpace(string(data))
	// Лимит 16 КБ — больше всё равно бесполезно, AI будет тонуть в шуме.
	const maxStdin = 16 * 1024
	if len(s) > maxStdin {
		s = s[len(s)-maxStdin:]
	}
	return s
}

// promptForQuestion — fallback когда юзер запустил `wtf` без аргументов
// и без pipe. Просим описать проблему.
func promptForQuestion() string {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return ""
	}
	r := bufio.NewReader(os.Stdin)
	return ui.Prompt(r, "опиши проблему", "")
}

// consoleIO — реализация agent.IO для обычного терминала.
//
// Два режима:
//   - quiet (по умолчанию): команда сворачивается в одну строку со спиннером,
//     полный вывод не печатается (но скармливается AI).
//   - verbose (-v / --verbose): полные блоки команды и вывода как раньше.
type consoleIO struct {
	verbose       bool
	sp            *ui.Spinner // спиннер для AI-раздумий
	cmdSp         *ui.Spinner // спиннер во время выполнения команды (quiet)
	pendingReason string      // reason запущенной команды — печатаем при результате
	pendingCmd    string      // команда — то же самое
}

func (c *consoleIO) Thinking(label string) {
	c.sp = ui.NewSpinner(label)
	c.sp.Start()
}

func (c *consoleIO) StopThinking(success bool, _ string) {
	if c.sp == nil {
		return
	}
	if success {
		c.sp.StopOK("")
	} else {
		c.sp.StopFail("")
	}
	c.sp = nil
}

func (c *consoleIO) StepCommand(reason, command string) {
	if c.verbose {
		ui.CommandHeader(reason, command)
		return
	}
	// Quiet-режим: вместо заголовка крутим спиннер с reason+командой.
	c.pendingReason = reason
	c.pendingCmd = command
	label := command
	if reason != "" {
		label = reason + " · " + command
	}
	c.cmdSp = ui.NewSpinner(label)
	c.cmdSp.Start()
}

func (c *consoleIO) CommandOutput(_ string, output string, exit int, dur time.Duration, timedOut bool) {
	if c.verbose {
		ui.CommandResult(output, exit, dur, timedOut)
		return
	}
	// Quiet: останавливаем спиннер, печатаем итог одной строкой.
	if c.cmdSp != nil {
		c.cmdSp.StopOK("")
		c.cmdSp = nil
	}
	ui.CommandLineQuiet(c.pendingReason, c.pendingCmd, output, exit, dur, timedOut)
	c.pendingReason = ""
	c.pendingCmd = ""
}

func (c *consoleIO) UserCommand(reason, command string) {
	ui.UserCommandBlock(reason, command)
}

func (c *consoleIO) Refused(command, reason string) {
	ui.RefusedBlock(command, reason)
}

func (c *consoleIO) Final(summary string) {
	ui.FinalBlock(render.Markdown(summary, ui.FinalBodyWidth()))
}

// RateLimitWait вызывается когда провайдер вернул 429 и retry будет повторён
// автоматически. Останавливаем все активные спиннеры (иначе их строки будут
// перебивать сообщение), показываем юзеру что происходит, и не пытаемся
// возобновить — agent сам стартует следующий Thinking при retry.
func (c *consoleIO) RateLimitWait(wait time.Duration, attempt int) {
	if c.sp != nil {
		c.sp.StopFail("")
		c.sp = nil
	}
	if c.cmdSp != nil {
		c.cmdSp.StopFail("")
		c.cmdSp = nil
	}
	ui.Warn(fmt.Sprintf("rate limit · повтор через %s (попытка %d/%d)",
		wait.Round(100*time.Millisecond), attempt+1, 4))
}

func anyKeyConfigured(cfg *config.Config) bool {
	for _, p := range []config.Provider{config.ProviderClaude, config.ProviderOpenAI, config.ProviderGemini} {
		if cfg.APIKey(p) != "" {
			return true
		}
	}
	return false
}

// === wtf config ===

func runConfig(args []string) {
	cfg, err := config.Load()
	if err != nil {
		ui.Err(fmt.Sprintf("config: %v", err))
		os.Exit(1)
	}
	if len(args) == 0 {
		ui.Banner("wtf · настройка", "")
		runWizard(cfg)
		return
	}
	switch args[0] {
	case "show":
		showConfig(cfg)
	case "set":
		if len(args) < 2 {
			ui.Err("usage: wtf config set <key>=<value>")
			os.Exit(1)
		}
		setConfig(cfg, args[1])
	default:
		ui.Err(fmt.Sprintf("неизвестная подкоманда: %s", args[0]))
		os.Exit(1)
	}
}

func showConfig(cfg *config.Config) {
	ui.Section("конфиг")
	ui.KV("provider", string(cfg.DefaultProvider))
	ui.KV("language", cfg.Language)
	fmt.Fprintln(os.Stderr)
	ui.Section("провайдеры")
	for _, p := range []config.Provider{config.ProviderClaude, config.ProviderOpenAI, config.ProviderGemini} {
		key := cfg.APIKey(p)
		mask := "(не задан)"
		if key != "" {
			mask = mask4(key)
		}
		ui.KV(string(p), fmt.Sprintf("model=%s  key=%s", cfg.Model(p), mask))
	}
}

func setConfig(cfg *config.Config, kv string) {
	k, v, ok := strings.Cut(kv, "=")
	if !ok {
		ui.Err("формат: key=value")
		os.Exit(1)
	}
	k = strings.TrimSpace(k)
	v = strings.TrimSpace(v)
	switch {
	case k == "provider", k == "default_provider":
		cfg.DefaultProvider = config.Provider(v)
	case k == "language", k == "lang":
		cfg.Language = v
	case strings.HasPrefix(k, "claude."), strings.HasPrefix(k, "openai."), strings.HasPrefix(k, "gemini."):
		parts := strings.SplitN(k, ".", 2)
		p := config.Provider(parts[0])
		field := parts[1]
		pc := cfg.Providers[p]
		switch field {
		case "key", "api_key":
			pc.APIKey = v
		case "model":
			pc.Model = v
		default:
			ui.Err(fmt.Sprintf("неизвестное поле: %s", field))
			os.Exit(1)
		}
		cfg.Providers[p] = pc
	default:
		ui.Err(fmt.Sprintf("неизвестный ключ: %s", k))
		os.Exit(1)
	}
	if err := cfg.Save(); err != nil {
		ui.Err(fmt.Sprintf("save: %v", err))
		os.Exit(1)
	}
	ui.OK("сохранено")
}

func runWizard(cfg *config.Config) {
	r := bufio.NewReader(os.Stdin)

	ui.Section("общие настройки")
	cfg.DefaultProvider = config.Provider(ui.Choice(r, "Провайдер по умолчанию",
		[]string{"claude", "openai", "gemini"}, string(cfg.DefaultProvider)))
	cfg.Language = ui.Choice(r, "Язык ответа",
		[]string{"ru", "en"}, cfg.Language)

	configureProvider(r, cfg, cfg.DefaultProvider)

	allProviders := []config.Provider{config.ProviderClaude, config.ProviderOpenAI, config.ProviderGemini}
	for {
		var remaining []string
		for _, p := range allProviders {
			if p == cfg.DefaultProvider {
				continue
			}
			if cfg.APIKey(p) == "" {
				remaining = append(remaining, string(p))
			}
		}
		if len(remaining) == 0 {
			break
		}
		add := ui.Choice(r, "Добавить ещё одного провайдера? (можно переключаться через --provider)",
			append([]string{"нет"}, remaining...), "нет")
		if add == "нет" || add == "" {
			break
		}
		configureProvider(r, cfg, config.Provider(add))
	}

	if err := cfg.Save(); err != nil {
		ui.Err(fmt.Sprintf("save: %v", err))
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr)
	ui.OK("конфиг сохранён в ~/.wtf/config.json")

	if !anyKeyConfigured(cfg) {
		ui.Warn("ни одного ключа не задано — wtf не сможет работать")
	}
}

func configureProvider(r *bufio.Reader, cfg *config.Config, p config.Provider) {
	ui.Section(string(p))

	curKey := cfg.APIKey(p)
	mask := "не задан, Enter — пропустить"
	if curKey != "" {
		mask = mask4(curKey) + ", Enter — оставить"
	}
	newKey := ui.Prompt(r, fmt.Sprintf("API key для %s", p), mask)
	if newKey != "" && newKey != mask {
		pc := cfg.Providers[p]
		pc.APIKey = newKey
		cfg.Providers[p] = pc
		ui.OK("ключ сохранён локально (~/.wtf/config.json, mode 0600)")
	}

	curModel := cfg.Model(p)
	known := config.KnownModels[p]
	newModel := ui.ChoiceOrCustom(r, fmt.Sprintf("Модель для %s", p), known, curModel)
	if newModel != "" && newModel != curModel {
		pc := cfg.Providers[p]
		pc.Model = newModel
		cfg.Providers[p] = pc
	}
}

// === wtf memory ===

func runMemoryCmd(args []string) {
	if len(args) == 0 {
		ui.Err("usage: wtf memory show|clear")
		os.Exit(1)
	}
	store, err := memory.Load()
	if err != nil {
		ui.Err(fmt.Sprintf("memory: %v", err))
		os.Exit(1)
	}
	switch args[0] {
	case "show":
		if len(store.Entries) == 0 {
			ui.Info("память пуста")
			return
		}
		fmt.Println(store.SystemContext(0))
		fmt.Printf("\n[всего записей: %d, сессий: %d]\n", len(store.Entries), store.SessionCount)
	case "clear":
		store.Entries = nil
		store.SessionCount = 0
		if err := store.Save(); err != nil {
			ui.Err(fmt.Sprintf("save: %v", err))
			os.Exit(1)
		}
		ui.OK("память очищена")
	default:
		ui.Err(fmt.Sprintf("неизвестная подкоманда: %s", args[0]))
		os.Exit(1)
	}
}

func mask4(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "…" + s[len(s)-4:]
}

// fallbackIssueKey строит стабильный ключ для авто-заметки из вопроса юзера.
// Берём первые слова, нормализуем — нужно для дедупа: повторный wtf про ту же
// проблему перезапишет старую запись, а не плодит копии.
func fallbackIssueKey(question string) string {
	words := strings.Fields(strings.ToLower(question))
	if len(words) > 6 {
		words = words[:6]
	}
	return "issue_" + strings.Join(words, "_")
}

// firstLine — первая непустая строка текста.
func firstLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		t := strings.TrimSpace(line)
		if t != "" {
			return t
		}
	}
	return s
}

// truncate обрезает строку до n рун (не байт — иначе режется UTF-8 на полуслове).
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}
