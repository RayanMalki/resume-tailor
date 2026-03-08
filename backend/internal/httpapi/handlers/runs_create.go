package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"resume-tailor/internal/auth"
	"resume-tailor/internal/httpapi/middleware"
	"resume-tailor/internal/notify"
	"resume-tailor/internal/resumes"
	"resume-tailor/internal/runs"

	"github.com/google/uuid"
)

type CreateRunRequest struct {
	ResumeID           string                `json:"resumeId"`
	JobText            string                `json:"jobText"`
	ProjectControls    []runs.ProjectControl `json:"projectControls"`
	DisciplineOverride string                `json:"disciplineOverride"`
}

type CreateRunResponse struct {
	RunID string `json:"runId"`
}

const (
	freeTierDailyUserLimit = 3
	freeTierDailyIPLimit   = 6
)

func CreateRunHandler(runsSvc *runs.Service, resumesSvc *resumes.Service, authSvc *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		if allowed, res := middleware.AllowRunCreate(userID); !allowed {
			go func() {
				_ = notify.SendEvent(context.Background(), notify.Event{
					Type:      "rate_limited",
					Path:      r.URL.Path,
					Method:    r.Method,
					Status:    http.StatusTooManyRequests,
					IP:        middleware.ClientIP(r),
					UserAgent: r.UserAgent(),
					UserID:    userID.String(),
					Meta: map[string]string{
						"limit":     fmt.Sprintf("%d", res.Limit),
						"remaining": fmt.Sprintf("%d", res.Remain),
						"window":    res.Window.String(),
						"scope":     "user_runs",
					},
					TS: time.Now().UTC(),
				})
			}()
			writeError(w, http.StatusTooManyRequests, "rate_limited")
			return
		}

		// DB-backed daily limit for users without a personal API key.
		user, err := authSvc.GetUserByID(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		creatorIP := ""
		if user.EncryptedOpenAIKey == nil {
			ip := middleware.ClientIP(r)

			userCount, err := runsSvc.CountRunsTodayForUser(r.Context(), userID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "internal server error")
				return
			}
			if userCount >= freeTierDailyUserLimit {
				writeError(w, http.StatusTooManyRequests, "daily_limit_reached")
				return
			}

			ipCount, err := runsSvc.CountRunsTodayForIP(r.Context(), ip)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "internal server error")
				return
			}
			if ipCount >= freeTierDailyIPLimit {
				writeError(w, http.StatusTooManyRequests, "ip_limit_reached")
				return
			}

			creatorIP = ip
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

		var disciplineOverride *runs.Discipline
		if req.DisciplineOverride != "" {
			parsed, ok := runs.ParseDiscipline(req.DisciplineOverride)
			if !ok {
				writeError(w, http.StatusBadRequest, "invalid_discipline_override")
				return
			}
			disciplineOverride = &parsed
		}

		run, err := runsSvc.CreateRun(r.Context(), userID, resumeID, req.JobText, req.ProjectControls, disciplineOverride, creatorIP)
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

		go func() {
			_ = notify.SendEvent(context.Background(), notify.Event{
				Type:      "run_created",
				Path:      r.URL.Path,
				Method:    r.Method,
				Status:    http.StatusCreated,
				IP:        middleware.ClientIP(r),
				UserAgent: r.UserAgent(),
				UserID:    userID.String(),
				Meta: map[string]string{
					"run_id": run.ID.String(),
				},
				TS: time.Now().UTC(),
			})
		}()
	}
}
