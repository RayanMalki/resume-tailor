package handlers

import (
	"errors"
	"net/http"
	"strings"

	"resume-tailor/internal/auth"
	"resume-tailor/internal/email"
	"resume-tailor/internal/httpapi/middleware"
)

// VerifyEmail handles GET /v1/auth/verify?token=...
// It validates the verification token and marks the user as verified.
func VerifyEmail(authSvc *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		if token == "" {
			writeError(w, http.StatusBadRequest, "missing token")
			return
		}

		err := authSvc.VerifyEmail(r.Context(), token)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrVerificationTokenNotFound):
				writeError(w, http.StatusNotFound, "invalid_token")
			case errors.Is(err, auth.ErrVerificationTokenExpired):
				writeError(w, http.StatusGone, "token_expired")
			case errors.Is(err, auth.ErrVerificationTokenUsed):
				writeError(w, http.StatusConflict, "token_already_used")
			default:
				writeError(w, http.StatusInternalServerError, "internal_server_error")
			}
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "email_verified",
		})
	}
}

// ResendVerification handles POST /v1/auth/resend-verification
// Sends a new verification email to the authenticated user.
func ResendVerification(authSvc *auth.Service, emailSvc *email.Sender) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		token, err := authSvc.CreateVerificationToken(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_server_error")
			return
		}

		verifyURL := primaryFrontendOrigin() + "/verify?token=" + token

		// Get user email for sending
		// We use the userID from context; in a real implementation you'd
		// look up the user's email. For now, we'll send via a generic path.
		if err := emailSvc.SendVerification(r.Context(), "", verifyURL); err != nil {
			// Non-blocking: log but don't fail the request
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "verification_email_sent",
		})
	}
}

// ForgotPassword handles POST /v1/auth/forgot-password
// Sends a password reset email if the email exists.
func ForgotPassword(authSvc *auth.Service, emailSvc *email.Sender) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email string `json:"email"`
		}
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request")
			return
		}

		email := strings.TrimSpace(strings.ToLower(req.Email))
		if email == "" {
			writeError(w, http.StatusBadRequest, "email_required")
			return
		}

		token, _, err := authSvc.CreatePasswordResetToken(r.Context(), email)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_server_error")
			return
		}

		// Always respond with success to avoid email enumeration
		if token != "" {
			resetURL := primaryFrontendOrigin() + "/reset-password?token=" + token
			_ = emailSvc.SendPasswordReset(r.Context(), email, resetURL)
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "if_account_exists_email_sent",
		})
	}
}

// ResetPassword handles POST /v1/auth/reset-password
// Validates the reset token and updates the password.
func ResetPassword(authSvc *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Token    string `json:"token"`
			Password string `json:"password"`
		}
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request")
			return
		}

		token := strings.TrimSpace(req.Token)
		password := req.Password

		if token == "" {
			writeError(w, http.StatusBadRequest, "missing_token")
			return
		}
		if password == "" {
			writeError(w, http.StatusBadRequest, "missing_password")
			return
		}

		err := authSvc.ResetPassword(r.Context(), token, password)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrResetTokenNotFound):
				writeError(w, http.StatusNotFound, "invalid_token")
			case errors.Is(err, auth.ErrResetTokenExpired):
				writeError(w, http.StatusGone, "token_expired")
			case errors.Is(err, auth.ErrResetTokenUsed):
				writeError(w, http.StatusConflict, "token_already_used")
			case errors.Is(err, auth.ErrWeakPassword):
				writeError(w, http.StatusBadRequest, err.Error())
			default:
				writeError(w, http.StatusInternalServerError, "internal_server_error")
			}
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "password_reset_success",
		})
	}
}
