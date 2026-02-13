package handlers

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"resume-tailor/internal/httpapi/middleware"
	"resume-tailor/internal/resumes"
)

const maxUploadSize = 2 << 20 // 2 MB

// UploadResumeHandler accepts a multipart/form-data file upload (PDF or DOCX),
// extracts the text content, saves it as a resume, and returns the extracted
// text so the frontend can display it in the textarea for review/editing.
func UploadResumeHandler(resumesSvc *resumes.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// Limit request body size
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			writeError(w, http.StatusBadRequest, "file too large (max 2MB)")
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			writeError(w, http.StatusBadRequest, "missing file field")
			return
		}
		defer file.Close()

		// Validate file extension
		ext := strings.ToLower(filepath.Ext(header.Filename))
		if ext != ".pdf" && ext != ".docx" {
			writeError(w, http.StatusBadRequest, "only PDF and DOCX files are accepted")
			return
		}

		// Read file bytes
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read file")
			return
		}

		// Extract text based on file type
		var extractedText string
		switch ext {
		case ".pdf":
			extractedText, err = extractTextFromPDF(fileBytes)
		case ".docx":
			extractedText, err = extractTextFromDOCX(fileBytes)
		}
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, fmt.Sprintf("failed to extract text: %s", err.Error()))
			return
		}

		extractedText = strings.TrimSpace(extractedText)
		if extractedText == "" {
			writeError(w, http.StatusUnprocessableEntity, "no text could be extracted from the file")
			return
		}

		// Derive a title from the filename (without extension)
		title := strings.TrimSuffix(header.Filename, ext)
		if title == "" {
			title = "My Resume"
		}

		// Save the extracted text as a resume
		resume, err := resumesSvc.CreateResume(r.Context(), userID, title, extractedText)
		if err != nil {
			errStr := err.Error()
			if strings.HasPrefix(errStr, "bad input:") {
				writeError(w, http.StatusBadRequest, errStr)
				return
			}
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		writeJSON(w, http.StatusCreated, map[string]string{
			"resumeId":    resume.ID.String(),
			"title":       title,
			"contentText": extractedText,
		})
	}
}
