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
	fmt.Fprintln(out, colorize(yellowBold, "  🤬  "+title))
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
	if !IsStderrTTY() || os.Getenv("NO_COLOR") != "" {
		fmt.Fprintf(os.Stderr, "  %s %s\n", colorize(cyan, "→"), s.prefix)
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
				fmt.Fprintf(os.Stderr, "%s  %s  %s", clearLine, colorize(yellow, spinFrames[i%len(spinFrames)]), s.prefix)
				i++
			}
		}
	}()
}

func (s *Spinner) Update(prefix string) {
	s.prefix = prefix
}

func (s *Spinner) StopOK(msg string) {
	if !s.live.Load() {
		OK(msg)
		return
	}
	close(s.stop)
	<-s.done
	OK(msg)
}

func (s *Spinner) StopFail(msg string) {
	if !s.live.Load() {
		Err(msg)
		return
	}
	close(s.stop)
	<-s.done
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
func Choice(reader *bufio.Reader, label string, options []string, def string) string {
	fmt.Fprintf(os.Stderr, "%s %s\n", colorize(yellow, "?"), label)
	for i, o := range options {
		marker := " "
		if o == def {
			marker = colorize(yellowBold, "›")
		}
		fmt.Fprintf(os.Stderr, "  %s %d) %s\n", marker, i+1, o)
	}
	fmt.Fprintf(os.Stderr, "  %s ", colorize(gray, "выбор:"))
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
