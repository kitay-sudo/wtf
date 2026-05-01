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

type openaiClient struct {
	apiKey string
	model  string
}

func (c *openaiClient) Name() string { return "openai" }

type openaiToolCall struct {
	ID       string `json:"id,omitempty"`
	Type     string `json:"type,omitempty"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"` // JSON-строка
	} `json:"function"`
}

type openaiMsg struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	Name       string           `json:"name,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
	ToolCalls  []openaiToolCall `json:"tool_calls,omitempty"`
}

type openaiTool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string         `json:"name"`
		Description string         `json:"description,omitempty"`
		Parameters  map[string]any `json:"parameters"`
	} `json:"function"`
}

type openaiReq struct {
	Model       string       `json:"model"`
	Messages    []openaiMsg  `json:"messages"`
	Tools       []openaiTool `json:"tools,omitempty"`
	Temperature float64      `json:"temperature"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
}

type openaiResp struct {
	Choices []struct {
		Message      openaiMsg `json:"message"`
		FinishReason string    `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func (c *openaiClient) Chat(ctx context.Context, req ChatRequest) (*Response, error) {
	return withRetry(ctx, req.OnRateLimit, func() (*Response, error) {
		return c.doChat(ctx, req)
	})
}

func (c *openaiClient) doChat(ctx context.Context, req ChatRequest) (*Response, error) {
	maxTok := req.MaxTokens
	if maxTok == 0 {
		maxTok = 2048
	}

	msgs := []openaiMsg{}
	if req.System != "" {
		msgs = append(msgs, openaiMsg{Role: "system", Content: req.System})
	}
	msgs = append(msgs, toOpenAIMessages(req.Messages)...)

	body := openaiReq{
		Model:       c.model,
		Temperature: 0.2,
		MaxTokens:   maxTok,
		Messages:    msgs,
	}
	for _, t := range req.Tools {
		ot := openaiTool{Type: "function"}
		ot.Function.Name = t.Name
		ot.Function.Description = t.Description
		ot.Function.Parameters = t.Schema
		body.Tools = append(body.Tools, ot)
	}

	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 429 {
		return nil, &rateLimitError{
			Provider:   "openai",
			RetryAfter: parseRateLimit(resp, string(data)),
			Body:       string(data),
		}
	}

	var parsed openaiResp
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("openai: parse response (status %d): %w", resp.StatusCode, err)
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("openai: %s", parsed.Error.Message)
	}
	if resp.StatusCode >= 400 || len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("openai: HTTP %d: %s", resp.StatusCode, string(data))
	}

	choice := parsed.Choices[0]
	out := &Response{
		Text:       choice.Message.Content,
		StopReason: mapOpenAIStopReason(choice.FinishReason),
	}
	for _, tc := range choice.Message.ToolCalls {
		var input map[string]any
		if tc.Function.Arguments != "" {
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &input); err != nil {
				// Если модель вернула невалидный JSON в arguments — кладём как есть,
				// агент потом это увидит и попросит перезапросить.
				input = map[string]any{"_raw": tc.Function.Arguments}
			}
		}
		out.ToolCalls = append(out.ToolCalls, ToolCall{
			ID:    tc.ID,
			Name:  tc.Function.Name,
			Input: input,
		})
	}
	return out, nil
}

func toOpenAIMessages(msgs []Message) []openaiMsg {
	out := make([]openaiMsg, 0, len(msgs))
	for _, m := range msgs {
		switch m.Role {
		case RoleUser:
			out = append(out, openaiMsg{Role: "user", Content: m.Text})
		case RoleAssistant:
			om := openaiMsg{Role: "assistant", Content: m.Text}
			for _, tc := range m.ToolCalls {
				args, _ := json.Marshal(tc.Input)
				oc := openaiToolCall{ID: tc.ID, Type: "function"}
				oc.Function.Name = tc.Name
				oc.Function.Arguments = string(args)
				om.ToolCalls = append(om.ToolCalls, oc)
			}
			out = append(out, om)
		case RoleTool:
			out = append(out, openaiMsg{
				Role:       "tool",
				Content:    m.Text,
				ToolCallID: m.ToolCallID,
			})
		}
	}
	return out
}

func mapOpenAIStopReason(r string) string {
	switch r {
	case "tool_calls":
		return "tool_use"
	case "stop":
		return "end_turn"
	case "length":
		return "max_tokens"
	}
	return r
}
