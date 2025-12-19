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

// GetUserByEmail fetches a user by email.
func (r *Repo) GetUserByEmail(ctx context.Context, email string) (User, error) {
	const q = `
	SELECT id, email, password_hash, display_name, created_at
	FROM users
	WHERE email = $1`

	var u User
	err := r.db.QueryRow(ctx, q, email).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.DisplayName,
		&u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}

	return u, nil
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
