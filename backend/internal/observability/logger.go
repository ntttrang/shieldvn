package observability

import (
	"context"
	"log"
	"log/slog"
	"os"
	"strings"
)

// Well-known metadata keys that LogMeta accepts.
// PII (phone, bank account, national ID, prompt, image bytes) is
// structurally unreachable — only these keys can be passed.
var allowedMetaKeys = map[string]bool{
	"request_id":    true,
	"latency_ms":    true,
	"verdict_level": true,
	"error_type":    true,
	"tier":          true,
	"gemini_model":  true,
	"attempt":       true,
	"status":        true,
}

// Meta is a key-value pair for structured logging.
// Only well-known metadata keys are accepted; unknown keys are
// dropped with a warning. This prevents PII from reaching logs.
type Meta struct {
	Key   string
	Value any
}

// InitLogger configures the global slog default to a JSON handler
// writing to stdout at the given level. It also bridges stdlib log
// output through slog so that any remaining log.Printf calls produce
// structured JSON.
func InitLogger(level slog.Level) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Bridge stdlib log → slog so legacy log.Printf produces JSON.
	log.SetOutput(&slogWriter{logger: logger})
	log.SetFlags(0)

	return logger
}

// FromContext returns a logger pre-stamped with request_id and
// trace_id from the context. Safe to call with an empty context —
// returns the default logger with no extra attributes.
func FromContext(ctx context.Context) *slog.Logger {
	logger := slog.Default()
	if id := RequestIDFrom(ctx); id != "" {
		logger = logger.With("request_id", id)
	}
	if id := TraceIDFrom(ctx); id != "" {
		logger = logger.With("trace_id", id)
	}
	return logger
}

// LogMeta emits a structured log line with only allowed metadata keys.
// Unknown keys are silently dropped with a warning to prevent PII leaks.
func LogMeta(ctx context.Context, level slog.Level, msg string, fields ...Meta) {
	logger := FromContext(ctx)
	var attrs []any
	for _, f := range fields {
		if !allowedMetaKeys[f.Key] {
			slog.Warn("dropping unknown metadata key from log",
				"dropped_key", f.Key,
			)
			continue
		}
		attrs = append(attrs, f.Key, f.Value)
	}
	logger.Log(ctx, level, msg, attrs...)
}

// ParseLevel converts a string level name to slog.Level.
// Defaults to info on unrecognized input.
func ParseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// slogWriter adapts slog.Logger to io.Writer for stdlib log bridging.
type slogWriter struct {
	logger *slog.Logger
}

func (w *slogWriter) Write(p []byte) (n int, err error) {
	// Strip trailing newline that stdlib log appends.
	msg := strings.TrimRight(string(p), "\n")
	w.logger.Info(msg, "source", "stdlib-log")
	return len(p), nil
}
