package runs

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"resume-tailor/internal/jobs"

	"github.com/google/uuid"
)

type Service struct {
	repo      *Repo
	jobsEnq   jobs.JobsEnqueuer
}

func NewService(repo *Repo, jobsEnq jobs.JobsEnqueuer) *Service {
	return &Service{
		repo:    repo,
		jobsEnq: jobsEnq,
	}
}

func (s *Service) CreateRun(ctx context.Context, userID,
	resumeID uuid.UUID, jobText string) (Run, error) {

	if userID == uuid.Nil {
		return Run{}, fmt.Errorf("bad input: user_id")

	}
	if resumeID == uuid.Nil {
		return Run{}, fmt.Errorf("bad input: resume_id")

	}
	if strings.TrimSpace(jobText) == "" {
		return Run{}, fmt.Errorf("bad input: job_text")

	}

	jobText = strings.TrimSpace(jobText)

	run, err := s.repo.CreateRun(ctx, userID, resumeID, jobText)
	if err != nil {
		return Run{}, err
	}

	// Enqueue job for processing
	if s.jobsEnq != nil {
		_, err = s.jobsEnq.EnqueueProcessRun(ctx, run.ID)
		if err != nil {
			return Run{}, fmt.Errorf("failed to enqueue job: %w", err)
		}
	}

	return run, nil

}

func (s *Service) GetRunByID(ctx context.Context, userID, runID uuid.UUID) (Run, error) {
	if userID == uuid.Nil {
		return Run{}, fmt.Errorf("bad input: user_id")

	}
	if runID == uuid.Nil {
		return Run{}, fmt.Errorf("bad input: run_id")

	}

	run, err := s.repo.GetRunByID(ctx, runID)
	if errors.Is(err, ErrRunNotFound) {
		return Run{}, ErrRunNotFound

	}
	if err != nil {
		return Run{}, err

	}
	if run.UserID != userID {
		return Run{}, ErrRunNotFound

	}

	return run, nil
}

func (s *Service) ListRunsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Run, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("bad input: user_id")
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListRunsByUser(ctx, userID, limit, offset)
}
