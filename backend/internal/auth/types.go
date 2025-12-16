package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	DisplayName  string
	CreatedAt    time.Time
}

var ErrEmailTaken = errors.New("email already in use")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrUserNotfound = errors.New("user not found")
