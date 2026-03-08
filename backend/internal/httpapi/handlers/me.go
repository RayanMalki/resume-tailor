package handlers

import (
	"net/http"

	"resume-tailor/internal/auth"
	"resume-tailor/internal/httpapi/middleware"
)

type meResponse struct {
	UserID       string  `json:"userId"`
	Email        string  `json:"email"`
	DisplayName  string  `json:"displayName"`
	AvatarURL    *string `json:"avatarUrl"`
	AuthProvider string  `json:"authProvider"`
	HasApiKey    bool    `json:"hasApiKey"`
}

func Me(authSvc *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		user, err := authSvc.GetUserByID(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load user")
			return
		}

		writeJSON(w, http.StatusOK, meResponse{
			UserID:       user.ID.String(),
			Email:        user.Email,
			DisplayName:  user.DisplayName,
			AvatarURL:    user.AvatarURL,
			AuthProvider: user.AuthProvider,
			HasApiKey:    user.EncryptedOpenAIKey != nil,
		})
	}
}
