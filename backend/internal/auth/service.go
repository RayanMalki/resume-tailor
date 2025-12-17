package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

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

func (s *Service) Signup(ctx context.Context, email string, password string, displayName string) (uuid.UUID, error) {

	signupEmail := strings.TrimSpace(strings.ToLower(email))
	signupDisplayName := strings.TrimSpace(displayName)

	if signupEmail == "" {
		return uuid.Nil, fmt.Errorf("you must enter an email")
	}
	if signupDisplayName == "" {
		return uuid.Nil, fmt.Errorf("you must enter a username")
	}
	if password == "" {
		return uuid.Nil, fmt.Errorf("you must enter a password")
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error while hashing password: %v", err)
	}

	id, err := s.repo.CreateUser(ctx, signupEmail, passwordHash, signupDisplayName)

	if errors.Is(err, ErrEmailTaken) {
		return uuid.Nil, ErrEmailTaken
	}

	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (s *Service) Login(ctx context.Context, email string, password string) (token string, expiresAt time.Time, err error) {

	loginEmail := strings.TrimSpace(strings.ToLower(email))
	if loginEmail == "" {
		return "", time.Time{}, fmt.Errorf("you must enter an email")
	}

	if password == "" {
		return "", time.Time{}, fmt.Errorf("you must enter a password")
	}

	u, err := s.repo.GetUserByEmail(ctx, loginEmail)
	if errors.Is(err, ErrUserNotFound) {
		return "", time.Time{}, ErrInvalidCredentials
	}
	if err != nil {
		return "", time.Time{}, fmt.Errorf("error while searching the email of the user: %v", err)

	}

	psswErr := CheckPassword(u.PasswordHash, password)
	if psswErr != nil {
		return "", time.Time{}, ErrInvalidCredentials

	}

	generatedToken, err := NewToken()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("error generating token: %v", err)

	}

	tokenHash := HashToken(generatedToken)

	timeExpiry := time.Now().Add(30 * 24 * time.Hour)

	sessErr := s.repo.CreateSession(ctx, u.ID, tokenHash, timeExpiry)
	if sessErr != nil {
		return "", time.Time{}, fmt.Errorf("error while creating session: %v", sessErr)

	}

	return generatedToken, timeExpiry, nil

}
