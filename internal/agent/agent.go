// Package agent — итеративный цикл диалога с моделью + инструменты.
//
// Логика цикла:
//  1. Юзер задал вопрос → собираем system + первый user message.
//  2. Модель отвечает с одним или несколькими tool_use.
//  3. Для каждого tool_use:
//     - run_command: классифицируем; safe → выполняем сами; иначе → показываем
//       юзеру и сообщаем модели через tool_result (отказ).
//     - show_command: рендерим юзеру и в tool_result говорим "показано, ждём".
//     - finish: финал. Печатаем summary, сохраняем notes в память, выходим.
//  4. Если итераций > MaxIterations — форсим finish и выходим с предупреждением.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	wctx "github.com/bitcoff/wtf/internal/context"
	wexec "github.com/bitcoff/wtf/internal/exec"
	"github.com/bitcoff/wtf/internal/memory"
	"github.com/bitcoff/wtf/internal/provider"
)

// MaxIterations — жёсткий лимит раундов AI↔tool. Защищает от
// бесконечных циклов и слива денег.
//
// Раунд = один полный запрос к API провайдера. В одном раунде модель может
// вернуть несколько tool_calls параллельно (см. промпт), поэтому 5 раундов
// = до ~30 команд, чего хватает на любую разумную диагностику.
//
// Раньше было 15 — но это провоцировало модель шлёпать по одной команде
// за раунд и упираться в лимит. С новым промптом и параллелизмом 5 раундов
// — щедрый запас.
const MaxIterations = 5

// KeepFullRounds — сколько последних раундов держать в истории с ПОЛНЫМИ
// tool_result. Старые раунды сжимаются до короткой строки-маркера, чтобы
// не раздувать контекст и не упираться в TPM-лимиты провайдеров.
//
// 4 раунда обычно достаточно: модель помнит самое свежее, а старое всё равно
// уже отражено в её собственных reasoning-сообщениях.
const KeepFullRounds = 4

// MaxToolResultBytes — потолок одного tool_result после которого его обрежем
// при trim'е. До этого порога — оставляем как есть.
const MaxToolResultBytes = 2048

// MaxSameCommandRepeats — сколько раз модели можно вызвать одну и ту же команду.
// 1 = никогда не повторяем (после первого вызова — отказ). Это убирает класс
// багов "модель забыла что уже это смотрела и крутится по кругу".
const MaxSameCommandRepeats = 1

// MaxCommandsPerRound — потолок параллельных run_command в одном раунде.
// Защищает от взрыва токенов: 8 команд × 8 КБ вывода = 64 КБ tool_results
// в следующий запрос → лёгкий 429 на провайдерах с TPM<60k.
//
// 5 — золотая середина: достаточно для диагностики типичного кейса
// (например status + journalctl + nginx -t + ls конфигов + права на cert)
// и не разносит контекст. Лишние команды получают отказ с просьбой
// перевызвать их в следующем раунде если ещё актуально.
const MaxCommandsPerRound = 5

// InterRoundDelay — пауза между AI-раундами. Не даёт спамить провайдера на
// быстрых ответах: если модель ответила за 200мс и мы немедленно отправили
// следующий запрос, легко упереться в TPM. 800мс — компромисс между скоростью
// и шансом схватить 429.
const InterRoundDelay = 800 * time.Millisecond

// IO — абстракция вывода для агента. Главный поток печатает в терминал;
// тесты могут подсунуть буфер. Все методы принимают уже отформатированный текст.
type IO interface {
	StepCommand(reason, command string)        // → запускаем X (потому что Y)
	CommandOutput(command, output string, exit int, dur time.Duration, timedOut bool)
	UserCommand(reason, command string)        // ⚠ выполни сам: ...
	Refused(command, reason string)            // отказали в запуске
	Final(summary string)                      // финальный ответ
	Thinking(label string)                     // спиннер во время AI-запроса
	StopThinking(success bool, label string)   // конец спиннера
	RateLimitWait(wait time.Duration, attempt int) // провайдер просит подождать
}

// Result — что сессия вернула наружу.
type Result struct {
	Summary    string
	Notes      []memory.Entry
	Iterations int
	Stopped    string // "finish" | "max_iterations" | "error"
}

