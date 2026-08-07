package observability

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	gcpdetector "go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	gcptraceexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
)

// TelemetryConfig holds config needed by the telemetry init.
type TelemetryConfig struct {
	GCPProjectID string
	ServiceName  string
}

// shutdownFuncs collects provider shutdown functions.
type shutdownFuncs struct {
	fns []func(context.Context) error
}

func (s *shutdownFuncs) add(fn func(context.Context) error) {
	s.fns = append(s.fns, fn)
}

func (s *shutdownFuncs) shutdown(ctx context.Context) error {
	var firstErr error
	for _, fn := range s.fns {
		if err := fn(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// OnCloudRun returns true when the process is running on Cloud Run.
// Cloud Run sets K_REVISION for every container instance.
func OnCloudRun() bool {
	return os.Getenv("K_REVISION") != ""
}

// InitTelemetry sets up OTel trace and metric providers.
// On Cloud Run with a GCP project ID, it creates real GCP exporters.
// Otherwise it registers no-op providers — zero network, zero errors,
// so dev/CI needs no GCP credentials.
func InitTelemetry(ctx context.Context, cfg TelemetryConfig) (shutdown func(context.Context) error, err error) {
	var sd shutdownFuncs

	// Set up W3C Trace Context propagation.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if OnCloudRun() && cfg.GCPProjectID != "" {
		slog.Info("initializing GCP OTel exporters",
			"project", cfg.GCPProjectID,
			"revision", os.Getenv("K_REVISION"),
		)
		return initGCPProviders(ctx, cfg, &sd)
	}

	slog.Info("initializing no-op OTel providers (not on Cloud Run or no GCP project)")
	return initNoopProviders(&sd)
}

// initGCPProviders wires real GCP trace + metric exporters.
func initGCPProviders(ctx context.Context, cfg TelemetryConfig, sd *shutdownFuncs) (func(context.Context) error, error) {
	// Build resource with GCP-detected attributes.
	res, err := resource.New(ctx,
		resource.WithDetectors(gcpdetector.NewDetector()),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
		),
		resource.WithTelemetrySDK(),
	)
	if err != nil {
		return sd.shutdown, fmt.Errorf("build otel resource: %w", err)
	}

	// Trace exporter → Cloud Trace.
	traceExp, err := gcptraceexporter.New(gcptraceexporter.WithProjectID(cfg.GCPProjectID))
	if err != nil {
		return sd.shutdown, fmt.Errorf("create GCP trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
	)
	sd.add(tp.Shutdown)
	otel.SetTracerProvider(tp)

	// Metric provider — register with no reader for now.
	// Custom metrics (shieldvn.verdicts, etc.) are added in feature phases.
	// The GCP metric exporter will be wired when metrics are defined.
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
	)
	sd.add(mp.Shutdown)
	otel.SetMeterProvider(mp)

	return sd.shutdown, nil
}

// initNoopProviders registers providers that do nothing — no network,
// no errors, safe for dev/CI without GCP credentials.
func initNoopProviders(sd *shutdownFuncs) (func(context.Context) error, error) {
	otel.SetTracerProvider(tracenoop.NewTracerProvider())
	// No meter provider override needed — the global default is already no-op.
	return sd.shutdown, nil
}

// InstrumentedHTTPClient returns an *http.Client whose transport is
// wrapped with OpenTelemetry instrumentation. Each outbound HTTP call
// becomes a child span of the active trace. Use this for the Gemini
// genai SDK client.
func InstrumentedHTTPClient() *http.Client {
	return &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
}

// Tracer returns a named tracer from the global provider.
// Feature phases use this to create custom spans:
//
//	ctx, span := observability.Tracer("shieldvn").Start(ctx, "firestore.lookup")
//	defer span.End()
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
