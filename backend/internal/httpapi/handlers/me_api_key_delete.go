package handlers

import (
	"net/http"

	"resume-tailor/internal/auth"
	"resume-tailor/internal/httpapi/middleware"
)

func MeAPIKeyDelete(authSvc *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		if err := authSvc.RemoveAPIKey(r.Context(), userID); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to remove api key")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
