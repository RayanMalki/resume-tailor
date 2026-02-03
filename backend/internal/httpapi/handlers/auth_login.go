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
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		token, expiresAt, userID, err := authSvc.LoginWithUser(r.Context(), req.Email, req.Password)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidCredentials) {
				http.Error(w, "Unauthorized: invalid credentials", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
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
