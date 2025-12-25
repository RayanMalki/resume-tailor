package handlers

import (
	"net/http"
	"strconv"

	"resume-tailor/internal/httpapi/middleware"
	"resume-tailor/internal/resumes"
)

func ListResumesHandler(resumesSvc *resumes.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// Defaults
		limit := 20
		offset := 0

		// Parse ?limit=
		if raw := r.URL.Query().Get("limit"); raw != "" {
			v, err := strconv.Atoi(raw)
			if err != nil || v <= 0 {
				writeError(w, http.StatusBadRequest, "invalid limit")
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
				writeError(w, http.StatusBadRequest, "invalid offset")
				return
			}
			offset = v
		}

		list, err := resumesSvc.ListResumesByUser(r.Context(), userID, limit, offset)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		writeJSON(w, http.StatusOK, list)
	}
}

