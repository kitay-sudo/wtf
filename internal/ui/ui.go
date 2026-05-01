package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/term"
)

const (
	reset     = "\033[0m"
	bold      = "\033[1m"
	dim       = "\033[2m"
	hidden    = "\033[?25l"
	visible   = "\033[?25h"
	clearLine = "\r\033[2K"

	yellow     = "\033[33m"
	yellowBold = "\033[1;33m"
	cyan       = "\033[36m"
	red        = "\033[31m"
	gray       = "\033[90m"
	white      = "\033[97m"
)

func IsTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func IsStderrTTY() bool {
	return term.IsTerminal(int(os.Stderr.Fd()))
}

func colorize(c, s string) string {
	if !IsStderrTTY() || os.Getenv("NO_COLOR") != "" {
		return s
	}
	return c + s + reset
}

// Banner — first-run greeting in style of Claude CLI / goronin installer.
func Banner(title, subtitle string) {
	out := os.Stderr
	fmt.Fprintln(out)
	fmt.Fprintln(out, colorize(yellowBold, "  [!?]  "+title))
	if subtitle != "" {
		fmt.Fprintln(out, colorize(gray, "      "+subtitle))
	}
	fmt.Fprintln(out)
}

// Section header
func Section(title string) {
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, colorize(bold, "  "+title))
	fmt.Fprintln(os.Stderr, colorize(gray, "  "+strings.Repeat("─", len(title))))
}

func Step(s string)    { fmt.Fprintf(os.Stderr, "  %s  %s\n", colorize(cyan, "→"), s) }
func OK(s string)      { fmt.Fprintf(os.Stderr, "  %s  %s\n", colorize(yellowBold, "✓"), s) }
func Info(s string)    { fmt.Fprintf(os.Stderr, "  %s  %s\n", colorize(gray, "ⓘ"), colorize(gray, s)) }
func Warn(s string)    { fmt.Fprintf(os.Stderr, "  %s  %s\n", colorize(yellow, "⚠"), s) }
func Err(s string)     { fmt.Fprintf(os.Stderr, "  %s  %s\n", colorize(red, "✗"), s) }
func KV(k, v string)   { fmt.Fprintf(os.Stderr, "    %s  %s\n", colorize(gray, padRight(k+":", 14)), v) }
func Plain(s string)   { fmt.Fprintln(os.Stderr, s) }
func PlainOut(s string) { fmt.Fprintln(os.Stdout, s) }

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

// Spinner — Braille-style spinner на stderr. Безопасен в non-TTY (no-op).
type Spinner struct {
	prefix string
	stop   chan struct{}
	done   chan struct{}
	live   atomic.Bool
}

var spinFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func NewSpinner(prefix string) *Spinner {
	return &Spinner{
		prefix: prefix,
		stop:   make(chan struct{}),
		done:   make(chan struct{}),
	}
}

func (s *Spinner) Start() {
	// non-TTY (cron/pipe): без анимации, просто строка с временным префиксом.
	if !IsStderrTTY() || os.Getenv("NO_COLOR") != "" {
		fmt.Fprintf(os.Stderr, "%s%s\n", linePrefix("→", cyan), s.prefix)
		close(s.done)
		return
	}
	s.live.Store(true)
	fmt.Fprint(os.Stderr, hidden)
	go func() {
		defer close(s.done)
		i := 0
		t := time.NewTicker(80 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-s.stop:
				fmt.Fprint(os.Stderr, clearLine, visible)
				return
			case <-t.C:
				// Тот же префикс [HH:MM:SS] что и в финальных строках, но
				// иконка — это анимированный кадр спиннера. Получается
				// идеальное вертикальное выравнивание со строкой результата.
				frame := spinFrames[i%len(spinFrames)]
				fmt.Fprintf(os.Stderr, "%s%s%s",
					clearLine, linePrefix(frame, yellow), s.prefix)
				i++
			}
		}
	}()
}

func (s *Spinner) Update(prefix string) {
	s.prefix = prefix
}

