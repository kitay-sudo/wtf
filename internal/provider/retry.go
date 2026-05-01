package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// MaxRateLimitRetries — сколько раз пробуем повторить запрос при 429.
// Лимит 3 — больше уже бесполезно: либо у юзера действительно мало квоты,
// либо он бьёт спам-частотой и пора падать с понятной ошибкой.
const MaxRateLimitRetries = 3

// MaxRetryWait — потолок на одно ожидание между попытками. Если провайдер
// просит ждать дольше — отказываемся и пробрасываем ошибку, чтобы юзер
// решил сам (попозже / переключиться на другую модель).
const MaxRetryWait = 30 * time.Second

// rateLimitError — соглашение между HTTP-парсингом клиентов и retry-логикой.
// Каждый клиент при HTTP 429 должен возвращать ошибку которая Is(rateLimitError),
// и опционально содержит RetryAfter из заголовка / тела ответа.
type rateLimitError struct {
	Provider   string
	RetryAfter time.Duration
	Body       string
}

func (e *rateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("%s: rate limit (повтор через %s): %s",
			e.Provider, e.RetryAfter, e.Body)
	}
	return fmt.Sprintf("%s: rate limit: %s", e.Provider, e.Body)
}

// errIsRateLimit — sentinel для errors.Is.
var errIsRateLimit = errors.New("rate_limit")

func (e *rateLimitError) Is(target error) bool {
	return target == errIsRateLimit
}

// IsRateLimit публичный helper — agent его не использует, только внутри пакета.
func IsRateLimit(err error) bool {
	return errors.Is(err, errIsRateLimit)
}

// parseRateLimit извлекает время ожидания.
// Источники по приоритету:
//  1. HTTP Retry-After (стандартный заголовок)
//  2. x-ratelimit-reset-* у openai
//  3. парс "try again in 6.7s" из текста ошибки
//  4. fallback 5 секунд
func parseRateLimit(resp *http.Response, body string) time.Duration {
	if resp != nil {
		if v := resp.Header.Get("Retry-After"); v != "" {
			if secs, err := strconv.Atoi(v); err == nil {
				return time.Duration(secs) * time.Second
			}
			// Может быть HTTP-date, но провайдеры обычно дают секунды.
		}
		// OpenAI шлёт x-ratelimit-reset-tokens вида "6.7s" или "1m30s".
		for _, h := range []string{"x-ratelimit-reset-tokens", "x-ratelimit-reset-requests"} {
			if v := resp.Header.Get(h); v != "" {
				if d, err := time.ParseDuration(v); err == nil {
					return d
				}
			}
		}
	}
	// Текст: "Please try again in 6.718s." / "in 1m23s"
	if idx := strings.Index(body, "try again in "); idx >= 0 {
		rest := body[idx+len("try again in "):]
		end := strings.IndexAny(rest, ".\"`'\n ")
		if end > 0 {
			rest = rest[:end]
		}
		if d, err := time.ParseDuration(strings.TrimSpace(rest)); err == nil {
			return d
		}
	}
	return 5 * time.Second
}

// withRetry оборачивает HTTP-вызов в retry-цикл при rate-limit ошибках.
// Каждый клиент использует это: вызывает doFunc, если получает rateLimitError —
// ждёт указанное время и повторяет, до MaxRateLimitRetries попыток.
//
// Если ctx истекает или провайдер просит ждать > MaxRetryWait — возвращаем
// последнюю ошибку без дальнейших попыток.
func withRetry[T any](
	ctx context.Context,
	onWait func(d time.Duration, attempt int),
	doFunc func() (T, error),
) (T, error) {
	var last T
	var err error
	for attempt := 1; attempt <= MaxRateLimitRetries+1; attempt++ {
		last, err = doFunc()
		if err == nil {
			return last, nil
		}
		var rle *rateLimitError
		if !errors.As(err, &rle) {
			return last, err
		}
		if attempt > MaxRateLimitRetries {
			return last, err
		}
		wait := rle.RetryAfter
		if wait <= 0 {
			wait = 5 * time.Second
		}
		// Небольшой буфер сверху — провайдер часто округляет вниз.
		wait += 500 * time.Millisecond
		if wait > MaxRetryWait {
			return last, fmt.Errorf("%w (отказ от автоматического повтора: ждать %s — слишком долго)",
				err, wait)
		}
		if onWait != nil {
			onWait(wait, attempt)
		}
		select {
		case <-ctx.Done():
			return last, ctx.Err()
		case <-time.After(wait):
		}
	}
	return last, err
}
