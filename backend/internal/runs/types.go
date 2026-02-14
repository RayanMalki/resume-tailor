package runs

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
)

type Run struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	ResumeID         uuid.UUID
	JobText          string
	ProjectControls  []ProjectControl
	Discipline       *Discipline
	DisciplineScore  float64
	DisciplineSource DisciplineSource
	Status           Status
	ErrorMessage     *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type ProjectControlMode string

const (
	ProjectControlPinned  ProjectControlMode = "pinned"
	ProjectControlAuto    ProjectControlMode = "auto"
	ProjectControlExclude ProjectControlMode = "exclude"
)

type ProjectControl struct {
	Name string             `json:"name"`
	Mode ProjectControlMode `json:"mode"`
}

type Discipline string

const (
	DisciplineMechanical          Discipline = "mechanical"
	DisciplineElectrical          Discipline = "electrical"
	DisciplineIndustrialLogistics Discipline = "industrial_logistics"
	DisciplineAerospace           Discipline = "aerospace"
	DisciplineITSoftware          Discipline = "it_software"
)

func ParseDiscipline(raw string) (Discipline, bool) {
	switch Discipline(raw) {
	case DisciplineMechanical,
		DisciplineElectrical,
		DisciplineIndustrialLogistics,
		DisciplineAerospace,
		DisciplineITSoftware:
		return Discipline(raw), true
	default:
		return "", false
	}
}

type DisciplineSource string

const (
	DisciplineSourceAuto         DisciplineSource = "auto"
	DisciplineSourceUserOverride DisciplineSource = "user_override"
)

var (
	ErrRunNotFound = errors.New("run failed")
	ErrForbidden   = errors.New("forbidden")
	ErrBadInput    = errors.New("bad input")
)
