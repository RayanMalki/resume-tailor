package handlers

import (
	"net/http"

	"resume-tailor/internal/auth"
	"resume-tailor/internal/httpapi/middleware"
)

func MeDelete(authSvc *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		if err := authSvc.DeleteAccount(r.Context(), userID); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to delete account")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
