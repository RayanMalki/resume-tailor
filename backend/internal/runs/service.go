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
	repo    *Repo
	jobsEnq jobs.JobsEnqueuer
}

func NewService(repo *Repo, jobsEnq jobs.JobsEnqueuer) *Service {
	return &Service{
		repo:    repo,
		jobsEnq: jobsEnq,
	}
}

func (s *Service) CreateRun(ctx context.Context, userID,
	resumeID uuid.UUID, jobText string, projectControls []ProjectControl, disciplineOverride *Discipline, creatorIP string) (Run, error) {

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

	normalizedControls := normalizeProjectControls(projectControls)
	disciplineSource := DisciplineSourceAuto
	disciplineScore := 0.0
	var discipline *Discipline
	if disciplineOverride != nil {
		override := *disciplineOverride
		discipline = &override
		disciplineSource = DisciplineSourceUserOverride
		disciplineScore = 1.0
	}

	run, err := s.repo.CreateRun(ctx, userID, resumeID, jobText, normalizedControls, discipline, disciplineScore, disciplineSource, creatorIP)
	if err != nil {
		return Run{}, err
	}

	// Enqueue job for processing
	if s.jobsEnq != nil {
		_, err = s.jobsEnq.EnqueueProcessRun(ctx, run.ID)
		if err != nil {
			enqueueMsg := "enqueue_failed"
			_ = s.repo.UpdateRunStatus(ctx, run.ID, string(StatusFailed), &enqueueMsg)
			return Run{}, fmt.Errorf("failed to enqueue job: %w", err)
		}
	}

	return run, nil

}

func (s *Service) UpdateRunDiscipline(ctx context.Context, runID uuid.UUID, discipline Discipline, confidence float64, source DisciplineSource) error {
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
	return s.repo.UpdateRunDiscipline(ctx, runID, discipline, confidence, source)
}

func normalizeProjectControls(controls []ProjectControl) []ProjectControl {
	if len(controls) == 0 {
		return nil
	}
	out := make([]ProjectControl, 0, len(controls))
	seen := make(map[string]struct{}, len(controls))
	for _, c := range controls {
		name := strings.TrimSpace(c.Name)
		if name == "" {
			continue
		}
		mode := c.Mode
		switch mode {
		case ProjectControlPinned, ProjectControlAuto, ProjectControlExclude:
		default:
			mode = ProjectControlAuto
		}
		key := strings.ToLower(name)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, ProjectControl{Name: name, Mode: mode})
	}
	return out
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

// CountRunsTodayForUser returns how many runs the user has created today.
func (s *Service) CountRunsTodayForUser(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.repo.CountRunsTodayForUser(ctx, userID)
}

// CountRunsTodayForIP returns how many runs have been created today from the given IP.
func (s *Service) CountRunsTodayForIP(ctx context.Context, ip string) (int, error) {
	return s.repo.CountRunsTodayForIP(ctx, ip)
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
