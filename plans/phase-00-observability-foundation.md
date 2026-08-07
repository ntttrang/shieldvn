---
phase: 0
title: "Observability Foundation (slog + request-id + OTel → GCP)"
status: completed
priority: P2
effort: "10h (~1.5d)"
dependencies: []
branch: feat/phase-02-image-verdict-gemini-vision
---

# Phase 0: Observability Foundation

> Formerly the standalone plan `260801-1520-GH-02-observability-logging-tracing-foundation`. Absorbed here as the single plan of record. Source of truth (do not re-debate): [`plans/reports/brainstorm-260801-1520-GH-02-observability-logging-tracing-cloud-run-report.md`](../reports/brainstorm-260801-1520-GH-02-observability-logging-tracing-cloud-run-report.md).

## Overview

Foundation layer of ShieldVN's production-ready observability, scoped to **the plumbing only** — what must land so later feature phases just call into it. Free + Cloud Run-native: structured JSON logs (stdlib `slog`) → Cloud Logging (auto-ingested from stdout); distributed traces via OpenTelemetry → Cloud Trace; custom metrics → Cloud Monitoring. No paid third-party tool. No agent/collector.

**Scope of THIS phase = plumbing:**
- `slog` structured logger (replaces ad-hoc `log.Printf`).
- Request-ID Gin middleware (mint/propagate → context + log + trace + response header).
- OTel tracer/meter providers + GCP exporters, with **no-op fallback for dev/CI** (no GCP creds needed locally).
- Tests + docs.

