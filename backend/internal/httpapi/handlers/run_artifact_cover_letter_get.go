package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"resume-tailor/internal/artifacts"
	"resume-tailor/internal/httpapi/middleware"
	"resume-tailor/internal/runs"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type coverLetterArtifactResponse struct {
	CoverLetter string `json:"coverLetter"`
}

func GetCoverLetterArtifactHandler(runsSvc *runs.Service, artifactsSvc *artifacts.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		raw := chi.URLParam(r, "runID")
		runID, err := uuid.Parse(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid runID")
			return
		}

		_, err = runsSvc.GetRunByID(r.Context(), userID, runID)
		if err != nil {
			if errors.Is(err, runs.ErrRunNotFound) {
				writeError(w, http.StatusNotFound, "not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		artifact, err := artifactsSvc.GetByRunIDAndType(r.Context(), runID, artifacts.TypeCoverLetter)
		if err != nil {
			if errors.Is(err, artifacts.ErrArtifactNotFound) {
				writeError(w, http.StatusNotFound, "not ready")
				return
			}
			slog.Error("failed to load cover letter artifact", "run_id", runID.String(), "error", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		writeJSON(w, http.StatusOK, coverLetterArtifactResponse{CoverLetter: artifact.Content})
	}
}
