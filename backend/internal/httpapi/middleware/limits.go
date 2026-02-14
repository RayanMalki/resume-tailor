package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
)

type bodyLimitRule struct {
	Method string
	Path   string
	Limit  int64
}

const defaultBodyLimit = int64(1 << 20) // 1MB

var bodyLimitRules = []bodyLimitRule{
	{Method: http.MethodPost, Path: "/v1/resumes", Limit: 2 << 20},
	{Method: http.MethodPost, Path: "/v1/runs", Limit: 1 << 20},
	{Method: http.MethodPost, Path: "/v1/disciplines/detect", Limit: 1 << 20},
	{Method: http.MethodPost, Path: "/v1/auth/login", Limit: 64 << 10},
	{Method: http.MethodPost, Path: "/v1/auth/signup", Limit: 64 << 10},
}

func Limits(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil || r.Method == http.MethodGet || r.Method == http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}

		limit := matchBodyLimit(r.Method, r.URL.Path)
		if r.ContentLength > 0 && r.ContentLength > limit {
			writeJSONError(w, http.StatusRequestEntityTooLarge, "payload_too_large")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, limit)
		next.ServeHTTP(w, r)
	})
}

func matchBodyLimit(method, path string) int64 {
	for _, rule := range bodyLimitRules {
		if rule.Method == method && strings.HasPrefix(path, rule.Path) {
			return rule.Limit
		}
	}
	return defaultBodyLimit
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
