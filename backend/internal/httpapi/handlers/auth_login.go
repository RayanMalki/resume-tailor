package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"resume-tailor/internal/auth"
	"resume-tailor/internal/httpapi/cookies"
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

		token, expiresAt, err := authSvc.Login(r.Context(), req.Email, req.Password)
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
	}
}
