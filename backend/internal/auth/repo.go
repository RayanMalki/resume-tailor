package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repo provides persistence methods for auth-related data.
type Repo struct {
	db *pgxpool.Pool
}

// NewRepo creates a new Repo with the given pgxpool.
func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

// CreateUser inserts a new user and returns the generated ID.
func (r *Repo) CreateUser(ctx context.Context, email, passwordHash, displayName string) (uuid.UUID, error) {
	const q = `
INSERT INTO users (email, password_hash, display_name)
VALUES ($1, $2, $3)
RETURNING id`

	var id uuid.UUID
	err := r.db.QueryRow(ctx, q, email, passwordHash, displayName).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// unique_violation (likely email already taken)
			return uuid.Nil, ErrEmailTaken
		}
		return uuid.Nil, err
	}

	return id, nil
}

// CreateOAuthUser creates a user with an OAuth provider and sub.
func (r *Repo) CreateOAuthUser(ctx context.Context, email, passwordHash, displayName, provider, sub, avatarURL string) (uuid.UUID, error) {
	const q = `
INSERT INTO users (email, password_hash, display_name, auth_provider, oauth_provider, oauth_sub, avatar_url)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id`

	var avatarArg any
	if avatarURL != "" {
		avatarArg = avatarURL
	}

	var id uuid.UUID
	err := r.db.QueryRow(ctx, q, email, passwordHash, displayName, provider, provider, sub, avatarArg).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return uuid.Nil, ErrEmailTaken
		}
		return uuid.Nil, err
	}

	return id, nil
}

// GetUserByEmail fetches a user by email.
func (r *Repo) GetUserByEmail(ctx context.Context, email string) (User, error) {
	const q = `
	SELECT id, email, password_hash, display_name, auth_provider, oauth_provider, oauth_sub, avatar_url, encrypted_openai_key, onboarding_seen, created_at, updated_at
	FROM users
	WHERE email = $1`

	var u User
	err := r.db.QueryRow(ctx, q, email).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.DisplayName,
		&u.AuthProvider,
		&u.OAuthProvider,
		&u.OAuthSub,
		&u.AvatarURL,
		&u.EncryptedOpenAIKey,
		&u.OnboardingSeen,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}

	return u, nil
}

// GetUserByOAuth fetches a user by oauth provider+sub.
func (r *Repo) GetUserByOAuth(ctx context.Context, provider, sub string) (User, error) {
	const q = `
	SELECT id, email, password_hash, display_name, auth_provider, oauth_provider, oauth_sub, avatar_url, encrypted_openai_key, onboarding_seen, created_at, updated_at
	FROM users
	WHERE oauth_provider = $1 AND oauth_sub = $2`

	var u User
	err := r.db.QueryRow(ctx, q, provider, sub).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.DisplayName,
		&u.AuthProvider,
		&u.OAuthProvider,
		&u.OAuthSub,
		&u.AvatarURL,
		&u.EncryptedOpenAIKey,
		&u.OnboardingSeen,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}

	return u, nil
}

// GetUserByID fetches a user by their primary key.
func (r *Repo) GetUserByID(ctx context.Context, userID uuid.UUID) (User, error) {
	const q = `
	SELECT id, email, password_hash, display_name, auth_provider, oauth_provider, oauth_sub, avatar_url, encrypted_openai_key, onboarding_seen, created_at, updated_at
	FROM users
	WHERE id = $1`

	var u User
	err := r.db.QueryRow(ctx, q, userID).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.DisplayName,
		&u.AuthProvider,
		&u.OAuthProvider,
		&u.OAuthSub,
		&u.AvatarURL,
		&u.EncryptedOpenAIKey,
		&u.OnboardingSeen,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}

	return u, nil
}

// UpdatePasswordHash sets a new bcrypt hash for the user.
func (r *Repo) UpdatePasswordHash(ctx context.Context, userID uuid.UUID, hash string) error {
	const q = `UPDATE users SET password_hash = $2, updated_at = now() WHERE id = $1`
	_, err := r.db.Exec(ctx, q, userID, hash)
	return err
}

// UpdateAvatarURL sets or clears the avatar_url for a user.
func (r *Repo) UpdateAvatarURL(ctx context.Context, userID uuid.UUID, avatarURL *string) error {
	const q = `UPDATE users SET avatar_url = $2, updated_at = now() WHERE id = $1`
	_, err := r.db.Exec(ctx, q, userID, avatarURL)
	return err
}

