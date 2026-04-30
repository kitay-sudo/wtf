package main

import (
	"bufio"
	stdctx "context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/bitcoff/wtf/internal/cache"
	"github.com/bitcoff/wtf/internal/config"
	wctx "github.com/bitcoff/wtf/internal/context"
	"github.com/bitcoff/wtf/internal/provider"
	"github.com/bitcoff/wtf/internal/redact"
	"github.com/bitcoff/wtf/internal/render"
	"github.com/bitcoff/wtf/internal/shellhook"
	"github.com/bitcoff/wtf/internal/ui"
)

// version is overridden at build time via:
//   go build -ldflags "-X main.version=v0.1.0"
// release.yml passes the git tag here.
var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			runInit(os.Args[2:])
			return
		case "config":
			runConfig(os.Args[2:])
			return
		case "version", "--version", "-v":
			fmt.Println("wtf", version)
			return
		case "help", "--help", "-h":
			printHelp()
			return
		}
	}
	runExplain(os.Args[1:])
}

func printHelp() {
	fmt.Print(`wtf [!?] — AI-объяснялка для терминала

Использование:
  wtfc <команда>             запустить команду с захватом вывода, потом 'wtf'
  <команда> 2>&1 | wtf       запайпить вывод напрямую (без префикса)
  wtf --explain "<text>"     объяснить переданный текст
  wtf --rerun                перезапустить последнюю команду
  wtf                        объяснить последний захваченный вывод

  wtf --provider <name>      разово выбрать провайдера (claude|openai|gemini)
  wtf --no-cache             не использовать кеш
  wtf --lang <ru|en>         язык ответа

Команды:
  wtf init [shell]           установить shell-хук (bash|zsh|fish|powershell)
  wtf config                 интерактивная настройка (провайдер, ключи, модели)
  wtf config show            показать текущий конфиг
  wtf config set k=v         задать значение
  wtf version                версия
`)
}

