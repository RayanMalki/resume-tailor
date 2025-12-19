package runs

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{
		repo: repo,
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

	return run, nil

}
