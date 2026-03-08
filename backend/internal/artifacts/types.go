package artifacts

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Artifact struct {
	RunID     uuid.UUID
	Type      string
	Content   string
	CreatedAt time.Time
}

const (
	TypeResumeLatex    = "resume_latex"
	TypeResumePDF      = "resume_pdf"
	TypeResumeDOCX     = "resume_docx"
	TypeCoverLetter    = "cover_letter"
	TypeCoverLetterPDF = "cover_letter_pdf"
	TypeCoverLetterDOCX = "cover_letter_docx"
	TypeProjectReasons = "project_reasons"
)

var (
	ErrArtifactNotFound = errors.New("artifact not found")
)
