package runreports

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetRunReportByRunID(ctx context.Context, runID uuid.UUID) (RunReport, error) {
	if runID == uuid.Nil {
		return RunReport{}, fmt.Errorf("bad input: run_id")
	}

	return s.repo.GetRunReportByRunID(ctx, runID)
}

func (s *Service) UpsertRunReport(ctx context.Context, runID uuid.UUID, atsReport, changePlan json.RawMessage) error {
	if runID == uuid.Nil {
		return fmt.Errorf("bad input: run_id")
	}

	return s.repo.UpsertRunReport(ctx, runID, atsReport, changePlan)
}

