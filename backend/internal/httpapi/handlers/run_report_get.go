package handlers

import (
	"errors"
	"net/http"

	"resume-tailor/internal/httpapi/middleware"
	"resume-tailor/internal/runreports"
	"resume-tailor/internal/runs"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func GetRunReportHandler(runsSvc *runs.Service, reportsSvc *runreports.Service) http.HandlerFunc {
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

		// Ownership check: ensure the run belongs to the user
		_, err = runsSvc.GetRunByID(r.Context(), userID, runID)
		if err != nil {
			if errors.Is(err, runs.ErrRunNotFound) {
				writeError(w, http.StatusNotFound, "not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		// Fetch report
		report, err := reportsSvc.GetRunReportByRunID(r.Context(), runID)
		if err != nil {
			if errors.Is(err, runreports.ErrRunReportNotFound) {
				writeError(w, http.StatusNotFound, "report not ready")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		writeJSON(w, http.StatusOK, report)
	}
}

