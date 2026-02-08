package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"resume-tailor/internal/artifacts"
	"resume-tailor/internal/httpapi/middleware"
	"resume-tailor/internal/runs"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type projectReasonsResponse struct {
	Projects []projectReasonItem `json:"projects"`
}

type projectReasonItem struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

func GetProjectReasonsArtifactHandler(runsSvc *runs.Service, artifactsSvc *artifacts.Service) http.HandlerFunc {
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

		artifact, err := artifactsSvc.GetByRunIDAndType(r.Context(), runID, artifacts.TypeProjectReasons)
		if err != nil {
			if errors.Is(err, artifacts.ErrArtifactNotFound) {
				writeError(w, http.StatusNotFound, "not ready")
				return
			}
			slog.Error("failed to load project reasons artifact", "run_id", runID.String(), "error", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		var projects []projectReasonItem
		if err := json.Unmarshal([]byte(artifact.Content), &projects); err != nil {
			slog.Error("failed to decode project reasons artifact", "run_id", runID.String(), "error", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		writeJSON(w, http.StatusOK, projectReasonsResponse{Projects: projects})
	}
}
