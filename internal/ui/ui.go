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

	"github.com/charmbracelet/lipgloss"
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

// indentWidth — длина общего префикса всех строк wtf:
// "[HH:MM:SS] <icon>  " = 11 + 1 + 1 + 1 = 14 символов.
const indentWidth = 14

// minBodyWidth — минимальная ширина текстовой колонки даже на узких терминалах.
const minBodyWidth = 30

// Lipgloss-стили для блоков. Все цвета — 256-цветные коды, чтобы не зависеть
// от темы терминала.
var (
	styleGray       = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	styleYellow     = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	styleYellowBold = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	styleCyan       = lipgloss.NewStyle().Foreground(lipgloss.Color("51"))
	styleRed        = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	styleWhite      = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
)

func IsTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func IsStderrTTY() bool {
	return term.IsTerminal(int(os.Stderr.Fd()))
}

// TermWidth — ширина терминала в колонках. Если не получилось определить
// (pipe/cron) — 80.
func TermWidth() int {
	for _, fd := range []int{int(os.Stderr.Fd()), int(os.Stdout.Fd())} {
		if w, _, err := term.GetSize(fd); err == nil && w > 0 {
			return w
		}
	}
	return 80
}

// bodyWidth — ширина текстового блока с учётом 14-символьного префикса.
func bodyWidth() int {
	w := TermWidth() - indentWidth
	if w < minBodyWidth {
		w = minBodyWidth
	}
	return w
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

func Step(s string)     { fmt.Fprintf(os.Stderr, "  %s  %s\n", colorize(cyan, "→"), s) }
func OK(s string)       { fmt.Fprintf(os.Stderr, "  %s  %s\n", colorize(yellowBold, "✓"), s) }
func Info(s string)     { fmt.Fprintf(os.Stderr, "  %s  %s\n", colorize(gray, "ⓘ"), colorize(gray, s)) }
func Warn(s string)     { fmt.Fprintf(os.Stderr, "  %s  %s\n", colorize(yellow, "⚠"), s) }
func Err(s string)      { fmt.Fprintf(os.Stderr, "  %s  %s\n", colorize(red, "✗"), s) }
func KV(k, v string)    { fmt.Fprintf(os.Stderr, "    %s  %s\n", colorize(gray, padRight(k+":", 14)), v) }
func Plain(s string)    { fmt.Fprintln(os.Stderr, s) }
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

// StopOK останавливает спиннер. Если msg пустая — НИЧЕГО не печатает,
// только убирает текущую анимацию.
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
	top := "╭─ " + b.Title + " " + strings.Repeat("─", maxInt(0, width-len(b.Title)-3)) + "╮"
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

// ChoiceOrCustom — как Choice, но дополнительно позволяет ввести произвольное
// имя варианта.
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
	return v
}

// Choice — нумерованный выбор с дефолтом.
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// linePrefix возвращает единый префикс для всех событий wtf:
// "[HH:MM:SS] <icon>  ". Длина 14 видимых символов (см. indentWidth).
func linePrefix(icon string, iconColor string) string {
	ts := time.Now().Format("15:04:05")
	return fmt.Sprintf("%s %s ",
		colorize(gray, "["+ts+"]"),
		colorize(iconColor, icon))
}

// indentStr — пробельный отступ той же ширины что и linePrefix.
func indentStr() string {
	return strings.Repeat(" ", indentWidth)
}

// wrapBody переносит длинные строки по словам так, чтобы текст не выходил за
// (ширина терминала - indentWidth). Сохраняет существующие переводы строк
// (списки, абзацы) — только разбивает слишком длинные строки. Каждая строка
// продолжения получает 14-пробельный отступ снаружи (через prefixLines),
// здесь мы только обеспечиваем правильную ширину.
func wrapBody(text string, width int) string {
	if width <= 0 {
		return text
	}
	var out strings.Builder
	for i, line := range strings.Split(text, "\n") {
		if i > 0 {
			out.WriteByte('\n')
		}
		out.WriteString(wrapLine(line, width))
	}
	return out.String()
}

// wrapLine переносит одну строку по пробелам, не разрывая слова если возможно.
// ANSI-последовательности игнорируются при подсчёте ширины.
func wrapLine(line string, width int) string {
	if visualLen(line) <= width {
		return line
	}
	words := strings.Fields(line)
	if len(words) == 0 {
		return line
	}
	// Сохраним лидирующие пробелы (markdown-список, indent).
	leading := ""
	for _, r := range line {
		if r == ' ' || r == '\t' {
			leading += string(r)
			continue
		}
		break
	}
	var out strings.Builder
	col := visualLen(leading)
	out.WriteString(leading)
	for i, w := range words {
		wl := visualLen(w)
		if i == 0 {
			out.WriteString(w)
			col += wl
			continue
		}
		if col+1+wl > width {
			out.WriteByte('\n')
			out.WriteString(leading)
			col = visualLen(leading)
			out.WriteString(w)
			col += wl
		} else {
			out.WriteByte(' ')
			out.WriteString(w)
			col += 1 + wl
		}
	}
	return out.String()
}

