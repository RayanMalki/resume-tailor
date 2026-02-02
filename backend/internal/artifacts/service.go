package artifacts

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) InsertIfNotExists(ctx context.Context, runID uuid.UUID, artifactType string, content string) error {
	if runID == uuid.Nil {
		return fmt.Errorf("bad input: run_id")
	}
	if artifactType == "" {
		return fmt.Errorf("bad input: type")
	}
	return s.repo.InsertIfNotExists(ctx, runID, artifactType, content)
}

func (s *Service) GetByRunIDAndType(ctx context.Context, runID uuid.UUID, artifactType string) (Artifact, error) {
	if runID == uuid.Nil {
		return Artifact{}, fmt.Errorf("bad input: run_id")
	}
	if artifactType == "" {
		return Artifact{}, fmt.Errorf("bad input: type")
	}
	return s.repo.GetByRunIDAndType(ctx, runID, artifactType)
}
