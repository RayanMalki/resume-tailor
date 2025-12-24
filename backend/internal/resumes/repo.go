package resumes

import (
	"context"
	"errors"

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

func (r *Repo) CreateResume(ctx context.Context, userID uuid.UUID, title string, contentText string) (Resume, error) {
	const q = `
INSERT INTO resumes (user_id, title, content_text)
VALUES ($1, $2, $3)
RETURNING id, user_id, title, content_text, created_at, updated_at`

	var res Resume
	err := r.db.QueryRow(ctx, q, userID, title, contentText).Scan(
		&res.ID,
		&res.UserID,
		&res.Title,
		&res.ContentText,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	if err != nil {
		return Resume{}, err
	}

	return res, nil
}

func (r *Repo) GetResumeByID(ctx context.Context, resumeID uuid.UUID) (Resume, error) {
	const q = `
SELECT id, user_id, title, content_text, created_at, updated_at
FROM resumes
WHERE id = $1`

	var res Resume
	err := r.db.QueryRow(ctx, q, resumeID).Scan(
		&res.ID,
		&res.UserID,
		&res.Title,
		&res.ContentText,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Resume{}, ErrResumeNotFound
		}
		return Resume{}, err
	}

	return res, nil
}

func (r *Repo) ListResumesByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Resume, error) {
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
SELECT id, user_id, title, content_text, created_at, updated_at
FROM resumes
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3`

	rows, err := r.db.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resumes []Resume
	for rows.Next() {
		var res Resume
		err := rows.Scan(
			&res.ID,
			&res.UserID,
			&res.Title,
			&res.ContentText,
			&res.CreatedAt,
			&res.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		resumes = append(resumes, res)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return resumes, nil
}