// SetOnboardingSeen marks the onboarding_seen flag as true for a user.
func (r *Repo) SetOnboardingSeen(ctx context.Context, userID uuid.UUID) error {
	const q = `UPDATE users SET onboarding_seen = true, updated_at = now() WHERE id = $1`
	_, err := r.db.Exec(ctx, q, userID)
	return err
}

// SetEncryptedAPIKey stores or clears the encrypted_openai_key for a user.
func (r *Repo) SetEncryptedAPIKey(ctx context.Context, userID uuid.UUID, encrypted *string) error {
	const q = `UPDATE users SET encrypted_openai_key = $2, updated_at = now() WHERE id = $1`
	_, err := r.db.Exec(ctx, q, userID, encrypted)
	return err
}

// GetEncryptedOpenAIKey returns the encrypted_openai_key for a user (implements jobs.UserKeyRepo).
func (r *Repo) GetEncryptedOpenAIKey(ctx context.Context, userID uuid.UUID) (*string, error) {
	const q = `SELECT encrypted_openai_key FROM users WHERE id = $1`
	var key *string
	err := r.db.QueryRow(ctx, q, userID).Scan(&key)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return key, nil
}

// DeleteUser deletes the user row; sessions and runs cascade via FK.
func (r *Repo) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	const q = `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, q, userID)
	return err
}

// LinkOAuth sets oauth provider/sub for an existing user if not already set.
func (r *Repo) LinkOAuth(ctx context.Context, userID uuid.UUID, provider, sub string) error {
	if userID == uuid.Nil {
		return fmt.Errorf("bad input: user_id")
	}
	if provider == "" || sub == "" {
		return fmt.Errorf("bad input: oauth")
	}

	const q = `
	UPDATE users
	SET oauth_provider = $2, oauth_sub = $3, updated_at = now()
	WHERE id = $1 AND oauth_provider IS NULL AND oauth_sub IS NULL`

	cmd, err := r.db.Exec(ctx, q, userID, provider, sub)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("oauth already linked")
	}
	return nil
}

// CreateSession creates a new session row for the given user.
func (r *Repo) CreateSession(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	const q = `
	INSERT INTO sessions (user_id, token_hash, expires_at)
	VALUES ($1, $2, $3)`

	_, err := r.db.Exec(ctx, q, userID, tokenHash, expiresAt)
	return err
}

func (r *Repo) GetSessionByTokenHash(ctx context.Context, tokenHash string) (Session, error) {
	if tokenHash == "" {
		return Session{}, fmt.Errorf("bad imput: token_hash")
	}
	const q = `SELECT user_id, token_hash, expires_at FROM sessions WHERE token_hash = $1`

	var session Session
	err := r.db.QueryRow(ctx, q, tokenHash).Scan(
		&session.UserID,
		&session.TokenHash,
		&session.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Session{}, ErrTokenNotFound
		}
		return Session{}, err
	}
	return session, nil

}

func (r *Repo) DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error {
	if tokenHash == "" {
		return fmt.Errorf("bad input: token_hash")
	}

	const q = `DELETE FROM sessions WHERE token_hash = $1`

	cmdTag, err := r.db.Exec(ctx, q, tokenHash)
	if err != nil {
		return err

	}
	if cmdTag.RowsAffected() == 0 {
		return nil

	}
	return nil

}

// DeleteExpiredSessions removes all sessions that have passed their expiry time.
func (r *Repo) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	const q = `DELETE FROM sessions WHERE expires_at < now()`
	cmd, err := r.db.Exec(ctx, q)
	if err != nil {
		return 0, err
	}
	return cmd.RowsAffected(), nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (uuid.UUID, error) {
	if strings.TrimSpace(token) == "" {
		return uuid.Nil, ErrInvalidCredentials

	}

	tokenHash := HashToken(strings.TrimSpace(token))

	sess, err := s.repo.GetSessionByTokenHash(ctx, tokenHash)
	if errors.Is(err, ErrTokenNotFound) {
		return uuid.Nil, ErrInvalidCredentials
	}
	if err != nil {
		return uuid.Nil, err

	}

	now := time.Now()

	if now.After(sess.ExpiresAt) {
		err := s.repo.DeleteSessionByTokenHash(ctx, tokenHash)
		if err != nil {
			return uuid.Nil, err
		}
		return uuid.Nil, ErrInvalidCredentials
	}

	return sess.UserID, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	tokenTrim := strings.TrimSpace(token)
	if tokenTrim == "" {
		return nil
	}
	tokenHash := HashToken(tokenTrim)
	err := s.repo.DeleteSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		return err
	}
	return nil
}
