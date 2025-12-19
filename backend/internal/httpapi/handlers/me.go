package handlers

import (
	"net/http"

	"resume-tailor/internal/httpapi/middleware"
)

type meResponse struct {
	UserID string `json:"userId"`
}

func Me() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		writeJSON(w, http.StatusOK, meResponse{
			UserID: userID.String(),
		})
	}
}