// prefixLines добавляет prefix перед каждой строкой text. Используется для
// 14-пробельного отступа под продолжение блоков.
func prefixLines(text, prefix string) string {
	if text == "" {
		return ""
	}
	var out strings.Builder
	for i, line := range strings.Split(text, "\n") {
		if i > 0 {
			out.WriteByte('\n')
		}
		out.WriteString(prefix)
		out.WriteString(line)
	}
	return out.String()
}

// CommandHeader — verbose-режим: показываем reason + команду перед запуском.
func CommandHeader(reason, command string) {
	out := os.Stderr
	fmt.Fprintln(out)
	bw := bodyWidth()
	if reason != "" {
		wrapped := wrapBody(reason, bw)
		lines := strings.Split(wrapped, "\n")
		fmt.Fprintf(out, "%s%s\n", linePrefix("→", cyan), colorize(gray, lines[0]))
		for _, l := range lines[1:] {
			fmt.Fprintf(out, "%s%s\n", indentStr(), colorize(gray, l))
		}
	}
	cmdWrapped := wrapBody(command, bw)
	cmdLines := strings.Split(cmdWrapped, "\n")
	fmt.Fprintf(out, "%s%s\n", linePrefix("$", yellowBold), colorize(white, cmdLines[0]))
	for _, l := range cmdLines[1:] {
		fmt.Fprintf(out, "%s%s\n", indentStr(), colorize(white, l))
	}
}

// CommandLineQuiet печатает финальную строку об уже выполненной команде.
// Если строка не помещается — переносит части (reason / команда) с правильным
// 14-пробельным отступом продолжения.
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
	sep := colorize(gray, " · ")
	body := strings.Join(parts, sep)

	bw := bodyWidth()
	if visualLen(stripANSI(body)) <= bw {
		fmt.Fprintf(out, "%s%s\n", linePrefix(icon, color), body)
	} else {
		// Разбиваем по разделителям между частями: первая часть с временным
		// префиксом, остальные с 14-пробельным отступом.
		fmt.Fprintf(out, "%s%s\n", linePrefix(icon, color), parts[0])
		for _, p := range parts[1:] {
			fmt.Fprintf(out, "%s%s\n", indentStr(), p)
		}
	}

	// При ошибке показываем последние 5 строк вывода.
	if (exit != 0 || timedOut) && output != "" {
		lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
		const tailLines = 5
		start := 0
		indent := indentStr()
		if len(lines) > tailLines {
			start = len(lines) - tailLines
			fmt.Fprintf(out, "%s%s\n", indent, colorize(gray, fmt.Sprintf("(показаны последние %d из %d строк)", tailLines, len(lines))))
		}
		for _, l := range lines[start:] {
			fmt.Fprintf(out, "%s%s %s\n", indent, colorize(gray, "│"), colorize(gray, l))
		}
	}
}

// stripANSI — убирает ANSI escape-коды для измерения ширины.
func stripANSI(s string) string {
	var b strings.Builder
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
		b.WriteRune(r)
	}
	return b.String()
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
func UserCommandBlock(reason, command string) {
	out := os.Stderr
	indent := indentStr()
	bw := bodyWidth()
	fmt.Fprintln(out)
	fmt.Fprintf(out, "%s%s\n", linePrefix("⚠", yellow),
		colorize(yellow, "выполни сам (требует sudo / меняет систему):"))
	if reason != "" {
		for _, l := range strings.Split(wrapBody(reason, bw), "\n") {
			fmt.Fprintf(out, "%s%s\n", indent, colorize(gray, l))
		}
	}
	cmdLines := strings.Split(wrapBody(command, bw-2), "\n") // -2 под "$ "
	fmt.Fprintf(out, "%s%s %s\n", indent, colorize(yellowBold, "$"), colorize(white, cmdLines[0]))
	for _, l := range cmdLines[1:] {
		fmt.Fprintf(out, "%s  %s\n", indent, colorize(white, l))
	}
}

// RefusedBlock — мы отказали в авто-запуске.
func RefusedBlock(command, reason string) {
	out := os.Stderr
	fmt.Fprintf(out, "%s%s: %s\n", linePrefix("✗", red), reason, colorize(gray, command))
}

