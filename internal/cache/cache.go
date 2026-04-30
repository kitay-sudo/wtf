package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/bitcoff/wtf/internal/config"
)

type Entry struct {
	Key       string    `json:"key"`
	Provider  string    `json:"provider"`
	Model     string    `json:"model"`
	Language  string    `json:"language"`
	Answer    string    `json:"answer"`
	CreatedAt time.Time `json:"created_at"`
}

func Key(provider, model, language, output, lastCmd string) string {
	h := sha256.New()
	h.Write([]byte(provider))
	h.Write([]byte{0})
	h.Write([]byte(model))
	h.Write([]byte{0})
	h.Write([]byte(language))
	h.Write([]byte{0})
	h.Write([]byte(lastCmd))
	h.Write([]byte{0})
	h.Write([]byte(output))
	return hex.EncodeToString(h.Sum(nil))
}

func dir() (string, error) {
	base, err := config.Dir()
	if err != nil {
		return "", err
	}
	d := filepath.Join(base, "cache")
	if err := os.MkdirAll(d, 0o755); err != nil {
		return "", err
	}
	return d, nil
}

func Get(key string) (*Entry, bool) {
	d, err := dir()
	if err != nil {
		return nil, false
	}
	data, err := os.ReadFile(filepath.Join(d, key+".json"))
	if err != nil {
		return nil, false
	}
	var e Entry
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, false
	}
	if time.Since(e.CreatedAt) > 30*24*time.Hour {
		return nil, false
	}
	return &e, true
}

func Put(e Entry) error {
	d, err := dir()
	if err != nil {
		return err
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(d, e.Key+".json"), data, 0o600)
}
