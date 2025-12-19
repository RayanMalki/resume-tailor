package runs

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

func (r *Repo) CreateRun(ctx context.Context, userID, resumeID uuid.UUID, jobText string) (Run, error) {
	const q = `
INSERT INTO runs (user_id, resume_id, job_text, status)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, resume_id, job_text, status, error_message, created_at, updated_at
`

	var run Run
	err := r.db.QueryRow(ctx, q, userID, resumeID, jobText, StatusCreated).Scan(
		&run.ID,
		&run.UserID,
		&run.ResumeID,
		&run.JobText,
		&run.Status,
		&run.ErrorMessage,
		&run.CreatedAt,
		&run.UpdatedAt,
	)
	if err != nil {
		return Run{}, err
	}

	return run, nil
}

func (r *Repo) GetRunByID(ctx context.Context, runID uuid.UUID) (Run, error) {
	if runID == uuid.Nil {
		return Run{}, fmt.Errorf("bad input: run_id")

	}
	const q = `
		SELECT id, user_id, resume_id, job_text, status, error_message, created_at, updated_at 
		FROM runs where id = $1`

	var run Run
	err := r.db.QueryRow(ctx, q, runID).Scan(
		&run.ID,
		&run.UserID,
		&run.ResumeID,
		&run.JobText,
		&run.Status,
		&run.ErrorMessage,
		&run.CreatedAt,
		&run.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Run{}, ErrRunNotFound
		}
		return Run{}, err
	}
	return run, nil

}

func (r *Repo) ListRunsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Run, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("bad input: user_id")
	}

	// defaults / safety caps
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	const q = `
SELECT id, user_id, resume_id, job_text, status, error_message, created_at, updated_at
FROM runs
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	runs := make([]Run, 0, limit)
	for rows.Next() {
		var run Run
		if err := rows.Scan(
			&run.ID,
			&run.UserID,
			&run.ResumeID,
			&run.JobText,
			&run.Status,
			&run.ErrorMessage,
			&run.CreatedAt,
			&run.UpdatedAt,
		); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return runs, nil
}

func (r *Repo) UpdateRunStatus(ctx context.Context, runID uuid.UUID, status string, errorMessage *string) error {
	if runID == uuid.Nil {
		return fmt.Errorf("bad input: run_id")
	}
	if strings.TrimSpace(status) == "" {
		return fmt.Errorf("bad input: status")
	}

	const q = `
		UPDATE runs
		SET status = $2,
		    error_message = $3,
		    updated_at = now()
		WHERE id = $1
	`

	cmdTag, err := r.db.Exec(ctx, q, runID, status, errorMessage)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrRunNotFound
	}

	return nil
}
