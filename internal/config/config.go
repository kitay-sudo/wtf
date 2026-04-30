package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Provider string

const (
	ProviderClaude Provider = "claude"
	ProviderOpenAI Provider = "openai"
	ProviderGemini Provider = "gemini"
)

type ProviderConfig struct {
	APIKey string `json:"api_key,omitempty"`
	Model  string `json:"model"`
}

type Config struct {
	DefaultProvider Provider                    `json:"default_provider"`
	Language        string                      `json:"language"`
	Providers       map[Provider]ProviderConfig `json:"providers"`
	RedactionShown  bool                        `json:"redaction_consent_shown"`
	CacheEnabled    bool                        `json:"cache_enabled"`
}

func Default() *Config {
	return &Config{
		DefaultProvider: ProviderClaude,
		Language:        "ru",
		CacheEnabled:    true,
		Providers: map[Provider]ProviderConfig{
			ProviderClaude: {Model: "claude-haiku-4-5-20251001"},
			ProviderOpenAI: {Model: "gpt-4o-mini"},
			ProviderGemini: {Model: "gemini-2.0-flash"},
		},
	}
}

func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".wtf")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() (*Config, error) {
	p, err := path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return Default(), nil
	}
	if err != nil {
		return nil, err
	}
	cfg := Default()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Providers == nil {
		cfg.Providers = Default().Providers
	}
	return cfg, nil
}

func (c *Config) Save() error {
	p, err := path()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o600)
}

func (c *Config) APIKey(p Provider) string {
	if pc, ok := c.Providers[p]; ok && pc.APIKey != "" {
		return pc.APIKey
	}
	switch p {
	case ProviderClaude:
		return os.Getenv("ANTHROPIC_API_KEY")
	case ProviderOpenAI:
		return os.Getenv("OPENAI_API_KEY")
	case ProviderGemini:
		if v := os.Getenv("GEMINI_API_KEY"); v != "" {
			return v
		}
		return os.Getenv("GOOGLE_API_KEY")
	}
	return ""
}

func (c *Config) Model(p Provider) string {
	if pc, ok := c.Providers[p]; ok && pc.Model != "" {
		return pc.Model
	}
	return Default().Providers[p].Model
}
