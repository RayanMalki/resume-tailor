package handlers

import (
	"net/http"

	"resume-tailor/internal/auth"
	"resume-tailor/internal/httpapi/middleware"
)

func MeOnboardingSeen(authSvc *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		if err := authSvc.MarkOnboardingSeen(r.Context(), userID); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update onboarding status")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
