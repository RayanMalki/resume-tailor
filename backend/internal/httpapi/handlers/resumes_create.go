package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"resume-tailor/internal/httpapi/middleware"
	"resume-tailor/internal/resumes"
)

type CreateResumeRequest struct {
	Title       string `json:"title"`
	ContentText string `json:"contentText"`
}

type CreateResumeResponse struct {
	ResumeID string `json:"resumeId"`
}

func CreateResumeHandler(resumesSvc *resumes.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Get user ID from context
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// 2. Decode JSON body
		var req CreateResumeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request payload")
			return
		}

		// 3. Create resume
		resume, err := resumesSvc.CreateResume(r.Context(), userID, req.Title, req.ContentText)
		if err != nil {
			// Check if it's a validation error (starts with "bad input:")
			errStr := err.Error()
			if strings.HasPrefix(errStr, "bad input:") {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			// Other errors
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		// 4. Success response
		resp := CreateResumeResponse{
			ResumeID: resume.ID.String(),
		}
		writeJSON(w, http.StatusCreated, resp)
	}
}
