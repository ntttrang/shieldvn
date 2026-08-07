# ShieldVN — Deployment Guide

> Target: **Google Cloud Run** (scale-to-zero, +deployment bonus). 2 service: backend (Go) + frontend (Next.js).

## 1. Prerequisites

- GCP project, billing on.
- `gcloud` CLI đã auth (`gcloud auth login`).
- APIs enable: Cloud Run, Firestore, Sheets, Secret Manager (nếu dùng).
- Gemini API key (Google AI Studio).
- Firestore ở **Native mode**. Sheet Tier-1 tạo sẵn + share read.

## 2. Secrets (KHÔNG commit git)

```
GEMINI_API_KEY=...
GOOGLE_CLOUD_PROJECT=shieldvn-xxxx
FIRESTORE_COLLECTION_REPORTS=reports
SHEET_ID=...                   # Tier-1 seed, read-only
ALLOWED_ORIGIN=https://shieldvn-web-xxx.run.app
```

Backend dùng service account (Firestore) — download JSON, inject qua Cloud Run secret (không lưu file trong image).

## 3. Backend Dockerfile (multi-stage, Alpine + Tesseract + vie)

```dockerfile
# Build
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/api

# Runtime (Tesseract + vie traineddata)
FROM alpine:latest
RUN apk --no-cache add ca-certificates tesseract-ocr tesseract-ocr-data-vie
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

> Tesseract chạy CLI (`exec.Command("tesseract", img, "stdout", "tsv")`) — không cần cgo, chỉ cần binary + `vie.traineddata` (đã có qua `tesseract-ocr-data-vie`).

## 4. Deploy Backend

```bash
gcloud run deploy shieldvn-api \
  --source ./backend \
  --region asia-southeast1 \
  --port 8080 \
  --set-env-vars "ALLOWED_ORIGIN=https://shieldvn-web-xxx.run.app" \
  --set-secrets "GEMINI_API_KEY=gemini-key:latest" \
  --allow-unauthenticated \
  --memory 1Gi --cpu 1 \
  --min-instances 0 --max-instances 10
```

Lấy URL: `https://shieldvn-api-xxx.run.app/api/v1/analyze`.

## 5. Deploy Frontend (Next.js)

```bash
# Within ./frontend
gcloud run deploy shieldvn-web \
  --source . \
  --region asia-southeast1 \
  --set-env-vars "NEXT_PUBLIC_API_URL=https://shieldvn-api-xxx.run.app" \
  --allow-unauthenticated
```

(Hoặc Vercel cho FE nếu muốn simplest — nhưng ở Cloud Run thì gom điểm Google.)

## 6. Firestore Schema (collections)

```
reports/{id}:  { entity_type, entity_value, tier, reporter_id?, status, evidence_url?, created_at }
moderation/{id}: { report_id, status: pending|approved|rejected, reviewer, resolved_at }
appeals/{id}:  { report_id, claimant_proof, status, created_at }
```

## 7. Cost & Scale

- Scale-to-zero → ~0đ khi không traffic (hackathon). Tesseract cần RAM, set 1Gi.
- Free tier Gemini đủ demo. Firestore/Sheets free tier dư.
- Mở CORS đúng `ALLOWED_ORIGIN` = URL frontend.

## 8. Observability (Implemented — Phase 0)

### Structured Logging (slog → Cloud Logging)

- All logs are **structured JSON** on stdout via Go stdlib `slog` (zero dependency).
- Cloud Run auto-ingests stdout → **Cloud Logging**. Each log line includes `time`, `level`, `msg`, and structured metadata.
- Every request's logs carry `request_id` (from `X-Request-ID` header or auto-generated) for end-to-end tracing in Cloud Logging.
- Log level controlled by env var `LOG_LEVEL` (values: `debug`, `info`, `warn`, `error`; default: `info`).
- **PII is never logged.** The `LogMeta` helper only accepts a fixed set of metadata keys (`request_id`, `latency_ms`, `verdict_level`, `error_type`, `tier`, `gemini_model`, `attempt`, `status`). Unknown keys are dropped with a warning. Phone numbers, bank accounts, national IDs, prompts, and image bytes cannot reach logs by construction.

### Distributed Tracing (OpenTelemetry → Cloud Trace)

- `otelgin` middleware creates a **root span** for every HTTP request → auto-exported to **Cloud Trace** on Cloud Run.
- Outbound Gemini SDK calls use an `otelhttp`-instrumented HTTP client → each API call appears as a **child span**.
- Cloud Run auto-correlates log lines ↔ trace via `logging.googleapis.com/trace` (when trace ID is emitted).
- Custom business spans (Firestore lookup, image masking) are added inline during feature phases 2–8.

### Environment Variables

| Variable | Required | Default | Purpose |
|---|---|---|---|
| `LOG_LEVEL` | No | `info` | slog level (`debug`/`info`/`warn`/`error`) |
| `GOOGLE_CLOUD_PROJECT` | On Cloud Run | — | GCP project ID for OTel exporters |
| `K_REVISION` | Auto-set | — | Cloud Run revision (triggers real exporters) |

### No-Op Fallback (Dev / CI)

When `K_REVISION` is not set (i.e., not on Cloud Run), all OTel providers initialize as **no-op** — zero network calls, zero errors, no GCP credentials needed. The app boots and serves `/healthz` 200 with full logging but no trace export. This keeps `go test` and local dev frictionless.

### Ops Basics

- Cloud Run revision history — rollback via single `gcloud` command.
- Healthcheck: `GET /healthz` returns `200 OK`.
- Graceful shutdown: `SIGTERM` → HTTP server drains → OTel providers flush → clean exit.

## 9. Pre-Submit Checklist

- [ ] `ALLOWED_ORIGIN` đúng URL FE thật.
- [ ] Secrets không trong git/commit/image.
- [ ] `/healthz` 200.
- [ ] Tier-1 Sheet public read (judges mở được).
- [ ] PWA installable trên mobile.
- [ ] Demo video quay trên URL production.