func runExplain(args []string) {
	fs := flag.NewFlagSet("explain", flag.ExitOnError)
	rerun := fs.Bool("rerun", false, "перезапустить последнюю команду")
	explain := fs.String("explain", "", "явный текст ошибки")
	providerFlag := fs.String("provider", "", "claude|openai|gemini")
	noCache := fs.Bool("no-cache", false, "не использовать кеш")
	lang := fs.String("lang", "", "ru|en")
	_ = fs.Parse(args)

	cfg, err := config.Load()
	if err != nil {
		ui.Err(fmt.Sprintf("config: %v", err))
		os.Exit(1)
	}

	// First run: запускаем wizard если ни у одного провайдера нет ключа
	if !anyKeyConfigured(cfg) {
		ui.Banner("wtf — объяснялка ошибок", "первый запуск — настроим за минуту")
		runWizard(cfg)
		fmt.Fprintln(os.Stderr)
	}

	prov := cfg.DefaultProvider
	if *providerFlag != "" {
		prov = config.Provider(*providerFlag)
	}
	language := cfg.Language
	if *lang != "" {
		language = *lang
	}

	var cap *shellhook.Capture
	switch {
	case *explain != "":
		cap = &shellhook.Capture{Output: *explain, Source: "flag"}
	case hasPipedStdin():
		// `<команда> 2>&1 | wtf` — самый надёжный способ передать вывод.
		// Если stdin не-TTY но пустой (например, `wtf </dev/null` или запуск
		// из cron) — не выходим с ошибкой, проваливаемся в обычный flow
		// чтения хука, как будто стдина не было.
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			ui.Err(fmt.Sprintf("чтение stdin: %v", err))
			os.Exit(1)
		}
		if len(strings.TrimSpace(string(data))) == 0 {
			c, hookErr := shellhook.ReadCapture()
			if hookErr != nil {
				ui.Warn(hookErr.Error())
				ui.Info("совет: установи хук — wtf init")
				ui.Info("или запайпь вывод: <команда> 2>&1 | wtf")
				os.Exit(2)
			}
			cap = c
			break
		}
		cap = &shellhook.Capture{Output: string(data), Source: "pipe"}
	case *rerun:
		meta, _ := shellhook.ReadCapture()
		if meta == nil || meta.Command == "" {
			ui.Err("нечего перезапускать — нет последней команды")
			os.Exit(1)
		}
		ui.Step(fmt.Sprintf("перезапуск: %s", meta.Command))
		c, err := shellhook.Rerun(meta.Command)
		if err != nil {
			ui.Err(fmt.Sprintf("rerun: %v", err))
			os.Exit(1)
		}
		cap = c
	default:
		c, err := shellhook.ReadCapture()
		if err != nil {
			ui.Warn(err.Error())
			ui.Info("совет: установи хук — wtf init")
			ui.Info("или: wtf --rerun (перезапустит последнюю команду)")
			ui.Info("или: wtf --explain \"<текст ошибки>\"")
			os.Exit(2)
		}
		cap = c
	}

	if strings.TrimSpace(cap.Output) == "" && cap.Command != "" {
		// Хук пока пишет только metadata (cmd + exit). Сам stdout/stderr
		// он не сохраняет — на это нужен `script(1)`-обёртка, она в roadmap.
		// Поэтому здесь — внятное сообщение, а не silent rerun: rerun небезопасен
		// для side-effect команд (rm, kubectl apply, git push) и виснет на pager'ах.
		ui.Warn("вывод последней команды не захвачен")
		ui.Info(fmt.Sprintf("команда: %s", cap.Command))
		fmt.Fprintln(os.Stderr)
		ui.Info("варианты:")
		fmt.Fprintln(os.Stderr, "    wtf --rerun                       перезапустить последнюю команду (если она безопасна)")
		fmt.Fprintln(os.Stderr, "    <команда> 2>&1 | wtf              запайпить вывод напрямую")
		fmt.Fprintln(os.Stderr, "    wtf --explain \"<текст>\"           передать текст вручную")
		os.Exit(2)
	}

	info := wctx.Collect()
	red := redact.Apply(cap.Output)

	if !cfg.RedactionShown {
		showFirstRunNotice(cfg, red, info, cap)
	} else if summary := redact.Summary(red); summary != "" {
		ui.Info(summary)
	}

	req := provider.Request{
		Language:    language,
		OS:          info.OS,
		Shell:       info.Shell,
		Cwd:         info.Cwd,
		GitBranch:   info.GitBranch,
		PkgManager:  info.PkgManager,
		LastCommand: cap.Command,
		ExitCode:    cap.ExitCode,
		Output:      red.Text,
	}

	model := cfg.Model(prov)
	cacheKey := cache.Key(string(prov), model, language, req.Output, req.LastCommand)
	if cfg.CacheEnabled && !*noCache {
		if e, ok := cache.Get(cacheKey); ok {
			ui.Info(fmt.Sprintf("[cached • %s • %s]", e.Provider, e.Model))
			fmt.Println()
			fmt.Println(render.Markdown(e.Answer))
			return
		}
	}

	client, err := provider.New(cfg, prov)
	if err != nil {
		ui.Err(fmt.Sprintf("провайдер %s недоступен: %v", prov, err))
		ui.Info("настрой ключ: wtf config")
		os.Exit(2)
	}

	start := time.Now()
	sp := ui.NewSpinner(t(language, "Читаю stderr...", "Reading stderr..."))
	sp.Start()
	time.Sleep(180 * time.Millisecond)
	sp.Update(t(language, "Собираю контекст (OS, shell, git)...", "Collecting context (OS, shell, git)..."))
	time.Sleep(180 * time.Millisecond)
	sp.Update(t(language,
		fmt.Sprintf("Чищу секреты (regex × %d правил)...", len(red.Hits)+13),
		"Redacting secrets..."))
	time.Sleep(180 * time.Millisecond)
	sp.Update(t(language,
		fmt.Sprintf("Запрос в %s (%s)...", client.Name(), model),
		fmt.Sprintf("Calling %s (%s)...", client.Name(), model)))

	// Через 600мс после старта запроса — переключаем текст на "получаю ответ"
	// чтобы спиннер не "врал" что мы всё ещё отправляем.
	switchCtx, switchCancel := stdctx.WithCancel(stdctx.Background())
	defer switchCancel()
	go func() {
		select {
		case <-switchCtx.Done():
		case <-time.After(700 * time.Millisecond):
			sp.Update(t(language, "Получаю ответ...", "Receiving response..."))
		}
	}()

	ctx, cancel := stdctx.WithTimeout(stdctx.Background(), 90*time.Second)
	defer cancel()
	answer, err := client.Explain(ctx, req)
	switchCancel()
	if err != nil {
		sp.StopFail(fmt.Sprintf("%s: %v", client.Name(), err))
		os.Exit(1)
	}
	elapsed := time.Since(start)
	sp.StopOK(fmt.Sprintf("Готово · %s · %s", fmtDur(elapsed), client.Name()))

	if cfg.CacheEnabled && !*noCache {
		_ = cache.Put(cache.Entry{
			Key:       cacheKey,
			Provider:  string(prov),
			Model:     model,
			Language:  language,
			Answer:    answer,
			CreatedAt: time.Now(),
		})
	}

	fmt.Println()
	fmt.Println(render.Markdown(answer))
}

