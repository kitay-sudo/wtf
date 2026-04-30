package redact

import (
	"os"
	"regexp"
	"strings"
)

type rule struct {
	name string
	re   *regexp.Regexp
	repl string
}

var rules = []rule{
	{"anthropic_key", regexp.MustCompile(`sk-ant-[A-Za-z0-9_\-]{20,}`), "[REDACTED:anthropic_key]"},
	{"openai_key", regexp.MustCompile(`sk-[A-Za-z0-9_\-]{20,}`), "[REDACTED:api_key]"},
	{"google_key", regexp.MustCompile(`AIza[0-9A-Za-z_\-]{35}`), "[REDACTED:google_key]"},
	{"github_token", regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{20,}`), "[REDACTED:github_token]"},
	{"slack_token", regexp.MustCompile(`xox[baprs]-[A-Za-z0-9\-]{10,}`), "[REDACTED:slack_token]"},
	{"jwt", regexp.MustCompile(`eyJ[A-Za-z0-9_\-]+\.eyJ[A-Za-z0-9_\-]+\.[A-Za-z0-9_\-]+`), "[REDACTED:jwt]"},
	{"bearer", regexp.MustCompile(`(?i)Bearer\s+[A-Za-z0-9_\-\.=]+`), "Bearer [REDACTED]"},
	{"aws_access_key", regexp.MustCompile(`AKIA[0-9A-Z]{16}`), "[REDACTED:aws_access_key]"},
	{"aws_secret", regexp.MustCompile(`(?i)aws_secret[_a-z]*\s*[:=]\s*[A-Za-z0-9/+=]{30,}`), "aws_secret=[REDACTED]"},
	{"private_key", regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----[\s\S]*?-----END [A-Z ]*PRIVATE KEY-----`), "[REDACTED:private_key]"},
	{"password_kv", regexp.MustCompile(`(?i)(password|passwd|pwd|secret|token)\s*[:=]\s*[^\s'"]+`), "$1=[REDACTED]"},
	{"basic_auth_url", regexp.MustCompile(`([a-z]+)://[^:/\s]+:[^@/\s]+@`), "$1://[REDACTED]@"},
	{"email", regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`), "[REDACTED:email]"},
}

type Result struct {
	Text    string
	Hits    map[string]int
}

func Apply(s string) Result {
	hits := map[string]int{}
	out := s
	for _, r := range rules {
		matches := r.re.FindAllString(out, -1)
		if len(matches) > 0 {
			hits[r.name] += len(matches)
			out = r.re.ReplaceAllString(out, r.repl)
		}
	}
	out = redactHomePath(out)
	return Result{Text: out, Hits: hits}
}

func redactHomePath(s string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return s
	}
	return strings.ReplaceAll(s, home, "~")
}

func Summary(r Result) string {
	if len(r.Hits) == 0 {
		return ""
	}
	parts := []string{}
	for name, n := range r.Hits {
		parts = append(parts, name)
		_ = n
	}
	return "redacted: " + strings.Join(parts, ", ")
}
