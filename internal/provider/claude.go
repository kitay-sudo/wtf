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

type claudeMessage struct {
	Role    string          `json:"role"`
	Content []claudeContent `json:"content"`
}

type claudeContent struct {
	Type         string             `json:"type"`
	Text         string             `json:"text"`
	CacheControl *claudeCacheCtrl   `json:"cache_control,omitempty"`
}

type claudeCacheCtrl struct {
	Type string `json:"type"`
}

type claudeReq struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	System    []claudeContent `json:"system"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeResp struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *claudeClient) Explain(ctx context.Context, req Request) (string, error) {
	body := claudeReq{
		Model:     c.model,
		MaxTokens: 1024,
		System: []claudeContent{
			{
				Type:         "text",
				Text:         systemPrompt(req.Language),
				CacheControl: &claudeCacheCtrl{Type: "ephemeral"},
			},
		},
		Messages: []claudeMessage{
			{
				Role: "user",
				Content: []claudeContent{
					{Type: "text", Text: userPrompt(req)},
				},
			},
		},
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	var parsed claudeResp
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", fmt.Errorf("claude: parse response (status %d): %w\nbody: %s", resp.StatusCode, err, string(data))
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("claude: %s: %s", parsed.Error.Type, parsed.Error.Message)
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("claude: HTTP %d: %s", resp.StatusCode, string(data))
	}
	var out string
	for _, c := range parsed.Content {
		if c.Type == "text" {
			out += c.Text
		}
	}
	return out, nil
}
