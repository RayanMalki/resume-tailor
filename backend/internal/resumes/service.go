package resumes

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateResume(ctx context.Context, userID uuid.UUID, title, contentText string) (Resume, error) {
	if userID == uuid.Nil {
		return Resume{}, fmt.Errorf("bad input: user_id")
	}

	title = strings.TrimSpace(title)
	if title == "" {
		return Resume{}, fmt.Errorf("bad input: title")
	}

	contentText = strings.TrimSpace(contentText)
	if contentText == "" {
		return Resume{}, fmt.Errorf("bad input: content_text")
	}

	return s.repo.CreateResume(ctx, userID, title, contentText)
}

func (s *Service) GetResumeByID(ctx context.Context, userID, resumeID uuid.UUID) (Resume, error) {
	if userID == uuid.Nil {
		return Resume{}, fmt.Errorf("bad input: user_id")
	}
	if resumeID == uuid.Nil {
		return Resume{}, fmt.Errorf("bad input: resume_id")
	}

	res, err := s.repo.GetResumeByID(ctx, resumeID)
	if err != nil {
		if errors.Is(err, ErrResumeNotFound) {
			return Resume{}, ErrResumeNotFound
		}
		return Resume{}, err
	}

	if res.UserID != userID {
		return Resume{}, ErrResumeNotFound
	}

	return res, nil
}

func (s *Service) ListResumesByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Resume, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("bad input: user_id")
	}

	return s.repo.ListResumesByUser(ctx, userID, limit, offset)
}

