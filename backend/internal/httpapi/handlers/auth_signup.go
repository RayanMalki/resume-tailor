package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"resume-tailor/internal/auth"
	"resume-tailor/internal/httpapi/cookies"
)

type SignupRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

type SignupResponse struct {
	UserID string `json:"userId"`
}

func Signup(authSvc *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SignupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		id, err := authSvc.Signup(r.Context(), req.Email, req.Password, req.DisplayName)
		if err != nil {
			if errors.Is(err, auth.ErrEmailTaken) {
				http.Error(w, "Email already in use", http.StatusConflict)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// MVP: auto-login after signup (sets HttpOnly cookie)
		token, expiresAt, err := authSvc.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		cookies.SetSessionCookie(w, token, expiresAt)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(SignupResponse{UserID: id.String()})
	}
}
