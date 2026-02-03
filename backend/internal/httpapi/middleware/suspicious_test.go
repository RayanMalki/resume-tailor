package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSuspiciousPathBlocked(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/.env", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	rec := httptest.NewRecorder()

	handler := SuspiciousScanBlocker(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, rec.Code)
	}
}
