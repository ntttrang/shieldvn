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

## 8. Rollback / Ops

- Cloud Run revision history — rollback 1 lệnh.
- Logs: Cloud Logging. **Không log PII** (chỉ metadata: request id, latency, verdict level).
- Healthcheck: `GET /healthz` trả 200.

## 9. Pre-Submit Checklist

- [ ] `ALLOWED_ORIGIN` đúng URL FE thật.
- [ ] Secrets không trong git/commit/image.
- [ ] `/healthz` 200.
- [ ] Tier-1 Sheet public read (judges mở được).
- [ ] PWA installable trên mobile.
- [ ] Demo video quay trên URL production.
