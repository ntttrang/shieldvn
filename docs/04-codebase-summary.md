# ShieldVN — Codebase Summary

> Greenfield (từ 2026-07-31). Tài liệu mô tả cấu trúc **dự kiến** theo `03-system-architecture.md` + `05-code-standards.md`.

## 1. High-Level

2 thành phần deploy độc lập:
- **`backend/`** — Go + Gin API trên Cloud Run. Owner của: PII engine, Gemini call, Firestore/Sheet, 4-tier logic.
- **`frontend/`** — Next.js PWA. Owner của: 2 màn hình, image compress, Risk Card, Warning card.

## 2. Backend Module Map

| Package | Trách nhiệm | File chính (dự kiến) |
|---|---|---|
| `cmd/api` | bootstrap, wiring, Gin router | `main.go` |
| `internal/handler` | HTTP handlers (analyze, report, moderate, appeal) | `analyze_handler.go`, `report_handler.go` |
| `internal/service` | business logic | `gemini_service.go`, `pii_text_sanitizer.go`, `pii_image_masker.go`, `blacklist_service.go`, `moderation_service.go`, `analysis_service.go` (orchestrator) |
| `internal/store` | Firestore + Sheets clients | `firestore_store.go`, `sheets_store.go` |
| `internal/model` | domain + DTO | `analysis.go`, `report.go`, `verdict.go` |
| `internal/config` | env/secrets | `config.go` |

**Key flows:**
- `analysis_service.go` — orchestrator: text redact → image mask → parallel (Gemini ‖ Firestore lookup) → merge verdict.
- `pii_image_masker.go` — gọi `tesseract <img> stdout tsv`, parse TSV (word + bbox), regex match STK/phone/CCCD, `image/draw` mask.
- `blacklist_service.go` — tier-scoring (Tier 1 auto, Tier 2–3 pending, Tier 4 appeal).

## 3. Frontend Module Map

| Vị trí | Trách nhiệm |
|---|---|
| `app/page.tsx` | màn Upload (nút tải ảnh, khung dán text/URL, privacy banner) |
| `app/result/` | màn Kết quả (Risk Card + Evidence + actions) |
| `components/uploader.tsx` | file picker + client compress |
| `components/risk-card.tsx` | badge GREEN/YELLOW/RED + confidence |
| `components/evidence-list.tsx` | danh sách bằng chứng tiếng Việt |
| `components/warning-card.tsx` | Canvas render PNG chia sẻ |
| `lib/api.ts` | API client (single source) |
| `lib/image-compress.ts` | nén ảnh < 1.5MB trước upload |
| `public/manifest.json` | PWA |

## 4. How to Run (local)

```bash
# Backend
cd backend
cp .env.example .env          # điền GEMINI_API_KEY, FIRESTORE creds, SHEET_ID
go mod tidy
go run ./cmd/api              # :8080

# Frontend
cd frontend
npm install
echo "NEXT_PUBLIC_API_URL=http://localhost:8080" > .env.local
npm run dev                   # :3000
```

## 5. Dependencies (key)

**Go:** `github.com/gin-gonic/gin`, `google.golang.org/genai`, `cloud.google.com/go/firestore`, `google.golang.org/api/sheets/v4`, `github.com/stretchr/testify`. System: `tesseract-ocr` + `tesseract-ocr-data-vie`.

**FE:** `next`, `react`, `tailwindcss`. (PWA qua `next-pwa` hoặc manifest thủ công.)

## 6. State

Chưa có code (chỉ `docs/` + `plans/`). Bắt đầu từ Roadmap Phase 0.
