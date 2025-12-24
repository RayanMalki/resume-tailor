package jobs

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

