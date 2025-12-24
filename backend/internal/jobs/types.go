package jobs

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID         uuid.UUID
	Type       string
	RunID      uuid.UUID
	Status     string
	Attempts   int
	MaxAttempts int
	LockedBy   *string
	LockedAt   *time.Time
	LastError  *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

const (
	JobTypeProcessRun = "process_run"
)

const (
	JobStatusQueued  = "queued"
	JobStatusRunning = "running"
	JobStatusFailed  = "failed"
	JobStatusDone    = "done"
)

var (
	ErrJobNotFound = errors.New("job not found")
)

