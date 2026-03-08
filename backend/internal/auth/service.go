package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"resume-tailor/internal/crypto"

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

func (s *Service) LoginWithOAuth(ctx context.Context, provider, sub, email, displayName, avatarURL string) (token string, expiresAt time.Time, userID uuid.UUID, err error) {
	if strings.TrimSpace(provider) == "" || strings.TrimSpace(sub) == "" {
		return "", time.Time{}, uuid.Nil, fmt.Errorf("oauth provider info missing")
	}

	user, err := s.repo.GetUserByOAuth(ctx, provider, sub)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return "", time.Time{}, uuid.Nil, err
	}

	if err == nil {
		// Refresh avatar URL if it has changed.
		if avatarURL != "" {
			var current *string
			if user.AvatarURL != nil {
				current = user.AvatarURL
			}
			if current == nil || *current != avatarURL {
				_ = s.repo.UpdateAvatarURL(ctx, user.ID, &avatarURL)
			}
		}
		return s.createSession(ctx, user.ID)
	}

	email = strings.TrimSpace(strings.ToLower(email))
	displayName = strings.TrimSpace(displayName)
	if email != "" {
		existing, err := s.repo.GetUserByEmail(ctx, email)
		if err == nil {
			if existing.OAuthProvider == nil && existing.OAuthSub == nil {
				_ = s.repo.LinkOAuth(ctx, existing.ID, provider, sub)
				if avatarURL != "" {
					_ = s.repo.UpdateAvatarURL(ctx, existing.ID, &avatarURL)
				}
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

	id, err := s.repo.CreateOAuthUser(ctx, email, randomHash, displayName, provider, sub, avatarURL)
	if err != nil {
		return "", time.Time{}, uuid.Nil, err
	}

	return s.createSession(ctx, id)
}

// GetUserByID returns the user with the given ID.
func (s *Service) GetUserByID(ctx context.Context, userID uuid.UUID) (User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

// ChangePassword validates currentPassword and replaces the hash with newPassword.
func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.AuthProvider != "" && user.AuthProvider != "local" {
		return fmt.Errorf("password change not available for OAuth accounts")
	}
	if err := CheckPassword(user.PasswordHash, currentPassword); err != nil {
		return ErrInvalidCredentials
	}
	if err := validatePassword(newPassword); err != nil {
		return err
	}
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.repo.UpdatePasswordHash(ctx, userID, hash)
}

// SetAPIKey encrypts plaintextKey with encSecret and stores it for the user.
func (s *Service) SetAPIKey(ctx context.Context, userID uuid.UUID, plaintextKey, encSecret string) error {
	encrypted, err := crypto.Encrypt(encSecret, plaintextKey)
	if err != nil {
		return err
	}
	return s.repo.SetEncryptedAPIKey(ctx, userID, &encrypted)
}

// RemoveAPIKey clears the stored encrypted API key for the user.
func (s *Service) RemoveAPIKey(ctx context.Context, userID uuid.UUID) error {
	return s.repo.SetEncryptedAPIKey(ctx, userID, nil)
}

// DeleteAccount permanently deletes the user and all their data.
func (s *Service) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	return s.repo.DeleteUser(ctx, userID)
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
