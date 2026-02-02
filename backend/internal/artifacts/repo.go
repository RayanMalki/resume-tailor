package artifacts

import (
	"context"
	"errors"
	"fmt"

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

func (r *Repo) InsertIfNotExists(ctx context.Context, runID uuid.UUID, artifactType string, content string) error {
	if runID == uuid.Nil {
		return fmt.Errorf("bad input: run_id")
	}
	if artifactType == "" {
		return fmt.Errorf("bad input: type")
	}

	const q = `
INSERT INTO run_artifacts_items (run_id, type, content)
VALUES ($1, $2, $3)
ON CONFLICT (run_id, type) DO NOTHING`

	_, err := r.db.Exec(ctx, q, runID, artifactType, content)
	return err
}

func (r *Repo) GetByRunIDAndType(ctx context.Context, runID uuid.UUID, artifactType string) (Artifact, error) {
	if runID == uuid.Nil {
		return Artifact{}, fmt.Errorf("bad input: run_id")
	}
	if artifactType == "" {
		return Artifact{}, fmt.Errorf("bad input: type")
	}

	const q = `
SELECT run_id, type, content, created_at
FROM run_artifacts_items
WHERE run_id = $1 AND type = $2`

	var art Artifact
	err := r.db.QueryRow(ctx, q, runID, artifactType).Scan(
		&art.RunID,
		&art.Type,
		&art.Content,
		&art.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Artifact{}, ErrArtifactNotFound
		}
		return Artifact{}, err
	}

	return art, nil
}
