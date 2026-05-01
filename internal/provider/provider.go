package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/bitcoff/wtf/internal/config"
)

// Role — стандартизованные роли сообщений в чате.
// Маппятся на конкретные роли провайдера в каждом клиенте.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message — одно сообщение в истории чата.
//
// Для роли user/assistant — обычный текст в Text.
// Для роли assistant с вызовом инструмента — заполнен ToolCalls.
// Для роли tool — Text=результат, ToolCallID=id вызова на который отвечаем.
type Message struct {
	Role       Role
	Text       string
	ToolCalls  []ToolCall // только assistant
	ToolCallID string     // только tool
}

// ToolCall — вызов инструмента моделью.
type ToolCall struct {
	ID    string
	Name  string
	Input map[string]any
}

// Tool — описание инструмента который модель может вызвать.
// Schema — JSON Schema объекта параметров.
type Tool struct {
	Name        string
	Description string
	Schema      map[string]any
}

// Response — что вернула модель.
// Если ToolCalls не пуст — модель хочет выполнить инструменты.
// Если StopReason="end_turn" и Text не пуст — финальный ответ.
type Response struct {
	Text       string
	ToolCalls  []ToolCall
	StopReason string // "end_turn" | "tool_use" | "max_tokens" | "stop"
}

// ChatRequest — параметры одного раунда чата.
type ChatRequest struct {
	System    string
	Messages  []Message
	Tools     []Tool
	MaxTokens int

	// OnRateLimit вызывается перед паузой при автоматическом повторе после 429.
	// attempt — номер уже совершённой попытки (1, 2, ...). UI может использовать
	// для показа "ждём 5с (попытка 2/3)..." вместо тихого зависания.
	OnRateLimit func(wait time.Duration, attempt int)
}

type Client interface {
	Chat(ctx context.Context, req ChatRequest) (*Response, error)
	Name() string
}

func New(cfg *config.Config, p config.Provider) (Client, error) {
	key := cfg.APIKey(p)
	if key == "" {
		return nil, fmt.Errorf("API key for %s is not set", p)
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
