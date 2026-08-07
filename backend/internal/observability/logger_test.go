package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

// captureOutput runs fn while capturing slog output to a buffer.
// It temporarily sets the default logger to write JSON to the buffer.
func captureOutput(level slog.Level, fn func()) string {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: level})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	fn()
	return buf.String()
}

func TestInitLogger_SetsJSONHandler(t *testing.T) {
	output := captureOutput(slog.LevelInfo, func() {
		InitLogger(slog.LevelInfo)
		slog.Info("test boot message")
	})

	// Must be valid JSON.
	var m map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &m); err != nil {
		t.Fatalf("log output is not valid JSON: %v\nraw: %s", err, output)
	}
	if m["msg"] != "test boot message" {
		t.Errorf("expected msg 'test boot message', got %v", m["msg"])
	}
	if m["level"] != "INFO" {
		t.Errorf("expected level INFO, got %v", m["level"])
	}
}

func TestInitLogger_RespectsLevel(t *testing.T) {
	output := captureOutput(slog.LevelWarn, func() {
		InitLogger(slog.LevelWarn)
		slog.Info("should be suppressed")
		slog.Warn("should appear")
	})

	if strings.Contains(output, "should be suppressed") {
		t.Error("info message should be suppressed at warn level")
	}
	if !strings.Contains(output, "should appear") {
		t.Error("warn message should appear at warn level")
	}
}

func TestFromContext_EmptyContext_NoPanic(t *testing.T) {
	// Must not panic with an empty context.
	InitLogger(slog.LevelInfo)
	logger := FromContext(context.Background())
	if logger == nil {
		t.Fatal("FromContext returned nil")
	}
}

func TestFromContext_WithRequestID(t *testing.T) {
	// Set up a logger that writes to our buffer.
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))

	ctx := WithRequestID(context.Background(), "test-req-123")
	FromContext(ctx).Info("with request id")

	output := buf.String()
	if !strings.Contains(output, "test-req-123") {
		t.Errorf("expected request_id in log output, got: %s", output)
	}
}

func TestLogMeta_AcceptsKnownKeys(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))

	knownKeys := []string{
		"request_id", "latency_ms", "verdict_level",
		"error_type", "tier", "gemini_model", "attempt", "status",
	}
	ctx := context.Background()

	for _, k := range knownKeys {
		LogMeta(ctx, slog.LevelInfo, "test", Meta{Key: k, Value: "val"})
	}

	output := buf.String()
	for _, k := range knownKeys {
		if !strings.Contains(output, k) {
			t.Errorf("known key %q not found in log output", k)
		}
	}
}

func TestLogMeta_DropsUnknownKeys(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))

	ctx := context.Background()

	// "phone_number" is PII/unknown → must be dropped.
	LogMeta(ctx, slog.LevelInfo, "pii-attempt",
		Meta{Key: "phone_number", Value: "0901234567"},
		Meta{Key: "status", Value: "ok"},
	)

	output := buf.String()

	// Parse each JSON line.
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			continue
		}
		if m["msg"] == "pii-attempt" {
			if _, found := m["phone_number"]; found {
				t.Error("PII key 'phone_number' was NOT dropped from log line")
			}
			if m["status"] != "ok" {
				t.Error("known key 'status' was incorrectly dropped")
			}
		}
	}
}

func TestLogMeta_WarnsOnUnknownKey(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})
	slog.SetDefault(slog.New(handler))

	ctx := context.Background()
	LogMeta(ctx, slog.LevelWarn, "drop-test",
		Meta{Key: "bank_account", Value: "9876543210"},
	)

	output := buf.String()
	if !strings.Contains(output, "dropping unknown metadata key") {
		t.Errorf("expected warning about dropped key, got: %s", output)
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"unknown", slog.LevelInfo},
		{"", slog.LevelInfo},
	}
	for _, tt := range tests {
		got := ParseLevel(tt.input)
		if got != tt.want {
			t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
