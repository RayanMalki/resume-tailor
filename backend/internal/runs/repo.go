package runs

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
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

func (r *Repo) CreateRun(ctx context.Context, userID, resumeID uuid.UUID, jobText string, projectControls []ProjectControl, discipline *Discipline, disciplineScore float64, disciplineSource DisciplineSource, creatorIP string) (Run, error) {
	controlsJSON, err := marshalProjectControls(projectControls)
	if err != nil {
		return Run{}, err
	}
	if math.IsNaN(disciplineScore) || math.IsInf(disciplineScore, 0) {
		disciplineScore = 0
	}
	var disciplineRaw any
	if discipline != nil {
		disciplineRaw = string(*discipline)
	}
	if strings.TrimSpace(string(disciplineSource)) == "" {
		disciplineSource = DisciplineSourceAuto
	}
	var creatorIPArg any
	if creatorIP != "" {
		creatorIPArg = creatorIP
	}

	const q = `
INSERT INTO runs (user_id, resume_id, job_text, project_controls, discipline, discipline_confidence, discipline_source, status, creator_ip)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, user_id, resume_id, job_text, project_controls, discipline, discipline_confidence, discipline_source, status, error_message, created_at, updated_at
`

	var run Run
	var controlsRaw []byte
	var dbDiscipline sql.NullString
	err = r.db.QueryRow(ctx, q, userID, resumeID, jobText, controlsJSON, disciplineRaw, disciplineScore, disciplineSource, StatusQueued, creatorIPArg).Scan(
		&run.ID,
		&run.UserID,
		&run.ResumeID,
		&run.JobText,
		&controlsRaw,
		&dbDiscipline,
		&run.DisciplineScore,
		&run.DisciplineSource,
		&run.Status,
		&run.ErrorMessage,
		&run.CreatedAt,
		&run.UpdatedAt,
	)
	if err != nil {
		return Run{}, err
	}
	run.ProjectControls = unmarshalProjectControls(controlsRaw)
	if dbDiscipline.Valid {
		if parsed, ok := ParseDiscipline(dbDiscipline.String); ok {
			run.Discipline = &parsed
		}
	}

	return run, nil
}

func (r *Repo) GetRunByID(ctx context.Context, runID uuid.UUID) (Run, error) {
	if runID == uuid.Nil {
		return Run{}, fmt.Errorf("bad input: run_id")

	}
	const q = `
		SELECT id, user_id, resume_id, job_text, project_controls, discipline, discipline_confidence, discipline_source, status, error_message, created_at, updated_at 
		FROM runs where id = $1`

	var run Run
	var controlsRaw []byte
	var dbDiscipline sql.NullString
	err := r.db.QueryRow(ctx, q, runID).Scan(
		&run.ID,
		&run.UserID,
		&run.ResumeID,
		&run.JobText,
		&controlsRaw,
		&dbDiscipline,
		&run.DisciplineScore,
		&run.DisciplineSource,
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
	run.ProjectControls = unmarshalProjectControls(controlsRaw)
	if dbDiscipline.Valid {
		if parsed, ok := ParseDiscipline(dbDiscipline.String); ok {
			run.Discipline = &parsed
		}
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
SELECT id, user_id, resume_id, job_text, project_controls, discipline, discipline_confidence, discipline_source, status, error_message, created_at, updated_at
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
		var controlsRaw []byte
		var dbDiscipline sql.NullString
		if err := rows.Scan(
			&run.ID,
			&run.UserID,
			&run.ResumeID,
			&run.JobText,
			&controlsRaw,
			&dbDiscipline,
			&run.DisciplineScore,
			&run.DisciplineSource,
			&run.Status,
			&run.ErrorMessage,
			&run.CreatedAt,
			&run.UpdatedAt,
		); err != nil {
			return nil, err
		}
		run.ProjectControls = unmarshalProjectControls(controlsRaw)
		if dbDiscipline.Valid {
			if parsed, ok := ParseDiscipline(dbDiscipline.String); ok {
				run.Discipline = &parsed
			}
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

func (r *Repo) UpdateRunDiscipline(ctx context.Context, runID uuid.UUID, discipline Discipline, confidence float64, source DisciplineSource) error {
	if runID == uuid.Nil {
		return fmt.Errorf("bad input: run_id")
	}
	if _, ok := ParseDiscipline(string(discipline)); !ok {
		return fmt.Errorf("bad input: discipline")
	}
	if source != DisciplineSourceAuto && source != DisciplineSourceUserOverride {
		return fmt.Errorf("bad input: discipline_source")
	}
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	const q = `
		UPDATE runs
		SET discipline = $2,
		    discipline_confidence = $3,
		    discipline_source = $4,
		    updated_at = now()
		WHERE id = $1
	`
	cmdTag, err := r.db.Exec(ctx, q, runID, string(discipline), confidence, string(source))
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrRunNotFound
	}
	return nil
}

// CountRunsTodayForUser counts runs created today by the given user.
func (r *Repo) CountRunsTodayForUser(ctx context.Context, userID uuid.UUID) (int, error) {
	const q = `SELECT COUNT(*) FROM runs WHERE user_id = $1 AND created_at >= CURRENT_DATE`
	var count int
	err := r.db.QueryRow(ctx, q, userID).Scan(&count)
	return count, err
}

// CountRunsTodayForIP counts runs created today from the given IP address.
func (r *Repo) CountRunsTodayForIP(ctx context.Context, ip string) (int, error) {
	const q = `SELECT COUNT(*) FROM runs WHERE creator_ip = $1 AND created_at >= CURRENT_DATE`
	var count int
	err := r.db.QueryRow(ctx, q, ip).Scan(&count)
	return count, err
}

func marshalProjectControls(controls []ProjectControl) ([]byte, error) {
	if len(controls) == 0 {
		return []byte("[]"), nil
	}
	b, err := json.Marshal(controls)
	if err != nil {
		return nil, fmt.Errorf("marshal project_controls: %w", err)
	}
	return b, nil
}

func unmarshalProjectControls(raw []byte) []ProjectControl {
	if len(raw) == 0 {
		return nil
	}
	var controls []ProjectControl
	if err := json.Unmarshal(raw, &controls); err != nil {
		return nil
	}
	return controls
}
