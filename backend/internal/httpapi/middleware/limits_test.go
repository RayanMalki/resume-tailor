package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBodyLimitExceeded(t *testing.T) {
	tooLarge := bytes.Repeat([]byte("a"), int((2<<20)+1))
	req := httptest.NewRequest(http.MethodPost, "/v1/resumes", bytes.NewReader(tooLarge))
	req.RemoteAddr = "1.2.3.4:1234"
	rec := httptest.NewRecorder()

	called := false
	handler := Limits(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected %d, got %d", http.StatusRequestEntityTooLarge, rec.Code)
	}
	if called {
		t.Fatalf("handler should not be called on oversized payload")
	}
}
