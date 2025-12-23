package resumes

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Resume struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	ContentText string
	CreatedAt   time.Time
	UpadatedAt  time.Time
}

var (
	ErrResumeNotFound = errors.New("resume not found")
	ErrBadInput       = errors.New("bad input")
)
