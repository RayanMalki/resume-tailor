package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"resume-tailor/internal/auth"
	"resume-tailor/internal/httpapi/cookies"
	"resume-tailor/internal/httpapi/middleware"
	"resume-tailor/internal/notify"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(authSvc *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request_payload")
			return
		}

		token, expiresAt, userID, err := authSvc.LoginWithUser(r.Context(), req.Email, req.Password)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidCredentials) {
				writeError(w, http.StatusUnauthorized, "invalid_credentials")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal_server_error")
			return
		}

		cookies.SetSessionCookie(w, token, expiresAt)
		w.WriteHeader(http.StatusNoContent)

		go func() {
			_ = notify.SendEvent(context.Background(), notify.Event{
				Type:      "login_success",
				Path:      r.URL.Path,
				Method:    r.Method,
				Status:    http.StatusNoContent,
				IP:        middleware.ClientIP(r),
				UserAgent: r.UserAgent(),
				UserID:    userID.String(),
				Meta:      map[string]string{},
				TS:        time.Now().UTC(),
			})
		}()
	}
}
