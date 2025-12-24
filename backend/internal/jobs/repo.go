package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// JobsEnqueuer is an interface for enqueueing jobs.
// This allows runs.Service to depend on jobs without creating an import cycle.
type JobsEnqueuer interface {
	EnqueueProcessRun(ctx context.Context, runID uuid.UUID) (uuid.UUID, error)
}

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

func (r *Repo) EnqueueProcessRun(ctx context.Context, runID uuid.UUID) (uuid.UUID, error) {
	if runID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("runID cannot be nil")
	}

	const q = `
INSERT INTO jobs (type, run_id, status)
VALUES ($1, $2, $3)
RETURNING id`

	var id uuid.UUID
	err := r.db.QueryRow(ctx, q, JobTypeProcessRun, runID, JobStatusQueued).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (r *Repo) ClaimNextProcessRun(ctx context.Context, workerID string) (Job, error) {
	const q = `
SELECT id, type, run_id, status, attempts, max_attempts, locked_by, locked_at, last_error, created_at, updated_at
FROM jobs
WHERE type = $1 AND status = $2
ORDER BY created_at ASC
FOR UPDATE SKIP LOCKED
LIMIT 1`

	var job Job
	err := r.db.QueryRow(ctx, q, JobTypeProcessRun, JobStatusQueued).Scan(
		&job.ID,
		&job.Type,
		&job.RunID,
		&job.Status,
		&job.Attempts,
		&job.MaxAttempts,
		&job.LockedBy,
		&job.LockedAt,
		&job.LastError,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Job{}, ErrNoJobs
		}
		return Job{}, err
	}

	// Update job to running status
	now := time.Now()
	const updateQ = `
UPDATE jobs
SET status = $1,
    locked_by = $2,
    locked_at = $3,
    attempts = attempts + 1,
    updated_at = $4
WHERE id = $5`

	_, err = r.db.Exec(ctx, updateQ, JobStatusRunning, workerID, now, now, job.ID)
	if err != nil {
		return Job{}, err
	}

	job.Status = JobStatusRunning
	job.LockedBy = &workerID
	job.LockedAt = &now
	job.Attempts++
	job.UpdatedAt = now

	return job, nil
}

func (r *Repo) MarkJobDone(ctx context.Context, jobID uuid.UUID) error {
	const q = `
UPDATE jobs
SET status = $1,
    updated_at = now()
WHERE id = $2`

	_, err := r.db.Exec(ctx, q, JobStatusDone, jobID)
	return err
}

func (r *Repo) MarkJobFailed(ctx context.Context, jobID uuid.UUID, errorMsg string, requeue bool) error {
	var status string
	if requeue {
		status = JobStatusQueued
	} else {
		status = JobStatusFailed
	}

	const q = `
UPDATE jobs
SET status = $1,
    last_error = $2,
    updated_at = now()
WHERE id = $3`

	_, err := r.db.Exec(ctx, q, status, errorMsg, jobID)
	return err
}

