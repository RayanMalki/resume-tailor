package runs

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusCreated    Status = "created"
	StatusQueued     Status = "queued"
	StatusProcessing Status = "processing"
	StatusFailed     Status = "Failed"
)

type Run struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	ResumeID     uuid.UUID
	JobText      string
	Status       Status
	ErrorMessage *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

var (
	ErrRunNotFound = errors.New("run failed")
	ErrForbidden   = errors.New("forbidden")
	ErrBadInput    = errors.New("bad input")
)
