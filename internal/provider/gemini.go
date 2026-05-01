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

type geminiClient struct {
	apiKey string
	model  string
}

func (c *geminiClient) Name() string { return "gemini" }

type geminiFunctionCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args,omitempty"`
}

type geminiFunctionResponse struct {
	Name     string         `json:"name"`
	Response map[string]any `json:"response"`
}

type geminiPart struct {
	Text             string                  `json:"text,omitempty"`
	FunctionCall     *geminiFunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *geminiFunctionResponse `json:"functionResponse,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiFuncDecl struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type geminiToolDef struct {
	FunctionDeclarations []geminiFuncDecl `json:"functionDeclarations"`
}

type geminiGenCfg struct {
	Temperature     float64 `json:"temperature"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
}

type geminiReq struct {
	SystemInstruction *geminiContent  `json:"systemInstruction,omitempty"`
	Contents          []geminiContent `json:"contents"`
	Tools             []geminiToolDef `json:"tools,omitempty"`
	GenerationConfig  geminiGenCfg    `json:"generationConfig"`
}

type geminiResp struct {
	Candidates []struct {
		Content      geminiContent `json:"content"`
		FinishReason string        `json:"finishReason"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// Gemini не имеет concept'а tool_call_id — функции матчатся по имени.
// Для совместимости с нашим интерфейсом мы храним synthesised id в Tool-сообщениях,
// но при отправке в Gemini используем только имя функции.
func (c *geminiClient) Chat(ctx context.Context, req ChatRequest) (*Response, error) {
	maxTok := req.MaxTokens
	if maxTok == 0 {
		maxTok = 2048
	}

	body := geminiReq{
		Contents:         toGeminiContents(req.Messages),
		GenerationConfig: geminiGenCfg{Temperature: 0.2, MaxOutputTokens: maxTok},
	}
	if req.System != "" {
		body.SystemInstruction = &geminiContent{
			Parts: []geminiPart{{Text: req.System}},
		}
	}
	if len(req.Tools) > 0 {
		decls := make([]geminiFuncDecl, 0, len(req.Tools))
		for _, t := range req.Tools {
			decls = append(decls, geminiFuncDecl{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Schema,
			})
		}
		body.Tools = []geminiToolDef{{FunctionDeclarations: decls}}
	}

	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", c.model, c.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	var parsed geminiResp
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("gemini: parse response (status %d): %w", resp.StatusCode, err)
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("gemini: %s", parsed.Error.Message)
	}
	if resp.StatusCode >= 400 || len(parsed.Candidates) == 0 {
		return nil, fmt.Errorf("gemini: HTTP %d: %s", resp.StatusCode, string(data))
	}

	cand := parsed.Candidates[0]
	out := &Response{StopReason: mapGeminiStopReason(cand.FinishReason)}
	for i, p := range cand.Content.Parts {
		if p.Text != "" {
			out.Text += p.Text
		}
		if p.FunctionCall != nil {
			out.ToolCalls = append(out.ToolCalls, ToolCall{
				// Gemini не возвращает id, но нам он нужен для маппинга
				// tool_result обратно. Используем имя+индекс — этого достаточно
				// для пары call→result в одном раунде.
				ID:    fmt.Sprintf("%s_%d", p.FunctionCall.Name, i),
				Name:  p.FunctionCall.Name,
				Input: p.FunctionCall.Args,
			})
		}
	}
	if len(out.ToolCalls) > 0 {
		out.StopReason = "tool_use"
	}
	return out, nil
}

func toGeminiContents(msgs []Message) []geminiContent {
	out := make([]geminiContent, 0, len(msgs))
	for _, m := range msgs {
		switch m.Role {
		case RoleUser:
			out = append(out, geminiContent{
				Role:  "user",
				Parts: []geminiPart{{Text: m.Text}},
			})
		case RoleAssistant:
			parts := []geminiPart{}
			if m.Text != "" {
				parts = append(parts, geminiPart{Text: m.Text})
			}
			for _, tc := range m.ToolCalls {
				parts = append(parts, geminiPart{
					FunctionCall: &geminiFunctionCall{
						Name: tc.Name,
						Args: tc.Input,
					},
				})
			}
			out = append(out, geminiContent{Role: "model", Parts: parts})
		case RoleTool:
			// Имя функции хранится в ToolCallID (см. agent — он положит туда имя).
			// Это особенность gemini — нет id, есть только name.
			out = append(out, geminiContent{
				Role: "user",
				Parts: []geminiPart{{
					FunctionResponse: &geminiFunctionResponse{
						Name: m.ToolCallID,
						Response: map[string]any{
							"output": m.Text,
						},
					},
				}},
			})
		}
	}
	return out
}

func mapGeminiStopReason(r string) string {
	switch r {
	case "STOP":
		return "end_turn"
	case "MAX_TOKENS":
		return "max_tokens"
	case "TOOL_CALL", "FUNCTION_CALL":
		return "tool_use"
	}
	return r
}
