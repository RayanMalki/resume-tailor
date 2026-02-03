package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

func (sr *statusRecorder) Write(b []byte) (int, error) {
	// If WriteHeader wasn't called, status is implicitly 200
	if sr.status == 0 {
		sr.status = http.StatusOK
	}
	return sr.ResponseWriter.Write(b)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{
			ResponseWriter: w,
			status:         0, // 0 means unset
		}

		next.ServeHTTP(recorder, r)

		// If status was never set, default to 200 (HTTP spec)
		if recorder.status == 0 {
			recorder.status = http.StatusOK
		}

		duration := time.Since(start)
		userID, ok := UserIDFromContext(r.Context())
		userIDVal := ""
		if ok {
			userIDVal = userID.String()
		}
		ip := ClientIP(r)

		level := slog.LevelInfo
		if recorder.status == http.StatusTooManyRequests || isSuspiciousPath(r.URL.Path) {
			level = slog.LevelWarn
		}
		if recorder.status >= 500 {
			level = slog.LevelError
		}

		slog.Log(r.Context(), level, "request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.status,
			"latency_ms", duration.Milliseconds(),
			"ip", ip,
			"user_agent", r.UserAgent(),
			"user_id", userIDVal,
		)
	})
}
