# internal/email

- Relative path: `internal/email`
- Packages: `email`
- Source files: `email.go`

## Functions

### `func NewSender() *Sender`

- File: `email.go`
- Purpose: NewSender creates a new email sender from environment variables.

### `func (s *Sender) Send(ctx context.Context, to, subject, htmlBody string) error`

- File: `email.go`
- Purpose: Send sends an email. In log mode, it prints to slog instead.

### `func (s *Sender) SendPasswordReset(ctx context.Context, to, resetURL string) error`

- File: `email.go`
- Purpose: SendPasswordReset sends a password reset link.

### `func (s *Sender) SendVerification(ctx context.Context, to, verifyURL string) error`

- File: `email.go`
- Purpose: SendVerification sends an email verification link.

### `func (s *Sender) sendResend(ctx context.Context, to, subject, htmlBody string) error`

- File: `email.go`
- Purpose: Internal function supporting `internal/email` package workflows.

### `func (s *Sender) sendSendGrid(ctx context.Context, to, subject, htmlBody string) error`

- File: `email.go`
- Purpose: Internal function supporting `internal/email` package workflows.

