// Package email provides a simple transactional email sender.
//
// Set EMAIL_PROVIDER to "resend" or "sendgrid" and provide the corresponding
// API key via EMAIL_API_KEY. When EMAIL_PROVIDER is empty, the service runs in
// "log" mode — emails are printed to structured logs instead of actually sent.
// This is safe for development and test environments.
package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

// Sender sends transactional emails.
type Sender struct {
	provider string // "resend", "sendgrid", or "" (log mode)
	apiKey   string
	from     string
	client   *http.Client
}

// NewSender creates a new email sender from environment variables.
//
//	EMAIL_PROVIDER - "resend", "sendgrid", or "" (log mode)
//	EMAIL_API_KEY  - API key for the chosen provider
//	EMAIL_FROM     - sender address (default: noreply@resumetailor.app)
func NewSender() *Sender {
	provider := strings.ToLower(strings.TrimSpace(os.Getenv("EMAIL_PROVIDER")))
	apiKey := strings.TrimSpace(os.Getenv("EMAIL_API_KEY"))
	from := strings.TrimSpace(os.Getenv("EMAIL_FROM"))
	if from == "" {
		from = "Resume Tailor <noreply@resumetailor.app>"
	}

	if provider != "" && apiKey == "" {
		slog.Warn("EMAIL_PROVIDER is set but EMAIL_API_KEY is empty — falling back to log mode")
		provider = ""
	}

	if provider == "" {
		slog.Info("email sender running in log mode (no emails will be sent)")
	} else {
		slog.Info("email sender initialized", "provider", provider)
	}

	return &Sender{
		provider: provider,
		apiKey:   apiKey,
		from:     from,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

// Send sends an email. In log mode, it prints to slog instead.
func (s *Sender) Send(ctx context.Context, to, subject, htmlBody string) error {
	if s.provider == "" {
		slog.Info("email (log mode)",
			"to", to,
			"subject", subject,
			"body_length", len(htmlBody),
		)
		return nil
	}

	switch s.provider {
	case "resend":
		return s.sendResend(ctx, to, subject, htmlBody)
	case "sendgrid":
		return s.sendSendGrid(ctx, to, subject, htmlBody)
	default:
		return fmt.Errorf("unknown email provider: %s", s.provider)
	}
}

// SendVerification sends an email verification link.
func (s *Sender) SendVerification(ctx context.Context, to, verifyURL string) error {
	subject := "Verify your Resume Tailor account"
	body := fmt.Sprintf(`
<div style="font-family: system-ui, sans-serif; max-width: 480px; margin: 0 auto; padding: 24px;">
  <h2 style="color: #ff7a1a; margin-bottom: 16px;">Verify your email</h2>
  <p style="color: #334155; line-height: 1.6;">
    Thanks for signing up for Resume Tailor! Click the button below to verify your email address.
  </p>
  <div style="text-align: center; margin: 24px 0;">
    <a href="%s" style="display: inline-block; background: #ff7a1a; color: #0b0b0f; padding: 12px 32px; border-radius: 999px; text-decoration: none; font-weight: 600; font-size: 14px;">
      Verify Email
    </a>
  </div>
  <p style="color: #94a3b8; font-size: 13px;">
    This link expires in 24 hours. If you didn't create an account, you can safely ignore this email.
  </p>
</div>`, verifyURL)

	return s.Send(ctx, to, subject, body)
}

// SendPasswordReset sends a password reset link.
func (s *Sender) SendPasswordReset(ctx context.Context, to, resetURL string) error {
	subject := "Reset your Resume Tailor password"
	body := fmt.Sprintf(`
<div style="font-family: system-ui, sans-serif; max-width: 480px; margin: 0 auto; padding: 24px;">
  <h2 style="color: #ff7a1a; margin-bottom: 16px;">Reset your password</h2>
  <p style="color: #334155; line-height: 1.6;">
    We received a request to reset your password. Click the button below to choose a new one.
  </p>
  <div style="text-align: center; margin: 24px 0;">
    <a href="%s" style="display: inline-block; background: #ff7a1a; color: #0b0b0f; padding: 12px 32px; border-radius: 999px; text-decoration: none; font-weight: 600; font-size: 14px;">
      Reset Password
    </a>
  </div>
  <p style="color: #94a3b8; font-size: 13px;">
    This link expires in 1 hour. If you didn't request a password reset, you can safely ignore this email.
  </p>
</div>`, resetURL)

	return s.Send(ctx, to, subject, body)
}

// ── Provider implementations ───────────────────────────────────────

func (s *Sender) sendResend(ctx context.Context, to, subject, htmlBody string) error {
	payload := map[string]interface{}{
		"from":    s.from,
		"to":      []string{to},
		"subject": subject,
		"html":    htmlBody,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("resend: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("resend: send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend: status %d", resp.StatusCode)
	}
	return nil
}

func (s *Sender) sendSendGrid(ctx context.Context, to, subject, htmlBody string) error {
	payload := map[string]interface{}{
		"personalizations": []map[string]interface{}{
			{"to": []map[string]string{{"email": to}}},
		},
		"from":    map[string]string{"email": s.from},
		"subject": subject,
		"content": []map[string]string{
			{"type": "text/html", "value": htmlBody},
		},
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("sendgrid: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("sendgrid: send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("sendgrid: status %d", resp.StatusCode)
	}
	return nil
}