// Run выполняет одну сессию агента до finish или лимита итераций.
// Любая ошибка от провайдера прерывает сессию и возвращается наружу.
func Run(
	ctx context.Context,
	cli provider.Client,
	envInfo wctx.Info,
	memStore *memory.Store,
	io IO,
	userQuestion string,
	piped string, // если был pipe — содержимое stdin (опционально)
	lang string,
) (*Result, error) {
	system := SystemPrompt(lang, envInfo, memStore.SystemContext(2000))

	firstMsg := userQuestion
	if firstMsg == "" {
		firstMsg = "Опиши проблему"
	}
	if piped != "" {
		firstMsg = fmt.Sprintf("%s\n\nВот вывод который я хочу разобрать:\n```\n%s\n```", firstMsg, piped)
	}

	messages := []provider.Message{
		{Role: provider.RoleUser, Text: firstMsg},
	}

	res := &Result{}
	tools := Tools()

	// runHistory — сколько раз каждая команда уже была запущена в этой сессии.
	// Ключ — нормализованная строка команды (после TrimSpace + collapse spaces).
	// Используется в handleToolCall чтобы отказывать в повторах.
	runHistory := map[string]int{}

	for i := 0; i < MaxIterations; i++ {
		res.Iterations = i + 1

		// Throttle: пауза между раундами, чтобы не упереться в TPM провайдера.
		// Первый раунд — без паузы, юзер ждёт.
		if i > 0 {
			select {
			case <-ctx.Done():
				res.Stopped = "error"
				return res, ctx.Err()
			case <-time.After(InterRoundDelay):
			}
		}

		io.Thinking(fmt.Sprintf("Думаю (%s)...", cli.Name()))
		resp, err := cli.Chat(ctx, provider.ChatRequest{
			System:      system,
			Messages:    trimHistory(messages, KeepFullRounds),
			Tools:       tools,
			MaxTokens:   2048,
			OnRateLimit: io.RateLimitWait,
		})
		io.StopThinking(err == nil, "")
		if err != nil {
			res.Stopped = "error"
			return res, err
		}

		// Если модель вернула только текст без tool_use — это значит она
		// "сорвалась" из протокола (бывает у openai/gemini). Считаем это финалом.
		if len(resp.ToolCalls) == 0 {
			res.Summary = strings.TrimSpace(resp.Text)
			if res.Summary == "" {
				res.Summary = "(модель не вернула ответ)"
			}
			io.Final(res.Summary)
			res.Stopped = "finish"
			return res, nil
		}

		// Сохраняем ассистент-сообщение в историю КАК ЕСТЬ — со всеми tool_calls.
		// Дальше нам надо ответить tool_result для каждого вызова.
		messages = append(messages, provider.Message{
			Role:      provider.RoleAssistant,
			Text:      resp.Text,
			ToolCalls: resp.ToolCalls,
		})

		// Обрабатываем все tool_calls и готовим tool_result сообщения.
		// Ловушка: если среди вызовов есть finish — после него остальные
		// игнорируем, но всё равно ответим на каждый чтобы протокол не сломался.
		var finishCall *provider.ToolCall
		for i := range resp.ToolCalls {
			if resp.ToolCalls[i].Name == ToolFinish {
				finishCall = &resp.ToolCalls[i]
				break
			}
		}

		// Счётчик выполненных в этом раунде run_command — для лимита параллельности.
		// Превышение → отказ (нельзя пропустить вообще: tool-use API требует
		// ответа на каждый tool_call, иначе следующий запрос не примут).
		runCommandsThisRound := 0

		for _, tc := range resp.ToolCalls {
			var result string
			if tc.Name == ToolRunCommand && runCommandsThisRound >= MaxCommandsPerRound {
				cmdStr, _ := tc.Input["command"].(string)
				io.Refused(cmdStr, fmt.Sprintf(
					"лимит %d команд/раунд — модель просит слишком многого",
					MaxCommandsPerRound))
				result = fmt.Sprintf(
					"ОТКАЗАНО: лимит %d команд за один раунд исчерпан. "+
						"Команда %q не выполнена. Если она ещё нужна — вызови её "+
						"в следующем раунде, но сначала проанализируй уже полученные "+
						"результаты. Не запрашивай больше команд чем нужно.",
					MaxCommandsPerRound, cmdStr)
			} else {
				result = handleToolCall(ctx, tc, io, runHistory)
				if tc.Name == ToolRunCommand {
					runCommandsThisRound++
				}
			}
			messages = append(messages, provider.Message{
				Role:       provider.RoleTool,
				ToolCallID: toolCallIDFor(cli, tc),
				Text:       result,
			})
		}

		if finishCall != nil {
			parseFinish(finishCall, res)
			io.Final(res.Summary)
			res.Stopped = "finish"
			return res, nil
		}
	}

	// Достигли лимита — форсим финальный запрос с просьбой завершить.
	messages = append(messages, provider.Message{
		Role: provider.RoleUser,
		Text: "Достигнут лимит раундов диагностики. Заверши через finish с тем что есть.",
	})
	io.Thinking("финальный ответ...")
	resp, err := cli.Chat(ctx, provider.ChatRequest{
		System:      system,
		Messages:    trimHistory(messages, KeepFullRounds),
		Tools:       tools,
		MaxTokens:   1024,
		OnRateLimit: io.RateLimitWait,
	})
	io.StopThinking(err == nil, "")
	if err != nil {
		res.Stopped = "error"
		return res, err
	}
	for _, tc := range resp.ToolCalls {
		if tc.Name == ToolFinish {
			parseFinish(&tc, res)
			io.Final(res.Summary)
			res.Stopped = "max_iterations"
			return res, nil
		}
	}
	if resp.Text != "" {
		res.Summary = resp.Text
	} else {
		res.Summary = "Превышен лимит итераций без финального ответа."
	}
	io.Final(res.Summary)
	res.Stopped = "max_iterations"
	return res, nil
}

