package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"hello": "world"}

	writeJSON(w, http.StatusOK, data)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", ct)
	}

	var body map[string]string
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if body["hello"] != "world" {
		t.Fatalf("expected hello=world, got %v", body)
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, http.StatusBadRequest, "invalid_input")

	res := w.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if body["error"] != "invalid_input" {
		t.Fatalf("expected error=invalid_input, got %v", body)
	}
}

func TestDecodeJSON(t *testing.T) {
	body := strings.NewReader(`{"email":"test@example.com","password":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/json")

	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := decodeJSON(req, &payload)
	if err != nil {
		t.Fatalf("decodeJSON error: %v", err)
	}
	if payload.Email != "test@example.com" {
		t.Fatalf("expected test@example.com, got %s", payload.Email)
	}
}

func TestDecodeJSONNilBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Body = nil

	var payload struct{}
	err := decodeJSON(req, &payload)
	if err == nil {
		t.Fatal("expected error for nil body")
	}
}

func TestWriteJSONStatusCodes(t *testing.T) {
	codes := []int{200, 201, 400, 401, 403, 404, 500}
	for _, code := range codes {
		w := httptest.NewRecorder()
		writeJSON(w, code, map[string]string{"ok": "true"})
		if w.Code != code {
			t.Errorf("expected status %d, got %d", code, w.Code)
		}
	}
}
