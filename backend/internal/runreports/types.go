package runreports

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type RunReport struct {
	RunID      uuid.UUID
	ATSReport  json.RawMessage
	ChangePlan json.RawMessage
	CreatedAt  time.Time
}

var (
	ErrRunReportNotFound = errors.New("run report not found")
	ErrBadInput          = errors.New("bad input")
)