// handleToolCall выполняет один tool_call и возвращает строку для tool_result.
//
// runHistory — счётчик уже запущенных команд в этой сессии (нормализованная строка → count).
// Используем для отказа от повторного запуска одной и той же команды.
func handleToolCall(ctx context.Context, tc provider.ToolCall, io IO, runHistory map[string]int) string {
	switch tc.Name {
	case ToolRunCommand:
		command, _ := tc.Input["command"].(string)
		reason, _ := tc.Input["reason"].(string)
		command = strings.TrimSpace(command)
		if command == "" {
			return "ошибка: пустая команда"
		}

		// Защита от зацикливания: модель иногда забывает что уже это смотрела.
		// После N запусков той же команды — отказываем и просим использовать
		// предыдущий вывод.
		key := normalizeCommand(command)
		if runHistory[key] >= MaxSameCommandRepeats {
			io.Refused(command, "уже выполнялась — модель должна использовать предыдущий вывод")
			return fmt.Sprintf("ОТКАЗАНО: команда %q уже выполнялась в этой сессии. "+
				"Посмотри её вывод в предыдущих раундах. Если нужны другие данные — "+
				"запусти ДРУГУЮ команду или вызови finish.", command)
		}

		class := wexec.Classify(command)
		if class != wexec.ClassSafe {
			io.Refused(command, fmt.Sprintf("команда классифицирована как %s — выполни через show_command", class))
			return fmt.Sprintf("ОТКАЗАНО: команда %q не безопасна для авто-запуска (класс: %s). "+
				"Используй show_command вместо run_command для destructive-операций.",
				command, class)
		}

		runHistory[key]++

		io.StepCommand(reason, command)
		result := wexec.Run(ctx, command)
		io.CommandOutput(result.Command, result.Output, result.ExitCode, result.Duration, result.TimedOut)

		return formatRunResult(result)

	case ToolShowCommand:
		command, _ := tc.Input["command"].(string)
		reason, _ := tc.Input["reason"].(string)
		io.UserCommand(reason, command)
		return "Команда показана пользователю. Жди что он либо выполнит и пришлёт вывод, " +
			"либо завершит сессию. Не вызывай эту команду повторно через run_command — она destructive."

	case ToolFinish:
		// Обрабатывается в главном цикле, сюда не должно прилетать.
		// Но если прилетело (повторный вызов в одном раунде) — даём пустой результат.
		return "ok"

	default:
		return fmt.Sprintf("неизвестный инструмент: %s", tc.Name)
	}
}

func formatRunResult(r *wexec.Result) string {
	var b strings.Builder
	if r.TimedOut {
		fmt.Fprintf(&b, "[ТАЙМАУТ %s]\n", r.Duration)
	}
	if r.ErrorMessage != "" && !r.TimedOut {
		fmt.Fprintf(&b, "[ошибка запуска: %s]\n", r.ErrorMessage)
	}
	fmt.Fprintf(&b, "exit=%d duration=%s\n", r.ExitCode, r.Duration.Round(time.Millisecond))
	if r.TrimmedBytes > 0 {
		fmt.Fprintf(&b, "[вывод обрезан: показаны последние %d из %d байт]\n",
			len(r.Output), len(r.Output)+r.TrimmedBytes)
	}
	b.WriteString("---\n")
	b.WriteString(r.Output)
	return b.String()
}

func parseFinish(tc *provider.ToolCall, res *Result) {
	if s, ok := tc.Input["summary"].(string); ok {
		res.Summary = strings.TrimSpace(s)
	}
	if rawNotes, ok := tc.Input["notes"].([]any); ok {
		for _, raw := range rawNotes {
			m, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			entry := memory.Entry{}
			if t, ok := m["type"].(string); ok {
				entry.Type = memory.Type(t)
			}
			if k, ok := m["key"].(string); ok {
				entry.Key = k
			}
			if c, ok := m["content"].(string); ok {
				entry.Content = c
			}
			if ttl, ok := m["ttl_days"].(float64); ok {
				entry.TTLDays = int(ttl)
			}
			if entry.Type != "" && entry.Key != "" && entry.Content != "" {
				res.Notes = append(res.Notes, entry)
			}
		}
	}
}

