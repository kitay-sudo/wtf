package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type claudeClient struct {
	apiKey string
	model  string
}

func (c *claudeClient) Name() string { return "claude" }

type claudeContentBlock struct {
	Type      string         `json:"type"`
	Text      string         `json:"text,omitempty"`
	ID        string         `json:"id,omitempty"`
	Name      string         `json:"name,omitempty"`
	Input     map[string]any `json:"input,omitempty"`
	ToolUseID string         `json:"tool_use_id,omitempty"`
	Content   string         `json:"content,omitempty"` // для tool_result
	IsError   bool           `json:"is_error,omitempty"`

	CacheControl *claudeCacheCtrl `json:"cache_control,omitempty"`
}

type claudeCacheCtrl struct {
	Type string `json:"type"`
}

type claudeMessage struct {
	Role    string               `json:"role"`
	Content []claudeContentBlock `json:"content"`
}

type claudeTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

type claudeReq struct {
	Model     string               `json:"model"`
	MaxTokens int                  `json:"max_tokens"`
	System    []claudeContentBlock `json:"system,omitempty"`
	Tools     []claudeTool         `json:"tools,omitempty"`
	Messages  []claudeMessage      `json:"messages"`
}

type claudeResp struct {
	Content    []claudeContentBlock `json:"content"`
	StopReason string               `json:"stop_reason"`
	Error      *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *claudeClient) Chat(ctx context.Context, req ChatRequest) (*Response, error) {
	return withRetry(ctx, req.OnRateLimit, func() (*Response, error) {
		return c.doChat(ctx, req)
	})
}

func (c *claudeClient) doChat(ctx context.Context, req ChatRequest) (*Response, error) {
	maxTok := req.MaxTokens
	if maxTok == 0 {
		maxTok = 2048
	}

	body := claudeReq{
		Model:     c.model,
		MaxTokens: maxTok,
		Messages:  toClaudeMessages(req.Messages),
	}
	if req.System != "" {
		body.System = []claudeContentBlock{{
			Type:         "text",
			Text:         req.System,
			CacheControl: &claudeCacheCtrl{Type: "ephemeral"},
		}}
	}
	for _, t := range req.Tools {
		body.Tools = append(body.Tools, claudeTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.Schema,
		})
	}

	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 429 {
		return nil, &rateLimitError{
			Provider:   "claude",
			RetryAfter: parseRateLimit(resp, string(data)),
			Body:       string(data),
		}
	}

	var parsed claudeResp
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("claude: parse response (status %d): %w\nbody: %s", resp.StatusCode, err, string(data))
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("claude: %s: %s", parsed.Error.Type, parsed.Error.Message)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("claude: HTTP %d: %s", resp.StatusCode, string(data))
	}

	out := &Response{StopReason: parsed.StopReason}
	for _, block := range parsed.Content {
		switch block.Type {
		case "text":
			out.Text += block.Text
		case "tool_use":
			out.ToolCalls = append(out.ToolCalls, ToolCall{
				ID:    block.ID,
				Name:  block.Name,
				Input: block.Input,
			})
		}
	}
	return out, nil
}

// toClaudeMessages переводит универсальные Message в формат Claude.
// Claude хочет content-блоки: text, tool_use (в assistant), tool_result (в user).
// Tool-результаты идут под role="user", это особенность API.
func toClaudeMessages(msgs []Message) []claudeMessage {
	out := make([]claudeMessage, 0, len(msgs))
	for _, m := range msgs {
		switch m.Role {
		case RoleUser:
			out = append(out, claudeMessage{
				Role:    "user",
				Content: []claudeContentBlock{{Type: "text", Text: m.Text}},
			})
		case RoleAssistant:
			blocks := []claudeContentBlock{}
			if m.Text != "" {
				blocks = append(blocks, claudeContentBlock{Type: "text", Text: m.Text})
			}
			for _, tc := range m.ToolCalls {
				blocks = append(blocks, claudeContentBlock{
					Type:  "tool_use",
					ID:    tc.ID,
					Name:  tc.Name,
					Input: tc.Input,
				})
			}
			out = append(out, claudeMessage{Role: "assistant", Content: blocks})
		case RoleTool:
			out = append(out, claudeMessage{
				Role: "user",
				Content: []claudeContentBlock{{
					Type:      "tool_result",
					ToolUseID: m.ToolCallID,
					Content:   m.Text,
				}},
			})
		}
	}
	return out
}