// StopOK останавливает спиннер. Если msg пустая — НИЧЕГО не печатает
// (ни галочки, ни строки), только убирает текущую анимацию. Это нужно для
// quiet-режима: после спиннера каллер сам печатает финальную строку через
// CommandLineQuiet, и галочка от спиннера была бы лишним мусором.
func (s *Spinner) StopOK(msg string) {
	live := s.live.Load()
	if live {
		close(s.stop)
		<-s.done
	}
	if msg == "" {
		return
	}
	OK(msg)
}

// StopFail — то же самое для случая ошибки. Пустой msg = тихий стоп.
func (s *Spinner) StopFail(msg string) {
	live := s.live.Load()
	if live {
		close(s.stop)
		<-s.done
	}
	if msg == "" {
		return
	}
	Err(msg)
}

// Box — рамка с заголовком, для consent-баннера.
type Box struct {
	Title string
	Lines []string
}

func (b Box) Render(w io.Writer) {
	width := boxWidth(b)
	top := "╭─ " + b.Title + " " + strings.Repeat("─", max(0, width-len(b.Title)-3)) + "╮"
	bot := "╰" + strings.Repeat("─", width) + "╯"
	fmt.Fprintln(w, colorize(yellow, top))
	for _, line := range b.Lines {
		fmt.Fprintf(w, "%s %s\n", colorize(yellow, "│"), line)
	}
	fmt.Fprintln(w, colorize(yellow, bot))
}

func boxWidth(b Box) int {
	w := len(b.Title) + 4
	for _, l := range b.Lines {
		if visualLen(l)+2 > w {
			w = visualLen(l) + 2
		}
	}
	if w < 50 {
		w = 50
	}
	return w
}

// visualLen — длина строки без ANSI и с учётом ширины unicode (грубая).
func visualLen(s string) int {
	var n int
	inEsc := false
	for _, r := range s {
		if r == 0x1b {
			inEsc = true
			continue
		}
		if inEsc {
			if r == 'm' {
				inEsc = false
			}
			continue
		}
		n++
	}
	return n
}

// Prompt — простой ввод с дефолтом.
func Prompt(reader *bufio.Reader, label, def string) string {
	if def != "" {
		fmt.Fprintf(os.Stderr, "%s %s [%s]: ", colorize(yellow, "?"), label, colorize(gray, def))
	} else {
		fmt.Fprintf(os.Stderr, "%s %s: ", colorize(yellow, "?"), label)
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		return def
	}
	v := strings.TrimSpace(line)
	if v == "" {
		return def
	}
	return v
}

// Choice — выбор одного из вариантов.
// ChoiceOrCustom — как Choice, но дополнительно позволяет ввести произвольное
// имя варианта (для случаев "хочу новую модель которой ещё нет в списке").
// Если ввод не совпал ни с цифрой, ни с именем из options — возвращаем
// введённое значение как есть.
func ChoiceOrCustom(reader *bufio.Reader, label string, options []string, def string) string {
	fmt.Fprintf(os.Stderr, "%s %s\n", colorize(yellow, "?"), label)
	for i, o := range options {
		suffix := ""
		if o == def {
			suffix = colorize(gray, "  (по умолчанию)")
		}
		fmt.Fprintf(os.Stderr, "    %d) %s%s\n", i+1, o, suffix)
	}
	hint := fmt.Sprintf("номер 1-%d, своё имя или Enter:", len(options))
	fmt.Fprintf(os.Stderr, "  %s ", colorize(gray, hint))
	line, _ := reader.ReadString('\n')
	v := strings.TrimSpace(line)
	if v == "" {
		return def
	}
	for i, o := range options {
		if v == fmt.Sprintf("%d", i+1) {
			return o
		}
		if v == o {
			return o
		}
	}
	return v // custom value
}

