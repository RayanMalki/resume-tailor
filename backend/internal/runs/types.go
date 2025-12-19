package runs

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	SatusCreated     Status = "created"
	SatusQueued      Status = "queued"
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
)
