package render

import (
	"os"
	"regexp"
	"strings"

	"golang.org/x/term"
)

const (
	reset      = "\033[0m"
	bold       = "\033[1m"
	dim        = "\033[2m"
	yellow     = "\033[33m"
	yellowBold = "\033[1;33m"
	cyan       = "\033[36m"
	gray       = "\033[90m"
)

func Markdown(s string) string {
	if !isTTY() || os.Getenv("NO_COLOR") != "" {
		return s
	}
	out := s

	codeBlock := regexp.MustCompile("(?s)```[a-zA-Z0-9_+-]*\n(.*?)```")
	out = codeBlock.ReplaceAllStringFunc(out, func(m string) string {
		sub := codeBlock.FindStringSubmatch(m)
		if len(sub) < 2 {
			return m
		}
		body := strings.TrimRight(sub[1], "\n")
		var b strings.Builder
		b.WriteString("\n")
		for _, line := range strings.Split(body, "\n") {
			b.WriteString("    ")
			b.WriteString(yellow)
			b.WriteString(line)
			b.WriteString(reset)
			b.WriteString("\n")
		}
		return b.String()
	})

	inlineCode := regexp.MustCompile("`([^`]+)`")
	out = inlineCode.ReplaceAllString(out, cyan+"$1"+reset)

	boldRe := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	out = boldRe.ReplaceAllString(out, bold+"$1"+reset)

	headRe := regexp.MustCompile(`(?m)^(#{1,3})\s+(.*)$`)
	out = headRe.ReplaceAllString(out, yellowBold+"$2"+reset)

	return out
}

func isTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