**Out of scope (tracked in feature phases):**
- Per-call instrumentation (Gemini span, Firestore span, verdict metrics) → added **inline** as feature phases 2–8 are built (see [Instrumentation map](#instrumentation-map) below).
- Dashboards, alert policies, SLOs, prod verification → [Phase 9 (deploy & demo)](./phase-09-deploy-demo.md).
- Frontend client-side error tracking → deferred (YAGNI).

## Locked decisions (do not re-debate)

- **Tooling = all-native GCP via OpenTelemetry.** No Datadog/Grafana/Sentry.
- **Timing = hybrid.** Foundation now; instrument-as-you-build (phases 2–8); ops surface at phase 9.
- **`slog`** (stdlib, Go 1.21+; repo is Go 1.25) — zero new logging dependency.
- **No-op exporter fallback** off-Cloud-Run so dev/CI need no GCP credentials.
- **Privacy is structural:** a logging helper accepts metadata keys only (`request_id, latency_ms, verdict_level, error_type, tier, gemini_model`). PII (phone/bank account number/national ID/prompt/image bytes) cannot be passed.

## Architecture (data flow)

```
Request ──► Gin
   ├─ otelgin.Middleware          → creates root span (Cloud Trace)
   ├─ requestid.Middleware        → X-Request-ID: read | mint → ctx + slog attr + span attr + resp header
   └─ handler/service
        └─ slog.With(request_id)  → JSON line to stdout ──► Cloud Logging (auto)
            └─ tracer.Start(span) → child spans (Gemini/Firestore, added in feature phases) ──► Cloud Trace
```

Cloud Run auto-correlates a request's log lines ↔ trace when the trace ID is emitted in the structured log (`logging.googleapis.com/trace`). Cloud Run detected via `K_REVISION` env + `GOOGLE_CLOUD_PROJECT`.

## Privacy & security guardrail

- Never log PII. The logger helper takes a fixed set of metadata keys; PII by construction cannot reach logs.
- Cloud Trace/metrics carry metadata only (verdict level, latency, error type) — no prompt, no image bytes, no extracted entities.
- Secrets (`GEMINI_API_KEY`, Firestore svc account) already excluded from logs by existing config comment; preserved.

## Build steps

Build order: **1 → 2 → 3 → 4**. Step 3 wires otelgin/otelhttp which consume the request-ID from step 2's context; step 4 verifies the whole chain.

---

### Step 1 — Structured Logging (slog) · 2h

Replace ad-hoc `log.Printf` boot/infra lines with stdlib `slog` JSON → stdout (Cloud Run auto-ingests → Cloud Logging). Privacy-safe helper stamps `request_id`/`trace_id` and accepts **metadata keys only**. No new dependency.

**Requirements:** process boots with configured `slog` JSON logger; level via `LOG_LEVEL` (default `info`); helper returns logger pre-stamped with request/trace IDs from context. JSON single-line to stdout; zero new module deps; existing key-less boot for CI/`/healthz` preserved.

**Architecture:** `config.Config` gains `LogLevel` + `GCPProjectID` (read in step 3). New `observability` package owns logger setup. `slog.SetDefault` routes everything through the JSON handler so existing `log.Printf` (Gemini service) is captured too via `slog.NewLogLogger`. Step 1 keeps Gemini service `log.Printf` intact (migrated during feature instrumentation) — only boot/infra lines move now.

**Related code files:**
- Create: `backend/internal/observability/logger.go`
- Create: `backend/internal/observability/context.go` (request-id/trace-id context helpers)
- Modify: `backend/internal/config/config.go` (+`LogLevel`, +`GCPProjectID` getters)
- Modify: `backend/cmd/api/main.go` (`observability.InitLogger(cfg)`; replace boot `log.Printf` lines)
- Reference: `backend/internal/service/gemini_service.go` (existing `log.Printf` — untouched this step)

**Implementation steps:**
1. `config.go`: add `LogLevel` (from `LOG_LEVEL`, default `info`) + `GCPProjectID` (from `GOOGLE_CLOUD_PROJECT`); parse level via `slog.Level`.
2. `observability/logger.go`:
   - `InitLogger(level string) *slog.Logger` — `slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parseLevel(level)})`; `slog.SetDefault(logger)`.
   - `FromContext(ctx) *slog.Logger` — default logger with `request_id` + `trace_id` attrs from context (empty-safe). `trace_id` added in step 3; `request_id` in step 2.
   - `LogMeta(ctx, level, msg, fields ...Meta)` — privacy gate: `type Meta struct{ Key string; Value any }`; accepts only well-known keys (`request_id, latency_ms, verdict_level, error_type, tier, gemini_model, attempt, status`); unknown keys dropped + warned. **PII structurally unreachable.**
3. `observability/context.go`: `WithRequestID(ctx, id)`, `RequestIDFrom(ctx) string`, unexported key type. (Trace-ID getter added in step 3.)
4. `main.go`: call `observability.InitLogger(cfg.LogLevel)` before building the analyzer; replace the boot `log.Printf`/`log.Fatalf` lines with `slog`/`slog.Error`. Preserve graceful shutdown.
5. `go build ./...` + `go vet ./...` clean.

**Success criteria:**
- [ ] `LOG_LEVEL=debug go run ./cmd/api` boots; boot line is single-line JSON with `level`, `msg`, `time`.
- [ ] `FromContext(ctx)` with no request-id returns default logger (no panic).
- [ ] `LogMeta` drops + warns on unknown key; accepts all documented metadata keys.
- [ ] No PII string can be passed to `LogMeta` (compile-time: only `Meta` accepted).
- [ ] `go build ./... && go vet ./...` pass; existing tests green.

**Risks:** stdlib `log` bridging changes `log.Printf` output to JSON (desired — Cloud Logging-friendly); verify readability in step 4. Level parse errors default to `info` + warn.

---

### Step 2 — Request-ID Middleware · 2h

Gin middleware: read inbound `X-Request-ID` (or mint a UUID) per request, store in context (step 1's helpers), echo on response header, stamp every log line + the OTel span (step 3). Lets a single user-reported failure be traced end-to-end in Cloud Logging. (This is the "request id" `docs/07 §8` promised but never implemented.)

**Requirements:** every response carries `X-Request-ID`; client-supplied id honored (max 128 chars, ASCII-filtered); otherwise generated; retrievable via `observability.RequestIDFrom(ctx)`. ID generation via `crypto/rand` hex (not `math/rand` — not crypto-stable + banned in this harness); negligible overhead.

**Architecture:** middleware registered **before** route handlers and alongside (after, in chain order) the future otelgin middleware so the span attribute is set on the active span. Stores ID via step 1's `WithRequestID`. Handler calls `observability.FromContext(c.Request.Context())` to log with the ID attached.

**Related code files:**
- Create: `backend/internal/handler/middleware/requestid.go`
- Modify: `backend/cmd/api/main.go` (register `middleware.RequestID()` in the router chain)
- Modify: `backend/internal/handler/analyze_handler.go` (optional one-line: `observability.FromContext` for the `analyze failed` line)
- Reference: `backend/internal/observability/context.go` (from step 1)

**Implementation steps:**
1. `handler/middleware/requestid.go`: `func RequestID() gin.HandlerFunc` — read `X-Request-ID`; sanitize (length ≤128, strip non-printable); if empty, generate 16 bytes hex via `crypto/rand`. Store with `observability.WithRequestID`; set `c.Header("X-Request-ID", id)`. Stamp onto active OTel span if present (`trace.SpanFromContext`, attr `request_id`) — guarded no-op before step 3 wires tracing.
2. `main.go`: `router.Use(middleware.RequestID())` in the chain (after cors, before route group).
3. Handler: in `Analyze`, replace `log.Printf("analyze failed: %v", err)` with `observability.FromContext(c.Request.Context()).Error("analyze failed", "error", err)`. (Optional polish; keeps rate-limit logic intact.)
4. `go build ./...` + manual curl: confirm `X-Request-ID` echoed on `/healthz` and `/api/v1/analyze`; client-supplied id honored.

**Success criteria:**
- [ ] `curl -i localhost:8080/healthz` returns `X-Request-ID` header.
- [ ] `curl -H 'X-Request-ID: abc-123' ...` echoes `abc-123` back.
- [ ] Empty/oversized/malformed client id replaced with fresh server id (no crash, no header injection).
- [ ] Failing `analyze` logs a JSON line containing the same `request_id` as the response header.
- [ ] `go build ./...` clean; existing tests green.

**Risks:** header/log injection — sanitize client id (length + printable ASCII) before storing/logging (in place). Middleware order — request-id before handlers that log (right after CORS).

---

### Step 3 — OpenTelemetry Providers & GCP Exporters · 4h

OTel tracer + meter providers exporting to **Cloud Trace** + **Cloud Monitoring** on Cloud Run; **no-op exporters** locally/in CI (no GCP creds). Wire `otelgin` (HTTP server spans) + `otelhttp` on the Gemini SDK's HTTP client (each model call = child span). This makes "full ops" real; business-logic spans/metrics added inline in feature phases.

**Requirements:** in prod, each request creates a Cloud Trace span; each Gemini round-trip nests as a child span; custom meters registered (consumed later). Locally/CI: providers init no-op, zero network, zero errors. Exporters created only when `K_REVISION` set **and** `GOOGLE_CLOUD_PROJECT` present; otherwise no-op. Graceful `Shutdown()` flushes on SIGTERM. No new third-party SaaS.

**Architecture:** new `observability/otel.go` builds a `trace.TracerProvider` + `metric.MeterProvider` with a GCP-detected `Resource` (`go.opentelemetry.io/contrib/detectors/gcp` → `service.name`, `cloud.provider=gcp`, `service.instance.id` = Cloud Run revision). Exporters: `GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace` + `/exporter/metric` (straight to GCP, no collector). The genai client accepts an `HTTPClient`; pass an `otelhttp`-instrumented `*http.Client` so model calls become spans.

**Related code files:**
- Create: `backend/internal/observability/otel.go` (Init/Shutdown + Cloud Run detection + instrumented HTTP client factory)
- Modify: `backend/cmd/api/main.go` (`observability.InitTelemetry(ctx, cfg)` before routes; `defer observability.Shutdown(shutdownCtx)`; register `otelgin.Middleware("shieldvn-api")`)
- Modify: `backend/internal/service/gemini_client.go` (inject instrumented `*http.Client` into `genai.ClientConfig.HTTPClient`)
- Modify: `backend/go.mod` (OTel + GCP exporter modules via `go get`)
- Reference: `backend/internal/observability/context.go` (request-id stamped onto span attribute)

**Implementation steps:**
1. `go get` latest stable: `otel`, `otel/sdk`, `otel/sdk/trace`, `otel/sdk/metric`, `otel/exporters/stdout/stdouttrace` (optional dev debug), `contrib/instrumentation/github.com/gin-gonic/gin/otelgin`, `contrib/instrumentation/net/http/otelhttp`, `contrib/detectors/gcp`, `GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace`, `.../exporter/metric`. `go mod tidy`.
2. `observability/otel.go`:
   - `func OnCloudRun() bool` — `os.Getenv("K_REVISION") != ""`.
   - `func InitTelemetry(ctx, cfg) (shutdown func(context.Context) error, err error)` — build resource via GCP detector; if `OnCloudRun() && cfg.GCPProjectID != ""`, create real trace + metric exporters; else register no-op `trace.NewNoopTracerProvider()` / no metric reader. Register providers globally (`otel.SetTracerProvider`, `otel.SetMeterProvider`).
   - `func InstrumentedHTTPClient() *http.Client` — `&http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}`.
   - `Shutdown(ctx)` — `tp.Shutdown(ctx)` + `mp.Shutdown(ctx)`, log flush errors.
3. `main.go`: call `InitTelemetry` right after `InitLogger`; `router.Use(otelgin.Middleware("shieldvn-api"))` (finalize order vs request-id in step 4 — if otelgin creates the span, register request-id **after** otelgin so `trace.SpanFromContext` is non-nil). In graceful-shutdown block call `observability.Shutdown(shutdownCtx)`.
4. `gemini_client.go`: `NewRealGeminiService` passes `HTTPClient: observability.InstrumentedHTTPClient()` into `&genai.ClientConfig{...}`. (Confirm field name via `go doc` during implementation.)
5. Local smoke: `go run ./cmd/api` (no creds) → boots clean, no exporter errors. `curl /healthz` → 200. Traces go nowhere locally (no-op) — expected.

**Success criteria:**
- [ ] App boots locally with **no** `GOOGLE_CLOUD_PROJECT` and **no** `K_REVISION` → no exporter errors, no network attempts, `/healthz` 200.
- [ ] `OnCloudRun()` is the single switch; real exporters only when both signals present.
- [ ] `otelgin.Middleware` registered; a request's trace context is created (`trace.SpanFromContext` non-nil in middleware).
- [ ] Gemini client uses the instrumented transport (genai calls appear as child spans in prod).
- [ ] `Shutdown()` invoked on SIGTERM and returns without hanging.
- [ ] `go build ./... && go vet ./...` clean; tests green.

**Risks:** genai `HTTPClient` field name/semantics — confirm via `go doc google.golang.org/genai.ClientConfig`; fallback: wrap at transport level only. Middleware ordering vs span attribute — finalize in step 4 test. Exporter cost/quota — no custom metrics created this step (all within free tier for hackathon scale). OTel provider setup is the most novel part; budget the 4h. No-op fallback is the safety net.

---

### Step 4 — Tests & Docs Update · 2h

Lock the foundation with tests (logger privacy gate, request-id sanitize/honor, no-op telemetry fallback, span attribute stamping), run the full suite, and update docs (`07-deployment-guide.md §8` now concrete + roadmap changelog). Consistency/verification gate before handing instrumentation off to feature phases.

**Requirements:** unit tests prove (a) `LogMeta` rejects PII/unknown keys, (b) request-id sanitized + honored + generated, (c) telemetry inits no-op without Cloud Run signals and attempts no network, (d) request-id stamped on active span when tracing wired. `go test ./...` green; no flaky/networked tests (all no-op exporters); docs reflect implemented reality.

**Architecture:** tests beside their packages (`observability/*_test.go`, `handler/middleware/*_test.go`). Telemetry tests assert on provider type (noop vs real) — no real GCP calls. `httptest`-based Gin test proves request-id middleware end-to-end.

**Related code files:**
- Create: `backend/internal/observability/logger_test.go`
- Create: `backend/internal/observability/otel_test.go`
- Create: `backend/internal/handler/middleware/requestid_test.go`
- Modify: `docs/07-deployment-guide.md` (§8 — concrete observability: what's logged/traced, env vars, no-op fallback, Cloud Trace/Monitoring links)
- Modify: `docs/06-project-roadmap.md` (changelog entry: observability foundation landed)

**Implementation steps:**
1. `logger_test.go`: assert `InitLogger` sets default to JSON handler at given level; `LogMeta` accepts each documented metadata key, drops + warns on unknown/PII-looking keys; `FromContext` returns default when empty.
2. `requestid_test.go`: table-test sanitize/honor/generate (oversize → trimmed/replaced; malformed → replaced; valid client id → honored). `gin.CreateTestContext` + `httptest`.
3. `otel_test.go`: with neither `K_REVISION` nor project set → `InitTelemetry` returns no-op providers, no exporter, no error; `OnCloudRun()` false. (Do not test real exporter path — needs creds; covered manually on deploy.)
4. Run `go test ./... -race -count=1`. Fix any failure (do not skip/paper over).
5. Docs: rewrite `07 §8` — structured JSON logs → Cloud Logging; traces → Cloud Trace via otelgin + otelhttp; env: `LOG_LEVEL`, `GOOGLE_CLOUD_PROJECT`; no-op fallback in dev/CI; PII never logged. Add roadmap changelog line.
6. Final consistency sweep: re-read this phase file; confirm no stale "aspirational" language; confirm file paths match what was created.

**Success criteria:**
- [ ] `go test ./... -race -count=1` green; all new tests pass.
- [ ] No test makes a real network call (no-op exporters only).
- [ ] Privacy: a test proves a PII string passed as a value to `LogMeta` is dropped — PII cannot be logged.
- [ ] `docs/07 §8` describes the implemented observability (not a TODO); roadmap changelog updated.
- [ ] Whole-plan consistency sweep: zero unresolved contradictions.

**Risks:** testing real exporter path — deliberately skipped (needs creds); manual verification deferred to Phase 9 deploy. No-op path is what keeps CI green. Doc drift — the step-6 consistency sweep catches stale references.

## Success criteria (foundation-level)

- Every request produces a structured JSON log line in stdout (→ Cloud Logging in prod) carrying `request_id`.
- Every request produces a Cloud Trace root span (otelgin); the Gemini call nests as a child span (otelhttp on the genai HTTP client).
- `go test ./...` green; **no-op fallback verified**: app boots + serves `/healthz` with zero GCP creds, no exporter errors.
- No PII in any log line (verified by test).
- `docs/07-deployment-guide.md §8` updated to reflect concrete observability; roadmap changelog entry added.

## Instrumentation map

What each feature phase adds **on top of** this foundation (inline, during build — not part of Phase 0):

| Feature phase | Instrumentation added inline |
|---|---|
| [Phase 2 — Image Verdict](./phase-02-image-verdict-gemini-vision.md) | Gemini round-trip already a child span (otelhttp, wired here in step 3); add `verdict_level` + `gemini_model` log attrs on the analyze line. |
| [Phase 3 — Tier-1 Lookup](./phase-03-tier-1-blacklist-lookup.md) | `tracer.Start("firestore.lookup")` child span + `latency_ms` log; `tier` attr. |
| [Phase 4 — Text PII Redaction](./phase-04-text-pii-redaction.md) | Redaction is privacy-sensitive — no PII logged; optional `status=redacted` counter. |
| [Phase 5 — 4-Tier & Reporting](./phase-05-4-tier-model-reporting.md) | Custom metrics land here: `shieldvn.verdicts` (by level/tier), report-created counter. |
| [Phase 6 — Moderation & Appeals](./phase-06-moderation-appeals.md) | Moderation action spans; resolve-latency + appeal-created metrics. |
| [Phase 7 — Image Masking](./phase-07-image-pii-masking.md) | `tracer.Start("image.mask")` span + OCR `latency_ms` (the <3s budget is observable). |
| [Phase 8 — Viral Card](./phase-08-viral-warning-card.md) | Frontend-only (client-side Canvas); no backend instrumentation. FE error tracking deferred (YAGNI). |
| [Phase 9 — Deploy & Demo](./phase-09-deploy-demo.md) | **Ops capstone:** dashboards, alert policies, SLOs; prod verification of log↔trace correlation; real-exporter smoke. |

## Open questions

1. Exact OTel + GCP exporter module versions → pin at step 3 via `go get` (latest stable).
2. Custom metric names (`shieldvn.verdicts`, `shieldvn.gemini.latency_ms`) finalized when instrumentation lands in feature phases 2–8 — not blocking this foundation.
