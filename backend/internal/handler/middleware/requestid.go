package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"unicode"

	"shieldvn-backend/internal/observability"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	headerRequestID = "X-Request-ID"
	maxRequestIDLen = 128
)

// RequestID returns Gin middleware that reads or mints a request ID,
// stores it in the context, echoes it on the response header, and
// stamps it onto the active OTel span (if any).
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := sanitizeRequestID(c.GetHeader(headerRequestID))
		if id == "" {
			id = generateRequestID()
		}

		// Store in context via observability helpers.
		ctx := observability.WithRequestID(c.Request.Context(), id)
		c.Request = c.Request.WithContext(ctx)

		// Echo on response header.
		c.Header(headerRequestID, id)

		// Stamp onto active OTel span if tracing is wired.
		if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
			span.SetAttributes(attribute.String("request_id", id))
		}

		c.Next()
	}
}

// sanitizeRequestID enforces length and printable-ASCII constraints.
// Returns empty string if the input is entirely invalid.
func sanitizeRequestID(raw string) string {
	if raw == "" {
		return ""
	}

	// Truncate to max length.
	if len(raw) > maxRequestIDLen {
		raw = raw[:maxRequestIDLen]
	}

	// Strip non-printable and non-ASCII characters.
	clean := make([]byte, 0, len(raw))
	for i := 0; i < len(raw); i++ {
		r := rune(raw[i])
		if r > unicode.MaxASCII || !unicode.IsPrint(r) {
			continue
		}
		clean = append(clean, raw[i])
	}

	if len(clean) == 0 {
		return ""
	}
	return string(clean)
}

// generateRequestID produces a 32-char hex string (16 random bytes)
// using crypto/rand for unpredictability.
func generateRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback: should never happen with crypto/rand.
		return "fallback-request-id"
	}
	return hex.EncodeToString(b)
}

// isHealthzRequest is a helper to skip middleware for health checks if needed.
func isHealthzRequest(r *http.Request) bool {
	return r.URL.Path == "/healthz"
}
