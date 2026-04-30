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

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiGenCfg struct {
	Temperature     float64 `json:"temperature"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
}

type geminiReq struct {
	SystemInstruction *geminiContent  `json:"systemInstruction,omitempty"`
	Contents          []geminiContent `json:"contents"`
	GenerationConfig  geminiGenCfg    `json:"generationConfig"`
}

type geminiResp struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

func (c *geminiClient) Explain(ctx context.Context, req Request) (string, error) {
	body := geminiReq{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{{Text: systemPrompt(req.Language)}},
		},
		Contents: []geminiContent{
			{Role: "user", Parts: []geminiPart{{Text: userPrompt(req)}}},
		},
		GenerationConfig: geminiGenCfg{Temperature: 0.2, MaxOutputTokens: 1024},
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", c.model, c.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	var parsed geminiResp
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", fmt.Errorf("gemini: parse response (status %d): %w", resp.StatusCode, err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("gemini: %s", parsed.Error.Message)
	}
	if resp.StatusCode >= 400 || len(parsed.Candidates) == 0 {
		return "", fmt.Errorf("gemini: HTTP %d: %s", resp.StatusCode, string(data))
	}
	var out string
	for _, p := range parsed.Candidates[0].Content.Parts {
		out += p.Text
	}
	return out, nil
}
