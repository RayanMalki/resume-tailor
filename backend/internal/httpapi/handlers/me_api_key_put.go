package handlers

import (
	"encoding/json"
	"net/http"

	"resume-tailor/internal/auth"
	"resume-tailor/internal/httpapi/middleware"
)

type setAPIKeyRequest struct {
	APIKey string `json:"apiKey"`
}

func MeAPIKeyPut(authSvc *auth.Service, encSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if encSecret == "" {
			writeError(w, http.StatusServiceUnavailable, "api_key_feature_disabled")
			return
		}

		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req setAPIKeyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.APIKey == "" {
			writeError(w, http.StatusBadRequest, "apiKey is required")
			return
		}

		if err := authSvc.SetAPIKey(r.Context(), userID, req.APIKey, encSecret); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to store api key")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