// Choice — нумерованный выбор. Юзер вводит число (или имя варианта)
// или жмёт Enter чтобы оставить значение по умолчанию. Дефолт помечается
// "(по умолчанию)" в строке варианта — никаких стрелочек/курсоров,
// потому что мы не реализуем raw-mode перехват клавиш и стрелки нажимать
// бесполезно.
func Choice(reader *bufio.Reader, label string, options []string, def string) string {
	fmt.Fprintf(os.Stderr, "%s %s\n", colorize(yellow, "?"), label)
	defaultIdx := -1
	for i, o := range options {
		suffix := ""
		if o == def {
			defaultIdx = i + 1
			suffix = colorize(gray, "  (по умолчанию)")
		}
		fmt.Fprintf(os.Stderr, "    %d) %s%s\n", i+1, o, suffix)
	}
	hint := "номер или Enter"
	if defaultIdx > 0 {
		hint = fmt.Sprintf("номер 1-%d или Enter", len(options))
	}
	fmt.Fprintf(os.Stderr, "  %s ", colorize(gray, hint+":"))
	line, _ := reader.ReadString('\n')
	v := strings.TrimSpace(line)
	if v == "" {
		return def
	}
	for i, o := range options {
		if v == o || v == fmt.Sprintf("%d", i+1) {
			return o
		}
	}
	return def
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// linePrefix возвращает единый префикс для всех событий wtf:
//
//	"[HH:MM:SS] <icon>  "
//
// Все строки агента (команды, финал, предупреждения) должны начинаться с
// этого префикса, чтобы вертикали выравнивались.
//
// Длина префикса фиксирована: 11 символов времени + 1 пробел + 1 знак icon
// + 2 пробела = 15. Цвет применяется отдельно (в видимый размер не входит).
func linePrefix(icon string, iconColor string) string {
	ts := time.Now().Format("15:04:05")
	return fmt.Sprintf("%s %s ",
		colorize(gray, "["+ts+"]"),
		colorize(iconColor, icon))
}

// CommandHeader — verbose-режим: показываем reason + команду перед запуском.
// В тихом режиме это место занимает спиннер CommandRunningStart.
func CommandHeader(reason, command string) {
	out := os.Stderr
	fmt.Fprintln(out)
	fmt.Fprintf(out, "%s%s\n", linePrefix("→", cyan), colorize(gray, reason))
	fmt.Fprintf(out, "%s%s\n", linePrefix("$", yellowBold), colorize(white, command))
}

// CommandLineQuiet печатает финальную одну строку об уже выполненной команде
// в тихом режиме. Формат:
//
//	[HH:MM:SS] ✓ reason · команда · 142ms · 3.2KB
//	[HH:MM:SS] ✗ reason · команда · 142ms · exit=1 · 547B  (если упало)
//	[HH:MM:SS] ⏱ reason · команда · таймаут                (если timeout)
//
// Использовать ПОСЛЕ остановки спиннера выполнения.
func CommandLineQuiet(reason, command string, output string, exit int, dur time.Duration, timedOut bool) {
	out := os.Stderr
	icon, color := "✓", yellowBold
	if exit != 0 {
		icon, color = "✗", red
	}
	if timedOut {
		icon, color = "⏱", red
	}
	parts := []string{}
	if reason != "" {
		parts = append(parts, colorize(gray, reason))
	}
	parts = append(parts, colorize(white, command))
	parts = append(parts, colorize(gray, fmtDur(dur)))
	if exit != 0 && !timedOut {
		parts = append(parts, colorize(yellow, fmt.Sprintf("exit=%d", exit)))
	}
	if timedOut {
		parts = append(parts, colorize(red, "таймаут"))
	}
	if output != "" {
		parts = append(parts, colorize(gray, fmtBytes(len(output))))
	} else if !timedOut {
		parts = append(parts, colorize(gray, "(пусто)"))
	}
	fmt.Fprintf(out, "%s%s\n", linePrefix(icon, color), strings.Join(parts, colorize(gray, " · ")))

	// При ошибке показываем последние 5 строк вывода. Префикс пустой
	// (только пробелы под отступ времени+иконки), чтобы visualy "вложить"
	// строки внутрь предыдущей.
	if (exit != 0 || timedOut) && output != "" {
		lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
		const tailLines = 5
		start := 0
		// Отступ: 11 (время) + 1 (пробел) + 1 (icon) + 1 (пробел) = 14 символов.
		indent := strings.Repeat(" ", 14)
		if len(lines) > tailLines {
			start = len(lines) - tailLines
			fmt.Fprintf(out, "%s%s\n", indent, colorize(gray, fmt.Sprintf("(показаны последние %d из %d строк)", tailLines, len(lines))))
		}
		for _, l := range lines[start:] {
			fmt.Fprintf(out, "%s%s %s\n", indent, colorize(gray, "│"), colorize(gray, l))
		}
	}
}

func fmtBytes(n int) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%dB", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.1fKB", float64(n)/1024)
	default:
		return fmt.Sprintf("%.1fMB", float64(n)/(1024*1024))
	}
}

