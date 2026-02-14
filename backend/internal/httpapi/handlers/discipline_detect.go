package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"resume-tailor/internal/httpapi/middleware"
	"resume-tailor/internal/resumes"
	"resume-tailor/internal/scoring/classifier"

	"github.com/google/uuid"
)

type DetectDisciplineRequest struct {
	ResumeID string `json:"resumeId"`
	JobText  string `json:"jobText"`
}

type DetectDisciplineResponse struct {
	Discipline    string                    `json:"discipline"`
	Confidence    float64                   `json:"confidence"`
	LowConfidence bool                      `json:"lowConfidence"`
	Evidence      []classifier.EvidenceTerm `json:"evidence"`
	Scores        map[string]float64        `json:"scores,omitempty"`
	Source        string                    `json:"source"`
}

func DetectDisciplineHandler(resumesSvc *resumes.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req DetectDisciplineRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request_payload")
			return
		}
		if strings.TrimSpace(req.ResumeID) == "" {
			writeError(w, http.StatusBadRequest, "resume_id_required")
			return
		}
		if strings.TrimSpace(req.JobText) == "" {
			writeError(w, http.StatusBadRequest, "job_text_required")
			return
		}

		resumeID, err := uuid.Parse(req.ResumeID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_resume_id")
			return
		}

		resume, err := resumesSvc.GetResumeByID(r.Context(), userID, resumeID)
		if err != nil {
			writeError(w, http.StatusNotFound, "resume_not_found")
			return
		}

		detected := classifier.Detect(resume.ContentText, req.JobText)
		writeJSON(w, http.StatusOK, DetectDisciplineResponse{
			Discipline:    string(detected.Discipline),
			Confidence:    detected.Confidence,
			LowConfidence: detected.LowConfidence,
			Evidence:      detected.Evidence,
			Scores:        detected.Scores,
			Source:        string(detected.Source),
		})
	}
}