// hasPipedStdin — true если stdin это пайп/файл, не интерактивный терминал.
// Используется для поддержки `<команда> 2>&1 | wtf` без явных флагов.
func hasPipedStdin() bool {
	return !term.IsTerminal(int(os.Stdin.Fd()))
}

func anyKeyConfigured(cfg *config.Config) bool {
	for _, p := range []config.Provider{config.ProviderClaude, config.ProviderOpenAI, config.ProviderGemini} {
		if cfg.APIKey(p) != "" {
			return true
		}
	}
	return false
}

func showFirstRunNotice(cfg *config.Config, r redact.Result, info wctx.Info, cap *shellhook.Capture) {
	box := ui.Box{
		Title: "wtf · что отправляется на сервер",
		Lines: []string{
			fmt.Sprintf("OS:           %s", info.OS),
			fmt.Sprintf("Shell:        %s", info.Shell),
			fmt.Sprintf("Cwd:          %s", redact.Apply(info.Cwd).Text),
			fmt.Sprintf("Git branch:   %s", nz(info.GitBranch)),
			fmt.Sprintf("PkgManager:   %s", nz(info.PkgManager)),
			fmt.Sprintf("Команда:      %s", nz(cap.Command)),
			fmt.Sprintf("Exit code:    %d", cap.ExitCode),
			fmt.Sprintf("Вывод:        %d байт (после редакции)", len(r.Text)),
		},
	}
	if summary := redact.Summary(r); summary != "" {
		box.Lines = append(box.Lines, "Удалено:      "+summary)
	}
	box.Lines = append(box.Lines, "")
	box.Lines = append(box.Lines, "Токены, пароли, email, JWT, ключи — отфильтрованы regex'ом.")
	box.Lines = append(box.Lines, "Это уведомление показывается один раз.")
	box.Render(os.Stderr)
	cfg.RedactionShown = true
	_ = cfg.Save()
}

func nz(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func runInit(args []string) {
	var s shellhook.Shell
	if len(args) > 0 {
		s = shellhook.Shell(args[0])
	} else {
		s = shellhook.DetectInstall()
		ui.Info(fmt.Sprintf("обнаружен shell: %s", s))
	}
	rc, err := shellhook.Install(s)
	if err != nil {
		ui.Err(fmt.Sprintf("install: %v", err))
		os.Exit(1)
	}
	ui.OK(fmt.Sprintf("shell-хук установлен в %s", rc))
	ui.Info("перезапусти shell или выполни:")
	switch s {
	case shellhook.ShellBash:
		fmt.Fprintln(os.Stderr, "    source ~/.bashrc")
	case shellhook.ShellZsh:
		fmt.Fprintln(os.Stderr, "    source ~/.zshrc")
	case shellhook.ShellFish:
		fmt.Fprintln(os.Stderr, "    source ~/.config/fish/conf.d/wtf.fish")
	case shellhook.ShellPowerShell:
		fmt.Fprintln(os.Stderr, "    . $PROFILE")
	}
	cfg, _ := config.Load()
	if !anyKeyConfigured(cfg) {
		ui.Warn("API ключ не настроен. Запусти: wtf config")
	}
}

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
	ui.KV("cache", fmt.Sprintf("%v", cfg.CacheEnabled))
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
	case k == "cache":
		cfg.CacheEnabled = v == "true" || v == "1" || v == "on"
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

	// Сначала настраиваем выбранного провайдера. Других не трогаем —
	// после первой настройки спрашиваем, нужно ли добавить ещё.
	configureProvider(r, cfg, cfg.DefaultProvider)

	allProviders := []config.Provider{config.ProviderClaude, config.ProviderOpenAI, config.ProviderGemini}
	for {
		// Какие ещё провайдеры можно добавить — те, у которых ключ не задан.
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

	if anyKeyConfigured(cfg) {
		ui.Info("следующий шаг: wtf init  (установить shell-хук)")
	} else {
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

func t(lang, ru, en string) string {
	if lang == "ru" {
		return ru
	}
	return en
}

func fmtDur(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func mask4(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "…" + s[len(s)-4:]
}