// trimHistory возвращает копию messages где старые tool_result сжаты до
// коротких заглушек. Сохраняем всё что относится к последним keepRounds раундам
// "assistant с tool_calls + соответствующие tool результаты".
//
// Алгоритм:
//  1. Идём с конца, считаем "раунды" (раунд = assistant-сообщение с tool_calls).
//  2. Пока счётчик ≤ keepRounds — копируем как есть.
//  3. Дальше — для tool-сообщений заменяем Text на маркер, для assistant
//     с tool_calls обнуляем длинные текстовые поля.
//
// Это снижает context size в разы при долгих сессиях без потери "линии мысли"
// модели — она помнит ЧТО запускала и видит свежие результаты.
func trimHistory(messages []provider.Message, keepRounds int) []provider.Message {
	if len(messages) == 0 || keepRounds <= 0 {
		return messages
	}

	// Подсчёт раундов с конца. Раунд = появление assistant-сообщения.
	roundCount := 0
	cutoff := 0
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == provider.RoleAssistant && len(messages[i].ToolCalls) > 0 {
			roundCount++
			if roundCount > keepRounds {
				cutoff = i + 1 // всё что < cutoff — старое, пора сжимать
				break
			}
		}
	}
	if cutoff == 0 {
		// Раундов меньше keepRounds — ничего не трогаем.
		return messages
	}

	out := make([]provider.Message, len(messages))
	for i, m := range messages {
		if i >= cutoff {
			out[i] = m
			continue
		}
		switch m.Role {
		case provider.RoleTool:
			// Сжимаем результат старой команды до короткой пометки.
			out[i] = provider.Message{
				Role:       m.Role,
				ToolCallID: m.ToolCallID,
				Text:       summarizeToolResult(m.Text),
			}
		case provider.RoleAssistant:
			// Текст оставляем (он короткий — "мысли" модели), tool_calls тоже —
			// они нужны для протокола (пара call→result).
			out[i] = m
		default:
			// User-сообщения и прочее — копируем как есть.
			out[i] = m
		}
	}
	return out
}

// normalizeCommand схлопывает повторные пробелы и обрезает края, чтобы
// "ls   -la" и "ls -la" считались одной командой при детекте дублей.
func normalizeCommand(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// summarizeToolResult ужимает длинный stdout до 1-2 строк маркера.
// Сохраняем exit-код и факт что вывод был обрезан.
func summarizeToolResult(s string) string {
	if len(s) <= 200 {
		return s
	}
	// Пытаемся вытащить первую строку (там обычно "exit=N duration=...")
	lineEnd := strings.IndexByte(s, '\n')
	header := s
	if lineEnd > 0 && lineEnd < 200 {
		header = s[:lineEnd]
	} else if len(header) > 200 {
		header = header[:200]
	}
	return header + "\n[...вывод обрезан при сжатии истории, см. предыдущие раунды если нужны детали...]"
}

// toolCallIDFor возвращает идентификатор для tool_result.
//
// Claude/OpenAI используют именно ID. Gemini не имеет ID, но требует чтобы
// в functionResponse было ИМЯ функции — поэтому для gemini кладём имя.
func toolCallIDFor(cli provider.Client, tc provider.ToolCall) string {
	if cli.Name() == "gemini" {
		return tc.Name
	}
	return tc.ID
}

// Consolidate сжимает память через AI. Вызывается при превышении лимита entries
// или раз в N сессий. Не критична — при ошибке просто оставляем память как есть.
func Consolidate(ctx context.Context, cli provider.Client, store *memory.Store) error {
	if len(store.Entries) == 0 {
		return nil
	}

	prompt := fmt.Sprintf("%s\n\nТекущие записи:\n```json\n%s\n```",
		ConsolidationPrompt, store.AsJSON())

	resp, err := cli.Chat(ctx, provider.ChatRequest{
		System:    "Ты редактируешь память ассистента. Возвращай результат через инструмент save_consolidated_memory.",
		Messages:  []provider.Message{{Role: provider.RoleUser, Text: prompt}},
		Tools:     ConsolidationTools(),
		MaxTokens: 4096,
	})
	if err != nil {
		return err
	}

	for _, tc := range resp.ToolCalls {
		if tc.Name != ToolSaveConsolidated {
			continue
		}
		raw, _ := tc.Input["entries"].([]any)
		var entries []memory.Entry
		for _, item := range raw {
			data, err := json.Marshal(item)
			if err != nil {
				continue
			}
			var e memory.Entry
			if err := json.Unmarshal(data, &e); err == nil && e.Key != "" {
				entries = append(entries, e)
			}
		}
		if len(entries) > 0 {
			store.Replace(entries)
		}
		return nil
	}
	return fmt.Errorf("консолидатор не вернул save_consolidated_memory")
}
