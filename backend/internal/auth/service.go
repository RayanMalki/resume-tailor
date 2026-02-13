package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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

// StartSessionCleanup runs a background goroutine that periodically deletes
// expired sessions from the database so the sessions table doesn't grow
// without bound. It stops when the context is cancelled.
func (s *Service) StartSessionCleanup(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				deleted, err := s.repo.DeleteExpiredSessions(ctx)
				if err != nil {
					slog.Error("session cleanup failed", "error", err)
				} else if deleted > 0 {
					slog.Info("expired sessions cleaned up", "deleted", deleted)
				}
			}
		}
	}()
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

	if err := validatePassword(password); err != nil {
		return uuid.Nil, err
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
	token, expiresAt, _, err = s.LoginWithUser(ctx, email, password)
	return token, expiresAt, err
}

func (s *Service) LoginWithUser(ctx context.Context, email string, password string) (token string, expiresAt time.Time, userID uuid.UUID, err error) {
	loginEmail := strings.TrimSpace(strings.ToLower(email))
	if loginEmail == "" {
		return "", time.Time{}, uuid.Nil, fmt.Errorf("you must enter an email")
	}

	if password == "" {
		return "", time.Time{}, uuid.Nil, fmt.Errorf("you must enter a password")
	}

	u, err := s.repo.GetUserByEmail(ctx, loginEmail)
	if errors.Is(err, ErrUserNotFound) {
		return "", time.Time{}, uuid.Nil, ErrInvalidCredentials
	}
	if err != nil {
		return "", time.Time{}, uuid.Nil, fmt.Errorf("error while searching the email of the user: %v", err)

	}

	if u.AuthProvider != "" && u.AuthProvider != "local" {
		return "", time.Time{}, uuid.Nil, ErrInvalidCredentials
	}

	psswErr := CheckPassword(u.PasswordHash, password)
	if psswErr != nil {
		return "", time.Time{}, uuid.Nil, ErrInvalidCredentials

	}

	generatedToken, err := NewToken()
	if err != nil {
		return "", time.Time{}, uuid.Nil, fmt.Errorf("error generating token: %v", err)

	}

	tokenHash := HashToken(generatedToken)

	timeExpiry := time.Now().Add(30 * 24 * time.Hour)

	sessErr := s.repo.CreateSession(ctx, u.ID, tokenHash, timeExpiry)
	if sessErr != nil {
		return "", time.Time{}, uuid.Nil, fmt.Errorf("error while creating session: %v", sessErr)

	}

	return generatedToken, timeExpiry, u.ID, nil

}

func (s *Service) LoginWithOAuth(ctx context.Context, provider, sub, email, displayName string) (token string, expiresAt time.Time, userID uuid.UUID, err error) {
	if strings.TrimSpace(provider) == "" || strings.TrimSpace(sub) == "" {
		return "", time.Time{}, uuid.Nil, fmt.Errorf("oauth provider info missing")
	}

	user, err := s.repo.GetUserByOAuth(ctx, provider, sub)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return "", time.Time{}, uuid.Nil, err
	}

	if err == nil {
		return s.createSession(ctx, user.ID)
	}

	email = strings.TrimSpace(strings.ToLower(email))
	displayName = strings.TrimSpace(displayName)
	if email != "" {
		existing, err := s.repo.GetUserByEmail(ctx, email)
		if err == nil {
			if existing.OAuthProvider == nil && existing.OAuthSub == nil {
				_ = s.repo.LinkOAuth(ctx, existing.ID, provider, sub)
				return s.createSession(ctx, existing.ID)
			}
			return "", time.Time{}, uuid.Nil, ErrInvalidCredentials
		}
		if err != nil && !errors.Is(err, ErrUserNotFound) {
			return "", time.Time{}, uuid.Nil, err
		}
	}

	if displayName == "" {
		displayName = "Google User"
	}

	randomToken, err := NewToken()
	if err != nil {
		return "", time.Time{}, uuid.Nil, err
	}
	randomHash, err := HashPassword(randomToken)
	if err != nil {
		return "", time.Time{}, uuid.Nil, err
	}

	id, err := s.repo.CreateOAuthUser(ctx, email, randomHash, displayName, provider, sub)
	if err != nil {
		return "", time.Time{}, uuid.Nil, err
	}

	return s.createSession(ctx, id)
}

func (s *Service) createSession(ctx context.Context, userID uuid.UUID) (token string, expiresAt time.Time, outUserID uuid.UUID, err error) {
	generatedToken, err := NewToken()
	if err != nil {
		return "", time.Time{}, uuid.Nil, err
	}

	tokenHash := HashToken(generatedToken)
	timeExpiry := time.Now().Add(30 * 24 * time.Hour)

	if err := s.repo.CreateSession(ctx, userID, tokenHash, timeExpiry); err != nil {
		return "", time.Time{}, uuid.Nil, err
	}

	return generatedToken, timeExpiry, userID, nil
}

func validatePassword(password string) error {
	if len(password) < 10 {
		return fmt.Errorf("%w: must be at least 10 characters", ErrWeakPassword)
	}

	var hasUpper, hasLower, hasDigit, hasSymbol bool
	for _, ch := range password {
		switch {
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		default:
			hasSymbol = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit || !hasSymbol {
		return fmt.Errorf("%w: must include upper, lower, number, and symbol", ErrWeakPassword)
	}

	common := []string{"password", "1234567890", "qwerty", "letmein"}
	lower := strings.ToLower(password)
	for _, bad := range common {
		if strings.Contains(lower, bad) {
			return fmt.Errorf("%w: too common", ErrWeakPassword)
		}
	}

	return nil
}
