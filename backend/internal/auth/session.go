package auth

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
}
