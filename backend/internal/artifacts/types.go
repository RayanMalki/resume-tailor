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
	TypeResumeLatex = "resume_latex"
	TypeResumePDF   = "resume_pdf"
)

var (
	ErrArtifactNotFound = errors.New("artifact not found")
)
