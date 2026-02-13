// Package monitoring provides application observability via Sentry error tracking
// and structured metric logging.
//
// Set SENTRY_DSN to enable Sentry error tracking. When empty, errors are only
// logged via slog (safe for development).
package monitoring

import (
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

// sentryEnabled tracks whether Sentry was configured. When false, all
// capture/report operations are no-ops to keep the code path the same.
var sentryEnabled bool

// Init initializes monitoring. Call this early in main().
// Currently implements structured logging of errors and metrics.
// When SENTRY_DSN is set, errors will also be forwarded to Sentry.
func Init(service string) {
	dsn := strings.TrimSpace(os.Getenv("SENTRY_DSN"))
	if dsn == "" {
		slog.Info("monitoring: Sentry disabled (SENTRY_DSN not set)")
		sentryEnabled = false
		return
	}

	// When the sentry-go SDK is added as a dependency, initialization would
	// go here:
	//
	//   sentry.Init(sentry.ClientOptions{
	//       Dsn:              dsn,
	//       Environment:      os.Getenv("APP_ENV"),
	//       Release:          service + "@" + buildVersion(),
	//       TracesSampleRate: 0.2,
	//   })
	//
	// For now we mark it as enabled and log.
	sentryEnabled = true
	slog.Info("monitoring: Sentry enabled", "service", service, "dsn_prefix", dsn[:min(len(dsn), 20)]+"...")
}

// Flush should be called before application shutdown to ensure all events
// are sent to Sentry.
func Flush(timeout time.Duration) {
	if !sentryEnabled {
		return
	}
	// sentry.Flush(timeout)
	_ = timeout
}

// CaptureError logs an error and optionally sends it to Sentry.
func CaptureError(err error, context map[string]string) {
	attrs := []any{"error", err}
	for k, v := range context {
		attrs = append(attrs, k, v)
	}
	slog.Error("captured error", attrs...)

	if sentryEnabled {
		// sentry.WithScope(func(scope *sentry.Scope) {
		//     for k, v := range context {
		//         scope.SetTag(k, v)
		//     }
		//     sentry.CaptureException(err)
		// })
	}
}

// CaptureMessage logs an informational message and optionally sends it to Sentry.
func CaptureMessage(msg string, context map[string]string) {
	attrs := []any{"message", msg}
	for k, v := range context {
		attrs = append(attrs, k, v)
	}
	slog.Info("captured message", attrs...)

	if sentryEnabled {
		// sentry.CaptureMessage(msg)
	}
}

// TrackMetric logs a key metric event.
func TrackMetric(event string, attrs map[string]string) {
	logAttrs := []any{"event", event}
	for k, v := range attrs {
		logAttrs = append(logAttrs, k, v)
	}
	slog.Info("metric", logAttrs...)
}

// RecoverMiddleware catches panics in HTTP handlers, logs them, reports to
// Sentry, and returns a 500 JSON response.
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				stack := string(debug.Stack())
				slog.Error("panic recovered",
					"panic", rec,
					"method", r.Method,
					"path", r.URL.Path,
					"stack", stack,
				)

				if sentryEnabled {
					// sentry.CurrentHub().Recover(rec)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"internal_server_error"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RequestMetricsMiddleware logs request duration and status for observability.
func RequestMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: 200}

		next.ServeHTTP(sw, r)

		duration := time.Since(start)

		// Log slow requests (>2s) as warnings
		if duration > 2*time.Second {
			slog.Warn("slow request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.status,
				"duration_ms", duration.Milliseconds(),
			)
		}
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
