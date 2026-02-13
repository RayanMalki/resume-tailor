package handlers

import (
	"errors"
	"net/http"
	"resume-tailor/internal/httpapi/middleware"
	"resume-tailor/internal/runs"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func GetRunByIdHandler(runsSvc *runs.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		raw := chi.URLParam(r, "runID")
		runID, err := uuid.Parse(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_run_id")
			return
		}

		run, err := runsSvc.GetRunByID(r.Context(), userID, runID)
		if errors.Is(err, runs.ErrRunNotFound) {
			writeError(w, http.StatusNotFound, "run_not_found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_server_error")
			return
		}

		writeJSON(w, http.StatusOK, run)

	}

}
