package runreports

import (
	"context"
	"encoding/json"
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

func (r *Repo) UpsertRunReport(ctx context.Context, runID uuid.UUID, atsReport, changePlan json.RawMessage) error {
	if runID == uuid.Nil {
		return fmt.Errorf("bad input: run_id")
	}

	const q = `
INSERT INTO run_reports (run_id, ats_report, change_plan)
VALUES ($1, $2, $3)
ON CONFLICT (run_id) DO UPDATE
SET ats_report = $2, change_plan = $3, created_at = now()`

	_, err := r.db.Exec(ctx, q, runID, atsReport, changePlan)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) GetRunReportByRunID(ctx context.Context, runID uuid.UUID) (RunReport, error) {
	if runID == uuid.Nil {
		return RunReport{}, fmt.Errorf("bad input: run_id")
	}

	const q = `
SELECT run_id, ats_report, change_plan, created_at
FROM run_reports
WHERE run_id = $1`

	var report RunReport
	err := r.db.QueryRow(ctx, q, runID).Scan(
		&report.RunID,
		&report.ATSReport,
		&report.ChangePlan,
		&report.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RunReport{}, ErrRunReportNotFound
		}
		return RunReport{}, err
	}

	return report, nil
}

