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

type openaiMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiReq struct {
	Model       string      `json:"model"`
	Messages    []openaiMsg `json:"messages"`
	Temperature float64     `json:"temperature"`
	MaxTokens   int         `json:"max_tokens"`
}

type openaiResp struct {
	Choices []struct {
		Message openaiMsg `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func (c *openaiClient) Explain(ctx context.Context, req Request) (string, error) {
	body := openaiReq{
		Model:       c.model,
		Temperature: 0.2,
		MaxTokens:   1024,
		Messages: []openaiMsg{
			{Role: "system", Content: systemPrompt(req.Language)},
			{Role: "user", Content: userPrompt(req)},
		},
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	var parsed openaiResp
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", fmt.Errorf("openai: parse response (status %d): %w", resp.StatusCode, err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("openai: %s", parsed.Error.Message)
	}
	if resp.StatusCode >= 400 || len(parsed.Choices) == 0 {
		return "", fmt.Errorf("openai: HTTP %d: %s", resp.StatusCode, string(data))
	}
	return parsed.Choices[0].Message.Content, nil
}
