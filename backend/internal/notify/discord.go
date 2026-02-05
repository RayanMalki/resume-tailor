package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Event struct {
	Type      string
	Path      string
	Method    string
	Status    int
	IP        string
	UserAgent string
	UserID    string
	Meta      map[string]string
	TS        time.Time
}

type payload struct {
	Content string `json:"content"`
}

var warnMissingWebhookOnce sync.Once
var dispatcherOnce sync.Once

type queuedEvent struct {
	ctx     context.Context
	event   Event
	webhook string
}

var eventQueue = make(chan queuedEvent, 100)

func SendEvent(ctx context.Context, e Event) error {
	webhook := strings.TrimSpace(os.Getenv("DISCORD_WEBHOOK_URL"))
	if webhook == "" {
		warnMissingWebhookOnce.Do(func() {
			slog.Warn("discord webhook disabled (DISCORD_WEBHOOK_URL not set)")
		})
		return nil
	}

	if e.TS.IsZero() {
		e.TS = time.Now().UTC()
	}

	dispatcherOnce.Do(func() {
		go startDispatcher()
	})

	select {
	case eventQueue <- queuedEvent{ctx: ctx, event: e, webhook: webhook}:
		return nil
	default:
		slog.Warn("discord webhook queue full, dropping event")
		return nil
	}
}

func startDispatcher() {
	minInterval := parseDurationMsEnv("DISCORD_WEBHOOK_MIN_INTERVAL_MS", 1100)
	maxRetryAfter := parseDurationMsEnv("DISCORD_WEBHOOK_MAX_RETRY_AFTER_MS", 120000)
	if minInterval < 0 {
		minInterval = 0
	}
	if maxRetryAfter < 0 {
		maxRetryAfter = 0
	}

	client := &http.Client{Timeout: 10 * time.Second}
	var lastSent time.Time

	for item := range eventQueue {
		if minInterval > 0 && !lastSent.IsZero() {
			wait := minInterval - time.Since(lastSent)
			if wait > 0 {
				time.Sleep(wait)
			}
		}

		content := formatEvent(item.event)
		body, err := json.Marshal(payload{Content: content})
		if err != nil {
			slog.Warn("discord webhook marshal failed", "error", err)
			continue
		}

		req, err := http.NewRequestWithContext(item.ctx, http.MethodPost, item.webhook, bytes.NewReader(body))
		if err != nil {
			slog.Warn("discord webhook request build failed", "error", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		if err := sendWithRetry(item.ctx, client, req, body, maxRetryAfter); err != nil {
			slog.Warn("discord webhook send failed", "error", err)
		}

		lastSent = time.Now()
	}
}

func sendWithRetry(ctx context.Context, client *http.Client, baseReq *http.Request, body []byte, maxRetryAfter time.Duration) error {
	const maxAttempts = 3

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, baseReq.Method, baseReq.URL.String(), bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("build request: %w", err)
		}
		req.Header = baseReq.Header.Clone()

		resp, err := client.Do(req)
		if err != nil {
			slog.Warn("discord webhook send failed", "error", err)
			return fmt.Errorf("send webhook: %w", err)
		}

		bodyPreview, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		_ = resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		if resp.StatusCode == http.StatusTooManyRequests && attempt < maxAttempts {
			retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
			if retryAfter > 0 {
				if maxRetryAfter > 0 && retryAfter > maxRetryAfter {
					slog.Warn("discord webhook rate limited, retry-after too long, dropping", "retry_after", retryAfter)
					return fmt.Errorf("webhook status: %d", resp.StatusCode)
				}
				slog.Warn("discord webhook rate limited", "retry_after", retryAfter)
				if !sleepCtx(ctx, retryAfter) {
					return ctx.Err()
				}
				continue
			}
		}

		slog.Warn("discord webhook non-2xx", "status", resp.StatusCode, "body", string(bodyPreview))
		return fmt.Errorf("webhook status: %d", resp.StatusCode)
	}

	return fmt.Errorf("webhook status: %d", http.StatusTooManyRequests)
}

func parseRetryAfter(value string) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds <= 0 {
			return 0
		}
		return time.Duration(seconds) * time.Second
	}
	if when, err := http.ParseTime(value); err == nil {
		delay := time.Until(when)
		if delay < 0 {
			return 0
		}
		return delay
	}
	return 0
}

func parseDurationMsEnv(key string, defMs int) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return time.Duration(defMs) * time.Millisecond
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return time.Duration(defMs) * time.Millisecond
	}
	return time.Duration(parsed) * time.Millisecond
}

func sleepCtx(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func formatEvent(e Event) string {
	ip := redactIP(e.IP)
	ua := truncate(cleanOneLine(e.UserAgent), 120)

	parts := []string{
		"EVENT",
		e.Type,
		fmt.Sprintf("user=%s", fallback(e.UserID, "anon")),
		fmt.Sprintf("ip=%s", fallback(ip, "unknown")),
		fmt.Sprintf("ua=%s", fallback(ua, "unknown")),
		fmt.Sprintf("status=%d", e.Status),
		fmt.Sprintf("path=%s", fallback(e.Path, "-")),
		fmt.Sprintf("method=%s", fallback(e.Method, "-")),
	}

	if len(e.Meta) > 0 {
		keys := make([]string, 0, len(e.Meta))
		for k := range e.Meta {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		metaParts := make([]string, 0, len(keys))
		for _, k := range keys {
			metaParts = append(metaParts, fmt.Sprintf("%s:%s", k, e.Meta[k]))
		}
		parts = append(parts, "meta="+strings.Join(metaParts, ","))
	}

	return strings.Join(parts, " ")
}

func redactIP(ip string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return ""
	}
	if strings.Count(ip, ".") == 3 {
		parts := strings.Split(ip, ".")
		if len(parts) == 4 {
			return fmt.Sprintf("%s.%s.x.x", parts[0], parts[1])
		}
	}
	if strings.Contains(ip, ":") {
		parts := strings.Split(ip, ":")
		if len(parts) >= 2 {
			return fmt.Sprintf("%s:%s:xxxx", parts[0], parts[1])
		}
	}
	return ip
}

func cleanOneLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.TrimSpace(s)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func fallback(value, alt string) string {
	if strings.TrimSpace(value) == "" {
		return alt
	}
	return value
}
