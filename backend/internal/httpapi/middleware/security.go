package middleware

import "net/http"

// SecurityHeaders adds standard security headers to every response.
// These protect against clickjacking, MIME sniffing, and information leakage.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		h.Set("X-XSS-Protection", "1; mode=block")

		// HSTS: instruct browsers to only connect over HTTPS.
		// Render terminates TLS at the proxy so r.TLS may be nil even in production;
		// check the X-Forwarded-Proto header that Render (and most reverse proxies) set.
		proto := r.Header.Get("X-Forwarded-Proto")
		if r.TLS != nil || proto == "https" {
			h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		}

		next.ServeHTTP(w, r)
	})
}

// CSRFCheck protects mutating endpoints against Cross-Site Request Forgery.
//
// Strategy: require a custom header (X-Requested-With) on all non-safe methods.
// Browsers will not send custom headers on cross-origin simple requests, so a
// forged form POST from evil.com will be missing the header and get rejected.
//
// The CORS middleware must also list "X-Requested-With" in Access-Control-Allow-Headers
// so that the legitimate frontend can send it.
func CSRFCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Safe methods don't need CSRF protection.
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		if r.Header.Get("X-Requested-With") == "" {
			writeJSONError(w, http.StatusForbidden, "csrf_header_missing")
			return
		}

		next.ServeHTTP(w, r)
	})
}
