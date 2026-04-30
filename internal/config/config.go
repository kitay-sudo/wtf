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
			ProviderClaude: {Model: KnownModels[ProviderClaude][0]},
			ProviderOpenAI: {Model: KnownModels[ProviderOpenAI][0]},
			ProviderGemini: {Model: KnownModels[ProviderGemini][0]},
		},
	}
}

// KnownModels — рекомендуемые модели для каждого провайдера, отсортированные:
// первая — дефолт (быстрая+дешёвая), дальше — варианты «помощнее»,
// последняя — «предыдущий поколение / дешевле».
//
// Эти списки показываются в `wtf config` wizard'е. Юзер может ввести цифру
// для выбора из списка, ввести своё имя модели (для precision/edge cases),
// или нажать Enter чтобы оставить текущее значение.
var KnownModels = map[Provider][]string{
	ProviderClaude: {
		"claude-haiku-4-5-20251001",
		"claude-sonnet-4-6",
		"claude-opus-4-7",
	},
	ProviderOpenAI: {
		"gpt-4o-mini",
		"gpt-4o",
		"gpt-4.1",
		"gpt-4.1-mini",
		"o4-mini",
	},
	ProviderGemini: {
		"gemini-2.0-flash",
		"gemini-2.0-flash-lite",
		"gemini-2.5-pro",
		"gemini-1.5-flash",
		"gemini-1.5-pro",
	},
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
