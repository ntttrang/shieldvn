package observability

import (
	"context"
	"fmt"
	"os"
	"testing"

	"go.opentelemetry.io/otel"
)

func TestOnCloudRun_False_LocalDev(t *testing.T) {
	// Ensure K_REVISION is not set in test environment.
	os.Unsetenv("K_REVISION")
	if OnCloudRun() {
		t.Error("OnCloudRun() should be false without K_REVISION")
	}
}

func TestInitTelemetry_NoopWithoutCloudRun(t *testing.T) {
	// Clear env to simulate local dev.
	os.Unsetenv("K_REVISION")
	os.Unsetenv("GOOGLE_CLOUD_PROJECT")

	ctx := context.Background()
	shutdown, err := InitTelemetry(ctx, TelemetryConfig{
		GCPProjectID: "",
		ServiceName:  "test-service",
	})
	if err != nil {
		t.Fatalf("InitTelemetry should not error for no-op path: %v", err)
	}

	// The global tracer provider should be a noop type.
	tp := otel.GetTracerProvider()
	typeName := fmt.Sprintf("%T", tp)
	if typeName != "noop.TracerProvider" {
		t.Errorf("expected noop.TracerProvider, got %s", typeName)
	}

	// Shutdown should succeed without hanging.
	if err := shutdown(ctx); err != nil {
		t.Errorf("shutdown should not error: %v", err)
	}
}

func TestInitTelemetry_NoopWhenProjectSetButNotOnCloudRun(t *testing.T) {
	os.Unsetenv("K_REVISION")
	os.Setenv("GOOGLE_CLOUD_PROJECT", "test-project")
	defer os.Unsetenv("GOOGLE_CLOUD_PROJECT")

	ctx := context.Background()
	shutdown, err := InitTelemetry(ctx, TelemetryConfig{
		GCPProjectID: "test-project",
		ServiceName:  "test-service",
	})
	if err != nil {
		t.Fatalf("InitTelemetry should not error: %v", err)
	}

	// Still no-op because not on Cloud Run.
	tp := otel.GetTracerProvider()
	typeName := fmt.Sprintf("%T", tp)
	if typeName != "noop.TracerProvider" {
		t.Errorf("expected noop.TracerProvider, got %s", typeName)
	}

	if err := shutdown(ctx); err != nil {
		t.Errorf("shutdown should not error: %v", err)
	}
}

func TestInstrumentedHTTPClient_NotNil(t *testing.T) {
	client := InstrumentedHTTPClient()
	if client == nil {
		t.Fatal("InstrumentedHTTPClient returned nil")
	}
	if client.Transport == nil {
		t.Fatal("expected instrumented transport, got nil")
	}
}
