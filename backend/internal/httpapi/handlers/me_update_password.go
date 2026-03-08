package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"resume-tailor/internal/auth"
	"resume-tailor/internal/httpapi/middleware"
)

type changePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func MeUpdatePassword(authSvc *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req changePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.CurrentPassword == "" || req.NewPassword == "" {
			writeError(w, http.StatusBadRequest, "currentPassword and newPassword are required")
			return
		}

		if err := authSvc.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
			if errors.Is(err, auth.ErrInvalidCredentials) {
				writeError(w, http.StatusUnauthorized, "current password is incorrect")
				return
			}
			if errors.Is(err, auth.ErrWeakPassword) {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
