package handlers

import (
	"net/http"
	"strconv"

	"resume-tailor/internal/httpapi/middleware"
	"resume-tailor/internal/runs"
)

func ListRunsHandler(runsSvc *runs.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Defaults
		limit := 20
		offset := 0

		// Parse ?limit=
		if raw := r.URL.Query().Get("limit"); raw != "" {
			v, err := strconv.Atoi(raw)
			if err != nil || v <= 0 {
				http.Error(w, "Bad Request: invalid limit", http.StatusBadRequest)
				return
			}
			if v > 100 {
				v = 100
			}
			limit = v
		}

		// Parse ?offset=
		if raw := r.URL.Query().Get("offset"); raw != "" {
			v, err := strconv.Atoi(raw)
			if err != nil || v < 0 {
				http.Error(w, "Bad Request: invalid offset", http.StatusBadRequest)
				return
			}
			offset = v
		}

		list, err := runsSvc.ListRunsByUser(r.Context(), userID, limit, offset)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, list)
	}
}
