package provider

import (
	"context"
	"fmt"

	"github.com/bitcoff/wtf/internal/config"
)

type Request struct {
	Language    string
	OS          string
	Shell       string
	Cwd         string
	GitBranch   string
	LastCommand string
	ExitCode    int
	Output      string
	PkgManager  string
}

type Client interface {
	Explain(ctx context.Context, req Request) (string, error)
	Name() string
}

func New(cfg *config.Config, p config.Provider) (Client, error) {
	key := cfg.APIKey(p)
	if key == "" {
		return nil, fmt.Errorf("API key for %s is not set (config or env var)", p)
	}
	model := cfg.Model(p)
	switch p {
	case config.ProviderClaude:
		return &claudeClient{apiKey: key, model: model}, nil
	case config.ProviderOpenAI:
		return &openaiClient{apiKey: key, model: model}, nil
	case config.ProviderGemini:
		return &geminiClient{apiKey: key, model: model}, nil
	}
	return nil, fmt.Errorf("unknown provider: %s", p)
}

func systemPrompt(lang string) string {
	if lang == "ru" {
		return `Ты — эксперт-разработчик. Пользователь получил ошибку в терминале. Объясни кратко, без воды, что случилось и как починить.

Формат ответа строго в Markdown:

**Что случилось:** одно предложение.

**Как починить:**
1. Первый вариант — короткое описание
   ` + "```bash" + `
   команда
   ` + "```" + `
2. Второй вариант (если есть)
3. Третий вариант (если есть)

**Почему:** одна-две строки про причину (опционально).

Если есть релевантная дока — одна ссылка в конце. Не выдумывай ссылки. Не повторяй текст ошибки.`
	}
	return `You are an expert developer. The user got a terminal error. Explain briefly what happened and how to fix it. No fluff.

Strict Markdown format:

**What happened:** one sentence.

**Fix:**
1. First option — short description
   ` + "```bash" + `
   command
   ` + "```" + `
2. Second option (if any)
3. Third option (if any)

**Why:** one or two lines about the cause (optional).

One relevant doc link at the end if you actually know one. Don't invent links. Don't repeat the error text.`
}

func userPrompt(req Request) string {
	return fmt.Sprintf(`Context:
- OS: %s
- Shell: %s
- Cwd: %s
- Git branch: %s
- Package manager: %s
- Last command: %s
- Exit code: %d

Output:
%s`,
		nz(req.OS), nz(req.Shell), nz(req.Cwd), nz(req.GitBranch), nz(req.PkgManager),
		nz(req.LastCommand), req.ExitCode, req.Output)
}

func nz(s string) string {
	if s == "" {
		return "(unknown)"
	}
	return s
}