// CommandResult — verbose-режим: полный вывод с заголовком/футером.
func CommandResult(output string, exit int, dur time.Duration, timedOut bool) {
	out := os.Stderr
	status := colorize(gray, fmt.Sprintf("exit=%d · %s", exit, fmtDur(dur)))
	if exit != 0 {
		status = colorize(yellow, fmt.Sprintf("exit=%d · %s", exit, fmtDur(dur)))
	}
	if timedOut {
		status = colorize(red, fmt.Sprintf("таймаут · %s", fmtDur(dur)))
	}
	if output == "" {
		fmt.Fprintf(out, "  %s %s\n", colorize(gray, "└"), status+colorize(gray, " · (пусто)"))
		return
	}
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	maxLines := 20
	for i, l := range lines {
		if i >= maxLines {
			fmt.Fprintf(out, "  %s %s\n", colorize(gray, "│"),
				colorize(gray, fmt.Sprintf("...ещё %d строк...", len(lines)-maxLines)))
			break
		}
		fmt.Fprintf(out, "  %s %s\n", colorize(gray, "│"), l)
	}
	fmt.Fprintf(out, "  %s %s\n", colorize(gray, "└"), status)
}

// UserCommandBlock — destructive-команда которую юзер должен выполнить сам.
// Используем тот же linePrefix что и для остальных событий (с временем),
// плюс под-строки печатаем с отступом-под-префикс.
func UserCommandBlock(reason, command string) {
	out := os.Stderr
	indent := strings.Repeat(" ", 14)
	fmt.Fprintln(out)
	fmt.Fprintf(out, "%s%s\n", linePrefix("⚠", yellow),
		colorize(yellow, "выполни сам (требует sudo / меняет систему):"))
	if reason != "" {
		fmt.Fprintf(out, "%s%s\n", indent, colorize(gray, reason))
	}
	fmt.Fprintf(out, "%s%s %s\n", indent, colorize(yellowBold, "$"), colorize(white, command))
}

// RefusedBlock — мы отказали в авто-запуске.
func RefusedBlock(command, reason string) {
	out := os.Stderr
	fmt.Fprintf(out, "%s%s: %s\n", linePrefix("✗", red), reason, colorize(gray, command))
}

// FinalBlock — финальный ответ агента. Заголовок с тем же префиксом времени,
// чтобы вертикали с командами выровнялись. Текст ответа печатаем в stdout
// (можно запайпить в файл/jq) с тем же 14-символьным отступом.
// Trailing \n гарантируем — модели иногда забывают, и текст слипается с PS1.
func FinalBlock(text string) {
	indent := strings.Repeat(" ", 14)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "%s%s\n", linePrefix("★", yellowBold), colorize(yellowBold, "ответ:"))
	fmt.Fprintln(os.Stderr)

	text = strings.TrimRight(text, " \t\r\n")
	for _, line := range strings.Split(text, "\n") {
		fmt.Fprintln(os.Stdout, indent+line)
	}
	fmt.Fprintln(os.Stderr)
}

func fmtDur(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

var spinMu sync.Mutex

func WithSpinner(label string, fn func() error) error {
	spinMu.Lock()
	sp := NewSpinner(label)
	sp.Start()
	spinMu.Unlock()
	err := fn()
	if err != nil {
		sp.StopFail(label + " — " + err.Error())
		return err
	}
	sp.StopOK(label)
	return nil
}
