package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"resume-tailor/internal/notify"
)

const suspiciousNotifyWindow = 10 * time.Minute

var suspiciousPaths = []string{
	"/.env",
	"/wp-admin",
	"/phpmyadmin",
	"/admin.php",
	"/config",
	"/actuator",
}

var suspiciousNotifyLimiter = NewLimiter(1, suspiciousNotifyWindow)

func SuspiciousScanBlocker(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isSuspiciousPath(r.URL.Path) {
			ip := ClientIP(r)
			if allowed, _ := suspiciousNotifyLimiter.Allow(ip); allowed {
				go func() {
					_ = notify.SendEvent(context.Background(), notify.Event{
						Type:      "suspicious_scan",
						Path:      r.URL.Path,
						Method:    r.Method,
						Status:    http.StatusNotFound,
						IP:        ip,
						UserAgent: r.UserAgent(),
						UserID:    "",
						Meta:      map[string]string{"path": r.URL.Path},
						TS:        time.Now().UTC(),
					})
				}()
			}

			writeJSONError(w, http.StatusNotFound, "not_found")
			return
		}

		next.ServeHTTP(w, r.WithContext(r.Context()))
	})
}

func isSuspiciousPath(path string) bool {
	p := strings.ToLower(path)
	for _, bad := range suspiciousPaths {
		if p == bad || strings.HasPrefix(p, bad+"/") {
			return true
		}
	}
	return false
}
