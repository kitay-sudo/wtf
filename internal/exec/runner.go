package exec

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Result — что мы возвращаем агенту после выполнения команды.
//
// Output обрезан до MaxOutputBytes (по умолчанию 8 КБ), чтобы не сливать
// мегабайты в контекст модели. Если усечение было — TrimmedBytes > 0.
type Result struct {
	Command       string
	ExitCode      int
	Output        string
	Duration      time.Duration
	TimedOut      bool
	TrimmedBytes  int
	ErrorMessage  string // если не удалось даже запустить (пермишены, exec not found)
}

// MaxOutputBytes — потолок захвата вывода одной команды.
// 8 КБ это ~2к токенов — нормально для журналов, статусов, listings.
// Большие логи всё равно бесполезны для AI как сырьё, лучше grep/tail.
const MaxOutputBytes = 8 * 1024

// DefaultTimeout — максимум на выполнение одной диагностической команды.
// 30 секунд хватает на самые медленные (apt list, journalctl с фильтрами).
// Если упёрлись — лучше прервать и сказать модели "tаймаут", чем висеть.
const DefaultTimeout = 30 * time.Second

// Run выполняет команду через системный shell и возвращает Result.
//
// ВАЖНО: эта функция должна вызываться только для команд, прошедших
// Classify() == ClassSafe. Никаких внутренних проверок здесь нет —
// ответственность за классификацию на вызывающем коде.
func Run(ctx context.Context, command string) *Result {
	res := &Result{Command: command}
	start := time.Now()

	runCtx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	shell, args := shellInvocation(command)
	cmd := exec.CommandContext(runCtx, shell, args...)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			res.ExitCode = ee.ExitCode()
		} else {
			res.ExitCode = -1
			res.ErrorMessage = err.Error()
		}
	}
	res.Duration = time.Since(start)

	if runCtx.Err() == context.DeadlineExceeded {
		res.TimedOut = true
		res.ErrorMessage = fmt.Sprintf("команда не завершилась за %s", DefaultTimeout)
	}

	out := buf.String()
	if len(out) > MaxOutputBytes {
		res.TrimmedBytes = len(out) - MaxOutputBytes
		// Берём ХВОСТ, потому что для команд типа journalctl/dmesg
		// последние строки важнее первых (свежие события).
		out = out[len(out)-MaxOutputBytes:]
	}
	res.Output = strings.TrimRight(out, "\n")
	return res
}

// shellInvocation возвращает shell и аргументы для запуска команды строкой.
// На Unix это `sh -c "..."`, на Windows — powershell или cmd.
func shellInvocation(command string) (string, []string) {
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("pwsh"); err == nil {
			return "pwsh", []string{"-NoProfile", "-NonInteractive", "-Command", command}
		}
		if _, err := exec.LookPath("powershell"); err == nil {
			return "powershell", []string{"-NoProfile", "-NonInteractive", "-Command", command}
		}
		return "cmd", []string{"/c", command}
	}
	// На Unix используем sh — он есть везде, в отличие от bash.
	return "/bin/sh", []string{"-c", command}
}
