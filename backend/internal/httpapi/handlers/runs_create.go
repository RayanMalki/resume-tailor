package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"resume-tailor/internal/httpapi/middleware"
	"resume-tailor/internal/resumes"
	"resume-tailor/internal/runs"

	"github.com/google/uuid"
)

type CreateRunRequest struct {
	ResumeID string `json:"resumeId"`
	JobText  string `json:"jobText"`
}

type CreateRunResponse struct {
	RunID string `json:"runId"`
}

func CreateRunHandler(runsSvc *runs.Service, resumesSvc *resumes.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req CreateRunRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request payload")
			return
		}

		resumeID, err := uuid.Parse(req.ResumeID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid resumeId")
			return
		}

		// Ensure resume belongs to the current user (no ID leaking)
		_, err = resumesSvc.GetResumeByID(r.Context(), userID, resumeID)
		if err != nil {
			if errors.Is(err, resumes.ErrResumeNotFound) {
				writeError(w, http.StatusNotFound, "not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		run, err := runsSvc.CreateRun(r.Context(), userID, resumeID, req.JobText)
		if err != nil {
			if errors.Is(err, runs.ErrBadInput) {
				// Return the detailed validation message (ex: "bad input: job_text")
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			if errors.Is(err, runs.ErrRunNotFound) {
				writeError(w, http.StatusNotFound, "not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		writeJSON(w, http.StatusCreated, CreateRunResponse{
			RunID: run.ID.String(),
		})
	}
}
