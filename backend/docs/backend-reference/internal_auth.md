# internal/auth

- Relative path: `internal/auth`
- Packages: `auth`
- Source files: `password.go`, `repo.go`, `service.go`, `session.go`, `token.go`, `types.go`, `verification.go`

## Functions

### `func CheckPassword(hashedpassword string, password string) error`

- File: `password.go`
- Purpose: Performs credential hashing or verification logic.

### `func HashPassword(password string) (string, error)`

- File: `password.go`
- Purpose: Performs credential hashing or verification logic.

### `func (s *Service) Authenticate(ctx context.Context, token string) (uuid.UUID, error)`

- File: `repo.go`
- Purpose: Applies authentication checks and request identity propagation.

### `func (r *Repo) CreateOAuthUser(ctx context.Context, email, passwordHash, displayName, provider, sub string) (uuid.UUID, error)`

- File: `repo.go`
- Purpose: CreateOAuthUser creates a user with an OAuth provider and sub.

### `func (r *Repo) CreateSession(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error`

- File: `repo.go`
- Purpose: CreateSession creates a new session row for the given user.

### `func (r *Repo) CreateUser(ctx context.Context, email, passwordHash, displayName string) (uuid.UUID, error)`

- File: `repo.go`
- Purpose: CreateUser inserts a new user and returns the generated ID.

### `func (r *Repo) DeleteExpiredSessions(ctx context.Context) (int64, error)`

- File: `repo.go`
- Purpose: DeleteExpiredSessions removes all sessions that have passed their expiry time.

### `func (r *Repo) DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error`

- File: `repo.go`
- Purpose: Deletes or revokes an existing resource.

### `func (r *Repo) GetSessionByTokenHash(ctx context.Context, tokenHash string) (Session, error)`

- File: `repo.go`
- Purpose: Fetches a single resource and returns it with validation/error handling.

### `func (r *Repo) GetUserByEmail(ctx context.Context, email string) (User, error)`

- File: `repo.go`
- Purpose: GetUserByEmail fetches a user by email.

### `func (r *Repo) GetUserByOAuth(ctx context.Context, provider, sub string) (User, error)`

- File: `repo.go`
- Purpose: GetUserByOAuth fetches a user by oauth provider+sub.

### `func (r *Repo) LinkOAuth(ctx context.Context, userID uuid.UUID, provider, sub string) error`

- File: `repo.go`
- Purpose: LinkOAuth sets oauth provider/sub for an existing user if not already set.

### `func (s *Service) Logout(ctx context.Context, token string) error`

- File: `repo.go`
- Purpose: Internal function supporting `internal/auth` package workflows.

### `func NewRepo(db *pgxpool.Pool) *Repo`

- File: `repo.go`
- Purpose: NewRepo creates a new Repo with the given pgxpool.

### `func (s *Service) Login(ctx context.Context, email string, password string) (token string, expiresAt time.Time, err error)`

- File: `service.go`
- Purpose: Internal function supporting `internal/auth` package workflows.

### `func (s *Service) LoginWithOAuth(ctx context.Context, provider, sub, email, displayName string) (token string, expiresAt time.Time, userID uuid.UUID, err error)`

- File: `service.go`
- Purpose: Applies authentication checks and request identity propagation.

### `func (s *Service) LoginWithUser(ctx context.Context, email string, password string) (token string, expiresAt time.Time, userID uuid.UUID, err error)`

- File: `service.go`
- Purpose: Internal function supporting `internal/auth` package workflows.

### `func NewService(repo *Repo) *Service`

- File: `service.go`
- Purpose: Creates and returns a new instance used by this package.

### `func (s *Service) Signup(ctx context.Context, email string, password string, displayName string) (uuid.UUID, error)`

- File: `service.go`
- Purpose: Internal function supporting `internal/auth` package workflows.

### `func (s *Service) StartSessionCleanup(ctx context.Context, interval time.Duration)`

- File: `service.go`
- Purpose: StartSessionCleanup runs a background goroutine that periodically deletes.

### `func (s *Service) createSession(ctx context.Context, userID uuid.UUID) (token string, expiresAt time.Time, outUserID uuid.UUID, err error)`

- File: `service.go`
- Purpose: Internal function supporting `internal/auth` package workflows.

### `func validatePassword(password string) error`

- File: `service.go`
- Purpose: Internal function supporting `internal/auth` package workflows.

### `func HashToken(token string) string`

- File: `token.go`
- Purpose: Handles token creation, hashing, parsing, or token-derived state.

### `func NewToken() (string, error)`

- File: `token.go`
- Purpose: Creates and returns a new instance used by this package.

### `func (s *Service) CreatePasswordResetToken(ctx context.Context, email string) (string, uuid.UUID, error)`

- File: `verification.go`
- Purpose: CreatePasswordResetToken generates a password reset token.

### `func (r *Repo) CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error`

- File: `verification.go`
- Purpose: Creates a new record or resource and returns the result.

### `func (s *Service) CreateVerificationToken(ctx context.Context, userID uuid.UUID) (string, error)`

- File: `verification.go`
- Purpose: CreateVerificationToken generates a new email verification token for a user.

### `func (r *Repo) CreateVerificationToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error`

- File: `verification.go`
- Purpose: Creates a new record or resource and returns the result.

### `func (r *Repo) GetPasswordResetToken(ctx context.Context, tokenHash string) (PasswordResetToken, error)`

- File: `verification.go`
- Purpose: Fetches a single resource and returns it with validation/error handling.

### `func (r *Repo) GetVerificationToken(ctx context.Context, tokenHash string) (VerificationToken, error)`

- File: `verification.go`
- Purpose: Fetches a single resource and returns it with validation/error handling.

### `func (r *Repo) MarkPasswordResetTokenUsed(ctx context.Context, tokenHash string) error`

- File: `verification.go`
- Purpose: Marks state transitions on an existing record.

### `func (r *Repo) MarkVerificationTokenUsed(ctx context.Context, tokenHash string) error`

- File: `verification.go`
- Purpose: Marks state transitions on an existing record.

### `func (s *Service) ResetPassword(ctx context.Context, rawToken, newPassword string) error`

- File: `verification.go`
- Purpose: ResetPassword validates a reset token and updates the user's password.

### `func (r *Repo) SetUserVerified(ctx context.Context, userID uuid.UUID) error`

- File: `verification.go`
- Purpose: Internal function supporting `internal/auth` package workflows.

### `func (r *Repo) UpdateUserPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error`

- File: `verification.go`
- Purpose: Updates persisted state for an existing resource.

### `func (s *Service) VerifyEmail(ctx context.Context, rawToken string) error`

- File: `verification.go`
- Purpose: VerifyEmail validates a verification token and marks the user as verified.