// FinalBlock — финальный ответ агента. Печатается одной сплошной "лентой":
//
//	[HH:MM:SS] ★ Ответ: первая строка ответа, далее идут все
//	                    остальные слова через пробел с word-wrap
//	                    и выровнены под колонку текста.
//
// Все переводы строк / параграфы / списки из текста модели схлопываются в
// один поток слов — никаких пустых строк внутри ответа. Code-блоки fenced
// (```) выводятся отдельно после ленты текста, но без пустой строки между
// ними и хвостом текста.
func FinalBlock(text string) {
	out := os.Stderr
	fmt.Fprintln(out)

	prose, codes := flattenFinal(text)

	headerLabel := "Ответ: "
	header := linePrefix("★", yellowBold) + colorize(yellowBold, headerLabel)
	textIndent := strings.Repeat(" ", indentWidth+visualLen(headerLabel))
	bw := TermWidth() - len(textIndent)
	if bw < minBodyWidth {
		bw = minBodyWidth
	}

	if prose == "" && len(codes) == 0 {
		fmt.Fprintln(out, header)
		return
	}

	// Сплошной wrap всего текста. Первая строка идёт после header,
	// остальные — с textIndent.
	if prose != "" {
		body := highlightInline(prose)
		wrapped := wrapForFinal(body, bw)
		for j, l := range strings.Split(wrapped, "\n") {
			if j == 0 {
				fmt.Fprintln(os.Stdout, header+l)
			} else {
				fmt.Fprintln(os.Stdout, textIndent+l)
			}
		}
	} else {
		// Текста нет, только code — печатаем header один.
		fmt.Fprintln(os.Stdout, header)
	}

	for _, c := range codes {
		body := strings.TrimRight(c, "\n")
		lines := strings.Split(body, "\n")
		for j, l := range lines {
			if j == 0 {
				fmt.Fprintln(os.Stdout, textIndent+colorize(yellowBold, "$ ")+colorize(white, l))
			} else {
				fmt.Fprintln(os.Stdout, textIndent+"  "+colorize(white, l))
			}
		}
	}
	fmt.Fprintln(out)
}

// flattenFinal схлопывает текст модели в один параграф (prose) и список
// fenced-code-блоков (codes). Списки markdown превращаются в "1) item, 2) item"
// внутри prose; code-блоки вынимаются и идут отдельно.
func flattenFinal(text string) (prose string, codes []string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", nil
	}
	lines := strings.Split(text, "\n")
	var words []string
	listIdx := 0
	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			i++
			var code []string
			for i < len(lines) {
				if strings.HasPrefix(strings.TrimSpace(lines[i]), "```") {
					i++
					break
				}
				code = append(code, lines[i])
				i++
			}
			codes = append(codes, strings.Join(code, "\n"))
			continue
		}

		if trimmed == "" {
			i++
			continue
		}

		// Заголовок markdown — снимаем "#", оставляем как обычный текст.
		if strings.HasPrefix(trimmed, "#") {
			trimmed = strings.TrimLeft(trimmed, "# ")
		}

		// Маркер списка → переводим в "1) текст" (или просто " · текст" для bullet).
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			trimmed = " · " + trimmed[2:]
		} else if isListItem(trimmed) {
			body := stripListMarker(trimmed)
			listIdx++
			trimmed = fmt.Sprintf("%d) %s", listIdx, body)
		}

		words = append(words, trimmed)
		i++
	}
	prose = strings.Join(words, " ")
	return prose, codes
}

func isListItem(s string) bool {
	if strings.HasPrefix(s, "- ") || strings.HasPrefix(s, "* ") {
		return true
	}
	// "1. ", "2) "
	for i, r := range s {
		if r >= '0' && r <= '9' {
			continue
		}
		if i > 0 && (r == '.' || r == ')') && i+1 < len(s) && s[i+1] == ' ' {
			return true
		}
		break
	}
	return false
}

func stripListMarker(s string) string {
	if strings.HasPrefix(s, "- ") {
		return strings.TrimPrefix(s, "- ")
	}
	if strings.HasPrefix(s, "* ") {
		return strings.TrimPrefix(s, "* ")
	}
	for i, r := range s {
		if r >= '0' && r <= '9' {
			continue
		}
		if i > 0 && (r == '.' || r == ')') && i+1 < len(s) && s[i+1] == ' ' {
			return s[i+2:]
		}
		break
	}
	return s
}

// highlightInline подсвечивает `inline code` голубым и **bold** жёлто-жирным.
func highlightInline(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		// **bold**
		if i+1 < len(s) && s[i] == '*' && s[i+1] == '*' {
			end := strings.Index(s[i+2:], "**")
			if end >= 0 {
				inner := s[i+2 : i+2+end]
				b.WriteString(colorize(yellowBold, inner))
				i += 2 + end + 2
				continue
			}
		}
		// `code`
		if s[i] == '`' {
			end := strings.Index(s[i+1:], "`")
			if end >= 0 {
				inner := s[i+1 : i+1+end]
				b.WriteString(colorize(cyan, inner))
				i += 1 + end + 1
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// wrapForFinal переносит текст по словам с учётом ANSI escape-кодов.
// Не разрывает слова (даже длинные команды/URL), но не вставляет лишних
// пробелов в начале строки продолжения — отступ выставляется снаружи.
func wrapForFinal(text string, width int) string {
	if width <= 0 {
		return text
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}
	var out strings.Builder
	col := 0
	for i, w := range words {
		wl := visualLen(w)
		if i == 0 {
			out.WriteString(w)
			col = wl
			continue
		}
		if col+1+wl > width {
			out.WriteByte('\n')
			out.WriteString(w)
			col = wl
		} else {
			out.WriteByte(' ')
			out.WriteString(w)
			col += 1 + wl
		}
	}
	return out.String()
}

// FinalBodyWidth — ширина текста для финального ответа (для совместимости с
// caller'ом, но главная логика wrap — внутри FinalBlock).
func FinalBodyWidth() int {
	return bodyWidth()
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
