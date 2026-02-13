package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrVerificationTokenNotFound = errors.New("verification token not found")
	ErrVerificationTokenExpired  = errors.New("verification token expired")
	ErrVerificationTokenUsed     = errors.New("verification token already used")
	ErrResetTokenNotFound        = errors.New("reset token not found")
	ErrResetTokenExpired         = errors.New("reset token expired")
	ErrResetTokenUsed            = errors.New("reset token already used")
	ErrAlreadyVerified           = errors.New("email already verified")
)

// ── Verification token methods ─────────────────────────────────────

// CreateVerificationToken generates a new email verification token for a user.
// Returns the raw (unhashed) token for inclusion in the verification link.
func (s *Service) CreateVerificationToken(ctx context.Context, userID uuid.UUID) (string, error) {
	rawToken, err := NewToken()
	if err != nil {
		return "", fmt.Errorf("generate verification token: %w", err)
	}

	tokenHash := HashToken(rawToken)
	expiresAt := time.Now().Add(24 * time.Hour)

	err = s.repo.CreateVerificationToken(ctx, userID, tokenHash, expiresAt)
	if err != nil {
		return "", fmt.Errorf("store verification token: %w", err)
	}

	return rawToken, nil
}

// VerifyEmail validates a verification token and marks the user as verified.
func (s *Service) VerifyEmail(ctx context.Context, rawToken string) error {
	tokenHash := HashToken(rawToken)

	// Look up the token
	vt, err := s.repo.GetVerificationToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, ErrVerificationTokenNotFound) {
			return ErrVerificationTokenNotFound
		}
		return err
	}

	if vt.UsedAt != nil {
		return ErrVerificationTokenUsed
	}

	if time.Now().After(vt.ExpiresAt) {
		return ErrVerificationTokenExpired
	}

	// Mark token as used
	if err := s.repo.MarkVerificationTokenUsed(ctx, tokenHash); err != nil {
		return err
	}

	// Mark user as verified
	if err := s.repo.SetUserVerified(ctx, vt.UserID); err != nil {
		return err
	}

	return nil
}

// ── Password reset methods ─────────────────────────────────────────

// CreatePasswordResetToken generates a password reset token.
// Returns the raw token for inclusion in the reset link.
func (s *Service) CreatePasswordResetToken(ctx context.Context, email string) (string, uuid.UUID, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			// Don't reveal whether the email exists
			return "", uuid.Nil, nil
		}
		return "", uuid.Nil, err
	}

	rawToken, err := NewToken()
	if err != nil {
		return "", uuid.Nil, fmt.Errorf("generate reset token: %w", err)
	}

	tokenHash := HashToken(rawToken)
	expiresAt := time.Now().Add(1 * time.Hour)

	err = s.repo.CreatePasswordResetToken(ctx, user.ID, tokenHash, expiresAt)
	if err != nil {
		return "", uuid.Nil, fmt.Errorf("store reset token: %w", err)
	}

	return rawToken, user.ID, nil
}

// ResetPassword validates a reset token and updates the user's password.
func (s *Service) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	tokenHash := HashToken(rawToken)

	rt, err := s.repo.GetPasswordResetToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, ErrResetTokenNotFound) {
			return ErrResetTokenNotFound
		}
		return err
	}

	if rt.UsedAt != nil {
		return ErrResetTokenUsed
	}

	if time.Now().After(rt.ExpiresAt) {
		return ErrResetTokenExpired
	}

	passwordHash, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	if err := s.repo.UpdateUserPassword(ctx, rt.UserID, passwordHash); err != nil {
		return err
	}

	if err := s.repo.MarkPasswordResetTokenUsed(ctx, tokenHash); err != nil {
		return err
	}

	return nil
}

// ── Repo methods for verification ──────────────────────────────────

type VerificationToken struct {
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
}

type PasswordResetToken struct {
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
}

func (r *Repo) CreateVerificationToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	const q = `INSERT INTO email_verification_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, q, userID, tokenHash, expiresAt)
	return err
}

func (r *Repo) GetVerificationToken(ctx context.Context, tokenHash string) (VerificationToken, error) {
	const q = `SELECT user_id, token_hash, expires_at, used_at FROM email_verification_tokens WHERE token_hash = $1`
	var vt VerificationToken
	err := r.db.QueryRow(ctx, q, tokenHash).Scan(&vt.UserID, &vt.TokenHash, &vt.ExpiresAt, &vt.UsedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return VerificationToken{}, ErrVerificationTokenNotFound
		}
		return VerificationToken{}, err
	}
	return vt, nil
}

func (r *Repo) MarkVerificationTokenUsed(ctx context.Context, tokenHash string) error {
	const q = `UPDATE email_verification_tokens SET used_at = now() WHERE token_hash = $1`
	_, err := r.db.Exec(ctx, q, tokenHash)
	return err
}

func (r *Repo) SetUserVerified(ctx context.Context, userID uuid.UUID) error {
	const q = `UPDATE users SET is_verified = true, updated_at = now() WHERE id = $1`
	_, err := r.db.Exec(ctx, q, userID)
	return err
}

func (r *Repo) CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	const q = `INSERT INTO password_resets (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, q, userID, tokenHash, expiresAt)
	return err
}

func (r *Repo) GetPasswordResetToken(ctx context.Context, tokenHash string) (PasswordResetToken, error) {
	const q = `SELECT user_id, token_hash, expires_at, used_at FROM password_resets WHERE token_hash = $1`
	var rt PasswordResetToken
	err := r.db.QueryRow(ctx, q, tokenHash).Scan(&rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.UsedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PasswordResetToken{}, ErrResetTokenNotFound
		}
		return PasswordResetToken{}, err
	}
	return rt, nil
}

func (r *Repo) MarkPasswordResetTokenUsed(ctx context.Context, tokenHash string) error {
	const q = `UPDATE password_resets SET used_at = now() WHERE token_hash = $1`
	_, err := r.db.Exec(ctx, q, tokenHash)
	return err
}

func (r *Repo) UpdateUserPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	const q = `UPDATE users SET password_hash = $1, updated_at = now() WHERE id = $2`
	_, err := r.db.Exec(ctx, q, passwordHash, userID)
	return err
}
