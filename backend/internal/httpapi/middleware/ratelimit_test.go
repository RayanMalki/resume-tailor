package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRateLimitLogin(t *testing.T) {
	cfg := RateLimitConfig{
		Global:    LimitConfig{Limit: 100, Window: time.Minute},
		Login:     LimitConfig{Limit: 2, Window: time.Minute},
		Signup:    LimitConfig{Limit: 100, Window: time.Hour},
		Uploads:   LimitConfig{Limit: 100, Window: time.Minute},
		RunPerMin: LimitConfig{Limit: 100, Window: time.Minute},
		RunPerDay: LimitConfig{Limit: 1000, Window: 24 * time.Hour},
	}
	rl := NewRateLimiter(cfg)

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(`{}`))
		req.RemoteAddr = "1.2.3.4:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected %d, got %d", http.StatusNoContent, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(`{}`))
	req.RemoteAddr = "1.2.3.4:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected %d, got %d", http.StatusTooManyRequests, rec.Code)
	}
}
