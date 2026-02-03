package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"resume-tailor/internal/notify"

	"github.com/google/uuid"
)

type LimitConfig struct {
	Limit  int
	Window time.Duration
}

type RateLimitConfig struct {
	Global    LimitConfig
	Login     LimitConfig
	Signup    LimitConfig
	Uploads   LimitConfig
	RunPerMin LimitConfig
	RunPerDay LimitConfig
}

type RateLimitResult struct {
	Limit  int
	Window time.Duration
	Remain int
}

type RateLimiter struct {
	global    *Limiter
	login     *Limiter
	signup    *Limiter
	uploads   *Limiter
	runPerMin *Limiter
	runPerDay *Limiter
}

var defaultRateLimiter = NewRateLimiter(defaultRateLimitConfig())

func defaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Global:    LimitConfig{Limit: 60, Window: time.Minute},
		Login:     LimitConfig{Limit: 5, Window: time.Minute},
		Signup:    LimitConfig{Limit: 3, Window: time.Hour},
		Uploads:   LimitConfig{Limit: 10, Window: time.Minute},
		RunPerMin: LimitConfig{Limit: 1, Window: time.Minute},
		RunPerDay: LimitConfig{Limit: 10, Window: 24 * time.Hour},
	}
}

func NewRateLimiter(cfg RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		global:    NewLimiter(cfg.Global.Limit, cfg.Global.Window),
		login:     NewLimiter(cfg.Login.Limit, cfg.Login.Window),
		signup:    NewLimiter(cfg.Signup.Limit, cfg.Signup.Window),
		uploads:   NewLimiter(cfg.Uploads.Limit, cfg.Uploads.Window),
		runPerMin: NewLimiter(cfg.RunPerMin.Limit, cfg.RunPerMin.Window),
		runPerDay: NewLimiter(cfg.RunPerDay.Limit, cfg.RunPerDay.Window),
	}
}

func RateLimit(next http.Handler) http.Handler {
	return defaultRateLimiter.Middleware(next)
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := ClientIP(r)

		if allowed, res := rl.global.Allow(ip); !allowed {
			rl.respondRateLimited(w, r, res)
			return
		}

		if matchRoute(r, "/v1/auth/login") && r.Method == http.MethodPost {
			if allowed, res := rl.login.Allow(ip); !allowed {
				rl.respondRateLimited(w, r, res)
				return
			}
		}

		if matchRoute(r, "/v1/auth/signup") && r.Method == http.MethodPost {
			if allowed, res := rl.signup.Allow(ip); !allowed {
				rl.respondRateLimited(w, r, res)
				return
			}
		}

		if matchRoute(r, "/v1/resumes") && r.Method == http.MethodPost {
			if allowed, res := rl.uploads.Allow(ip); !allowed {
				rl.respondRateLimited(w, r, res)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func AllowRunCreate(userID uuid.UUID) (bool, RateLimitResult) {
	return defaultRateLimiter.AllowRunCreate(userID)
}

func (rl *RateLimiter) AllowRunCreate(userID uuid.UUID) (bool, RateLimitResult) {
	if userID == uuid.Nil {
		return false, RateLimitResult{Limit: 0, Window: 0}
	}
	key := "user:" + userID.String()
	if allowed, res := rl.runPerMin.Allow(key); !allowed {
		return false, res
	}
	if allowed, res := rl.runPerDay.Allow(key); !allowed {
		return false, res
	}
	return true, RateLimitResult{}
}

func (rl *RateLimiter) respondRateLimited(w http.ResponseWriter, r *http.Request, res RateLimitResult) {
	meta := map[string]string{
		"limit":     fmt.Sprintf("%d", res.Limit),
		"remaining": fmt.Sprintf("%d", res.Remain),
		"window":    res.Window.String(),
	}
	ip := ClientIP(r)
	go func() {
		_ = notify.SendEvent(context.Background(), notify.Event{
			Type:      "rate_limited",
			Path:      r.URL.Path,
			Method:    r.Method,
			Status:    http.StatusTooManyRequests,
			IP:        ip,
			UserAgent: r.UserAgent(),
			UserID:    "",
			Meta:      meta,
			TS:        time.Now().UTC(),
		})
	}()

	writeJSONError(w, http.StatusTooManyRequests, "rate_limited")
}

func matchRoute(r *http.Request, prefix string) bool {
	return strings.HasPrefix(r.URL.Path, prefix)
}

type Limiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	limit   int
	window  time.Duration
	rate    float64
}

type bucket struct {
	tokens float64
	last   time.Time
}

func NewLimiter(limit int, window time.Duration) *Limiter {
	rate := float64(limit) / window.Seconds()
	return &Limiter{
		buckets: make(map[string]*bucket),
		limit:   limit,
		window:  window,
		rate:    rate,
	}
}

func (l *Limiter) Allow(key string) (bool, RateLimitResult) {
	if l == nil {
		return true, RateLimitResult{}
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, ok := l.buckets[key]
	if !ok {
		l.buckets[key] = &bucket{tokens: float64(l.limit - 1), last: now}
		return true, RateLimitResult{Limit: l.limit, Window: l.window, Remain: l.limit - 1}
	}

	elapsed := now.Sub(b.last).Seconds()
	if elapsed > 0 {
		b.tokens = minFloat(float64(l.limit), b.tokens+elapsed*l.rate)
		b.last = now
	}

	if b.tokens < 1 {
		return false, RateLimitResult{Limit: l.limit, Window: l.window, Remain: 0}
	}

	b.tokens -= 1
	return true, RateLimitResult{Limit: l.limit, Window: l.window, Remain: int(b.tokens)}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
