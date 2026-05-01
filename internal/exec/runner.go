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
	Command      string
	ExitCode     int
	Output       string
	Duration     time.Duration
	TimedOut     bool
	TrimmedBytes int
	ErrorMessage string // если не удалось даже запустить (пермишены, exec not found)
}

// MaxOutputBytes — потолок захвата вывода одной команды.
// 8 КБ это ~2к токенов — нормально для журналов, статусов, listings.
const MaxOutputBytes = 8 * 1024

// DefaultTimeout — максимум на выполнение одной диагностической команды.
const DefaultTimeout = 30 * time.Second

// Run выполняет команду через системный shell и возвращает Result.
//
// ВАЖНО: эта функция должна вызываться только для команд, прошедших
// Classify() == ClassSafe.
//
// На Unix процесс запускается в отдельной process group, чтобы при таймауте
// мы могли убить всё дерево дочерних процессов (некоторые команды вроде
// `pm2 logs` форкают и держат pipe — без kill всей группы Run зависнет).
func Run(ctx context.Context, command string) *Result {
	command = normalizeStreamingCommand(command)
	res := &Result{Command: command}
	start := time.Now()

	runCtx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	shell, args := shellInvocation(command)
	cmd := exec.Command(shell, args...) // не CommandContext — мы сами рулим kill
	setProcessGroup(cmd)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Start(); err != nil {
		res.ExitCode = -1
		res.ErrorMessage = err.Error()
		res.Duration = time.Since(start)
		return res
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		if err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				res.ExitCode = ee.ExitCode()
			} else {
				res.ExitCode = -1
				res.ErrorMessage = err.Error()
			}
		}
	case <-runCtx.Done():
		res.TimedOut = true
		res.ErrorMessage = fmt.Sprintf("команда не завершилась за %s", DefaultTimeout)
		killProcessGroup(cmd)
		// Подождём до 2с пока процесс реально умрёт и pipe закроется,
		// иначе buf может быть в неконсистентном состоянии.
		select {
		case <-done:
		case <-time.After(2 * time.Second):
		}
		res.ExitCode = -1
	}
	res.Duration = time.Since(start)

	out := buf.String()
	if len(out) > MaxOutputBytes {
		res.TrimmedBytes = len(out) - MaxOutputBytes
		// Берём ХВОСТ — для journalctl/dmesg/logs последние строки важнее.
		out = out[len(out)-MaxOutputBytes:]
	}
	res.Output = strings.TrimRight(out, "\n")
	return res
}

// normalizeStreamingCommand переписывает команды которые по умолчанию
// открывают streaming-режим и никогда не завершаются сами (`pm2 logs`,
// `journalctl -f`, `tail -f`, `docker logs -f`). Без этой нормализации модель
// упирается в таймаут и теряет 30 секунд на каждой такой команде.
//
// Подход: если команда явно не запросила лимит/no-stream — добавляем флаги
// которые превращают её в одноразовый snapshot.
func normalizeStreamingCommand(command string) string {
	c := strings.TrimSpace(command)
	lower := strings.ToLower(c)
	tokens := strings.Fields(c)
	if len(tokens) < 2 {
		return command
	}

	// pm2 logs ... — добавляем --nostream и --lines 100 если их нет.
	if tokens[0] == "pm2" && tokens[1] == "logs" {
		if !strings.Contains(c, "--nostream") {
			c += " --nostream"
		}
		if !strings.Contains(c, "--lines") {
			c += " --lines 100"
		}
		return c
	}

	// journalctl -f / --follow — заменяем на -n 200 --no-pager.
	if tokens[0] == "journalctl" {
		hasFollow := false
		for _, t := range tokens[1:] {
			if t == "-f" || t == "--follow" {
				hasFollow = true
				break
			}
		}
		if hasFollow {
			c = strings.ReplaceAll(c, " -f ", " ")
			c = strings.ReplaceAll(c, " --follow ", " ")
			c = strings.TrimSuffix(c, " -f")
			c = strings.TrimSuffix(c, " --follow")
			if !strings.Contains(c, " -n ") && !strings.Contains(c, "--lines") {
				c += " -n 200"
			}
			if !strings.Contains(c, "--no-pager") {
				c += " --no-pager"
			}
		}
		return c
	}

	// tail -f / --follow — заменяем на tail -n 200.
	if tokens[0] == "tail" {
		if strings.Contains(lower, " -f") || strings.Contains(lower, "--follow") {
			c = strings.ReplaceAll(c, " -f ", " ")
			c = strings.ReplaceAll(c, " --follow ", " ")
			c = strings.TrimSuffix(c, " -f")
			c = strings.TrimSuffix(c, " --follow")
			if !strings.Contains(c, " -n ") && !strings.Contains(c, "--lines") {
				// Вставляем после tail
				c = strings.Replace(c, "tail ", "tail -n 200 ", 1)
			}
		}
		return c
	}

	// docker/podman logs -f / --follow — добавляем --tail 200, убираем follow.
	if (tokens[0] == "docker" || tokens[0] == "podman") && len(tokens) >= 2 && tokens[1] == "logs" {
		if strings.Contains(c, " -f") || strings.Contains(c, "--follow") {
			c = strings.ReplaceAll(c, " -f ", " ")
			c = strings.ReplaceAll(c, " --follow ", " ")
			c = strings.TrimSuffix(c, " -f")
			c = strings.TrimSuffix(c, " --follow")
		}
		if !strings.Contains(c, "--tail") {
			c += " --tail 200"
		}
		return c
	}

	return command
}

// shellInvocation возвращает shell и аргументы для запуска команды строкой.
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
	return "/bin/sh", []string{"-c", command}
}
