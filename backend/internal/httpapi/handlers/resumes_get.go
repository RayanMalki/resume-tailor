package handlers

import (
	"errors"
	"net/http"

	"resume-tailor/internal/httpapi/middleware"
	"resume-tailor/internal/resumes"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func GetResumeByIDHandler(resumesSvc *resumes.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		raw := chi.URLParam(r, "resumeID")
		resumeID, err := uuid.Parse(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid resume ID")
			return
		}

		resume, err := resumesSvc.GetResumeByID(r.Context(), userID, resumeID)
		if errors.Is(err, resumes.ErrResumeNotFound) {
			writeError(w, http.StatusNotFound, "resume not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		writeJSON(w, http.StatusOK, resume)
	}
}

