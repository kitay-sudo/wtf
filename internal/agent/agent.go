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
// бесконечных циклов и слива денег. На практике диагностика укладывается в 3-5
// раундов; 15 — щедрый запас на сложные кейсы.
const MaxIterations = 15

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

	for i := 0; i < MaxIterations; i++ {
		res.Iterations = i + 1

		io.Thinking(fmt.Sprintf("думаю (%s)...", cli.Name()))
		resp, err := cli.Chat(ctx, provider.ChatRequest{
			System:    system,
			Messages:  messages,
			Tools:     tools,
			MaxTokens: 2048,
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

		for _, tc := range resp.ToolCalls {
			result := handleToolCall(ctx, tc, io)
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
		System:    system,
		Messages:  messages,
		Tools:     tools,
		MaxTokens: 1024,
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
func handleToolCall(ctx context.Context, tc provider.ToolCall, io IO) string {
	switch tc.Name {
	case ToolRunCommand:
		command, _ := tc.Input["command"].(string)
		reason, _ := tc.Input["reason"].(string)
		command = strings.TrimSpace(command)
		if command == "" {
			return "ошибка: пустая команда"
		}

		class := wexec.Classify(command)
		if class != wexec.ClassSafe {
			io.Refused(command, fmt.Sprintf("команда классифицирована как %s — выполни через show_command", class))
			return fmt.Sprintf("ОТКАЗАНО: команда %q не безопасна для авто-запуска (класс: %s). "+
				"Используй show_command вместо run_command для destructive-операций.",
				command, class)
		}

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
