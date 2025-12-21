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
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return

		}
		raw := chi.URLParam(r, "runID")
		runID, err := uuid.Parse(raw)

		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		run, err := runsSvc.GetRunByID(r.Context(), userID, runID)
		if errors.Is(err, runs.ErrRunNotFound) {
			http.Error(w, "Not Found", http.StatusNotFound)
			return

		}
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, run)

	}

}
