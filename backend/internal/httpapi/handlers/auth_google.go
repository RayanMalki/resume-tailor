package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"resume-tailor/internal/auth"
	"resume-tailor/internal/httpapi/cookies"
	"resume-tailor/internal/httpapi/middleware"
	"resume-tailor/internal/notify"
)

const (
	googleAuthURL  = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL = "https://oauth2.googleapis.com/token"
	googleUserURL  = "https://openidconnect.googleapis.com/v1/userinfo"
)

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
}

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
}

func GoogleStart(authSvc *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID"))
		redirectURL := strings.TrimSpace(os.Getenv("GOOGLE_REDIRECT_URL"))
		if clientID == "" || redirectURL == "" {
			writeError(w, http.StatusNotImplemented, "google_oauth_not_configured")
			return
		}

		state, err := auth.NewToken()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "oauth_state_error")
			return
		}

		expiresAt := time.Now().Add(10 * time.Minute)
		cookies.SetOAuthStateCookie(w, state, expiresAt)

		if redirect := sanitizeRedirect(r.URL.Query().Get("redirect")); redirect != "" {
			cookies.SetOAuthRedirectCookie(w, redirect, expiresAt)
		}

		params := url.Values{}
		params.Set("client_id", clientID)
		params.Set("redirect_uri", redirectURL)
		params.Set("response_type", "code")
		params.Set("scope", "openid email profile")
		params.Set("state", state)
		params.Set("prompt", "select_account")

		http.Redirect(w, r, googleAuthURL+"?"+params.Encode(), http.StatusFound)
	}
}

func GoogleCallback(authSvc *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID"))
		clientSecret := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET"))
		redirectURL := strings.TrimSpace(os.Getenv("GOOGLE_REDIRECT_URL"))
		frontendOrigin := strings.TrimSpace(os.Getenv("FRONTEND_ORIGIN"))
		if clientID == "" || clientSecret == "" || redirectURL == "" {
			writeError(w, http.StatusNotImplemented, "google_oauth_not_configured")
			return
		}

		state := r.URL.Query().Get("state")
		code := r.URL.Query().Get("code")
		if state == "" || code == "" {
			writeError(w, http.StatusBadRequest, "oauth_missing_params")
			return
		}

		cookieState, ok := cookies.ReadOAuthStateCookie(r)
		cookies.ClearOAuthStateCookie(w)
		if !ok || cookieState != state {
			writeError(w, http.StatusBadRequest, "oauth_state_mismatch")
			return
		}

		tokenResp, err := exchangeGoogleToken(r.Context(), clientID, clientSecret, redirectURL, code)
		if err != nil {
			writeError(w, http.StatusBadRequest, "oauth_token_exchange_failed")
			return
		}

		userInfo, err := fetchGoogleUser(r.Context(), tokenResp.AccessToken)
		if err != nil {
			writeError(w, http.StatusBadRequest, "oauth_userinfo_failed")
			return
		}

		if userInfo.Sub == "" || userInfo.Email == "" || !userInfo.EmailVerified {
			writeError(w, http.StatusBadRequest, "oauth_invalid_userinfo")
			return
		}

		token, expiresAt, userID, err := authSvc.LoginWithOAuth(r.Context(), "google", userInfo.Sub, userInfo.Email, userInfo.Name)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidCredentials) {
				writeError(w, http.StatusUnauthorized, "oauth_login_failed")
				return
			}
			writeError(w, http.StatusInternalServerError, "oauth_login_failed")
			return
		}

		cookies.SetSessionCookie(w, token, expiresAt)

		go func() {
			_ = notify.SendEvent(context.Background(), notify.Event{
				Type:      "login_success",
				Path:      r.URL.Path,
				Method:    r.Method,
				Status:    http.StatusFound,
				IP:        middleware.ClientIP(r),
				UserAgent: r.UserAgent(),
				UserID:    userID.String(),
				Meta:      map[string]string{"provider": "google"},
				TS:        time.Now().UTC(),
			})
		}()

		redirect := "/dashboard"
		if stored, ok := cookies.ReadOAuthRedirectCookie(r); ok {
			cookies.ClearOAuthRedirectCookie(w)
			if safe := sanitizeRedirect(stored); safe != "" {
				redirect = safe
			}
		}

		if frontendOrigin == "" {
			frontendOrigin = "http://localhost:3000"
		}
		http.Redirect(w, r, strings.TrimRight(frontendOrigin, "/")+redirect, http.StatusFound)
	}
}

func exchangeGoogleToken(ctx context.Context, clientID, clientSecret, redirectURL, code string) (googleTokenResponse, error) {
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("redirect_uri", redirectURL)
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, googleTokenURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return googleTokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 6 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return googleTokenResponse{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return googleTokenResponse{}, fmt.Errorf("token status %d", resp.StatusCode)
	}

	var parsed googleTokenResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return googleTokenResponse{}, err
	}
	if parsed.AccessToken == "" {
		return googleTokenResponse{}, fmt.Errorf("missing access token")
	}
	return parsed, nil
}

func fetchGoogleUser(ctx context.Context, accessToken string) (googleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserURL, nil)
	if err != nil {
		return googleUserInfo{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 6 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return googleUserInfo{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return googleUserInfo{}, fmt.Errorf("userinfo status %d", resp.StatusCode)
	}

	var info googleUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return googleUserInfo{}, err
	}
	return info, nil
}

func sanitizeRedirect(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return ""
	}
	if !strings.HasPrefix(raw, "/") {
		return ""
	}
	return raw
}
