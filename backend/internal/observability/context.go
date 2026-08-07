package observability

import "context"

// Unexported key types prevent collisions with other packages.
type ctxKeyRequestID struct{}
type ctxKeyTraceID struct{}

// WithRequestID stores a request ID in the context.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKeyRequestID{}, id)
}

// RequestIDFrom retrieves the request ID from the context.
// Returns empty string if not set.
func RequestIDFrom(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyRequestID{}).(string); ok {
		return v
	}
	return ""
}

// WithTraceID stores a trace ID in the context.
func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKeyTraceID{}, id)
}

// TraceIDFrom retrieves the trace ID from the context.
// Returns empty string if not set.
func TraceIDFrom(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyTraceID{}).(string); ok {
		return v
	}
	return ""
}
