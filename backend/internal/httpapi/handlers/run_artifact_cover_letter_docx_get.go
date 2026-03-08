package handlers

import (
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"

	"resume-tailor/internal/artifacts"
	"resume-tailor/internal/httpapi/middleware"
	"resume-tailor/internal/runs"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func GetCoverLetterDOCXArtifactHandler(runsSvc *runs.Service, artifactsSvc *artifacts.Service) http.HandlerFunc {
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

		artifact, err := artifactsSvc.GetByRunIDAndType(r.Context(), runID, artifacts.TypeCoverLetterDOCX)
		if err != nil {
			if errors.Is(err, artifacts.ErrArtifactNotFound) {
				writeError(w, http.StatusNotFound, "not ready")
				return
			}
			slog.Error("failed to load cover letter docx artifact", "run_id", runID.String(), "error", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		docxBytes, err := base64.StdEncoding.DecodeString(artifact.Content)
		if err != nil {
			slog.Error("failed to decode cover letter docx artifact", "run_id", runID.String(), "error", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
		w.Header().Set("Content-Disposition", "attachment; filename=\"cover_letter.docx\"")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(docxBytes)
	}
}
