package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"shieldvn-backend/internal/observability"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		// Echo the request ID from context to verify it was stored.
		id := observability.RequestIDFrom(c.Request.Context())
		c.JSON(http.StatusOK, gin.H{"request_id": id})
	})
	return r
}

func TestRequestID_GeneratedWhenAbsent(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	id := w.Header().Get("X-Request-ID")
	if id == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}
	// Generated IDs are 32 hex chars (16 bytes).
	if len(id) != 32 {
		t.Errorf("expected 32-char hex ID, got %d chars: %s", len(id), id)
	}
}

func TestRequestID_ClientIDHonored(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "abc-123")
	router.ServeHTTP(w, req)

	id := w.Header().Get("X-Request-ID")
	if id != "abc-123" {
		t.Errorf("expected echoed client ID 'abc-123', got %q", id)
	}
}

func TestRequestID_OversizedTruncated(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	longID := strings.Repeat("a", 200)
	req.Header.Set("X-Request-ID", longID)
	router.ServeHTTP(w, req)

	id := w.Header().Get("X-Request-ID")
	if len(id) > maxRequestIDLen {
		t.Errorf("expected ID truncated to %d, got %d", maxRequestIDLen, len(id))
	}
	if len(id) != maxRequestIDLen {
		t.Errorf("expected exactly %d chars after truncation, got %d", maxRequestIDLen, len(id))
	}
}

func TestRequestID_NonPrintableStripped(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// Inject control characters.
	req.Header.Set("X-Request-ID", "abc\x00\x01\x02def")
	router.ServeHTTP(w, req)

	id := w.Header().Get("X-Request-ID")
	if id != "abcdef" {
		t.Errorf("expected 'abcdef' after stripping control chars, got %q", id)
	}
}

func TestRequestID_AllInvalid_GeneratesNew(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// All non-printable.
	req.Header.Set("X-Request-ID", "\x00\x01\x02")
	router.ServeHTTP(w, req)

	id := w.Header().Get("X-Request-ID")
	if id == "" {
		t.Fatal("expected generated ID when all chars stripped")
	}
	if len(id) != 32 {
		t.Errorf("expected 32-char generated ID, got %d chars: %s", len(id), id)
	}
}

func TestRequestID_ContextMatches(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "ctx-test-456")
	router.ServeHTTP(w, req)

	headerID := w.Header().Get("X-Request-ID")
	// The handler echoes the context value in the response body.
	body := w.Body.String()
	if !strings.Contains(body, "ctx-test-456") {
		t.Errorf("context request_id should match header: header=%s body=%s", headerID, body)
	}
}

func TestSanitizeRequestID_Table(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"valid", "abc-123", "abc-123"},
		{"max length", strings.Repeat("x", 128), strings.Repeat("x", 128)},
		{"over max", strings.Repeat("x", 200), strings.Repeat("x", 128)},
		{"control chars", "abc\x00def", "abcdef"},
		{"all control", "\x00\x01\x02", ""},
		{"unicode stripped", "abc🎉def", "abcdef"},
		{"tab stripped", "abc\tdef", "abcdef"},
		{"spaces kept", "abc def", "abc def"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeRequestID(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeRequestID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
