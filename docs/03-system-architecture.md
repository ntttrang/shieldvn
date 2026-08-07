# ShieldVN — System Architecture

## 1. Overview

```
PWA (Next.js)  — 2 screens: Upload / Result (+ Warning-card share)
        │ multipart/form-data (image) + JSON (text)
        ▼
GOLANG / Gin API   (Google Cloud Run · RAM-only · zero-disk)
  1. Text PII redact  (regex)
  2. Image PII mask   (Tesseract TSV → regex phone/STK/CCCD → bbox mask → clean JPEG)
  3. Parallel goroutines:
        (a) Gemini 2.5 Flash — vision + responseSchema
        (b) Firestore tier-lookup on extracted STK/SĐT
  4. Merge → GREEN/YELLOW/RED + evidence. Persist nothing raw.
        ├─► Gemini API          (clean payload only)
        └─► Firestore           (reports / 4-tier logic / moderation / appeals)
              + Google Sheet    (Tier-1 seed, read-only, judges-visible)
Stretch: Telegram bot entrypoint (forward msg → verdict)
```

## 2. Components

| Layer | Tech | Role |
|---|---|---|
| Frontend | Next.js (React) + Tailwind, PWA | 2 màn hình, client-side image compress trước upload, render Risk Card + Warning card (Canvas) |
| API Gateway | Go 1.22 + Gin | Routing, multipart parse, orchestration |
| PII Engine | Go `regexp` + `image/draw` + **Tesseract CLI** (`vie.traineddata`, TSV mode) | Redact text + mask image regions trên RAM |
| AI Core | Google Gen AI Go SDK (`google.golang.org/genai`) | `gemini-2.5-flash`, `ResponseMIMEType: application/json` + `responseSchema` |
| Dynamic DB | **Firestore** | reports, 4-tier scoring, moderation queue, appeals |
| Visible DB | **Google Sheets API** (read-only) | Tier-1 seed blacklist — judges mở thấy live |
| Deploy | Cloud Run (scale-to-zero) | Docker multi-stage (Alpine + tesseract + vie) |

## 3. E2E Data Flow

```
UI ──upload──► Gin ──text redact──► PII Engine ──clean──► parallel ──► {Gemini JSON, Firestore match}
                                                              │
UI ◄──merge JSON── Gin ◄─────────────────────────────────────┘
```

1. UI gửi image (đã nén client) + text/URL.
2. Gin: redact text (regex) + mask image regions (Tesseract bbox).
3. Goroutines song song: (a) Gemini vision trên clean payload; (b) Firestore lookup STK/SĐT.
4. Gin merge → JSON verdict. Không ghi đĩa; RAM giải phóng sau request.

## 4. Gemini Contract (Structured Output)

```jsonc
responseSchema:
{
  "risk_score": "GREEN|YELLOW|RED",
  "confidence_score": "number 0-1",
  "detected_patterns": ["string"],
  "evidence": ["string (Vietnamese, plain language)"],
  "recommendations": ["string"],
  "extracted_entities": { "bank_account"?: "string", "phone_number"?: "string", "url"?: "string" }
}
```

`extracted_entities` dùng cho Firestore tier-lookup (KHÔNG dùng raw để tra cứu — dùng entity đã extract, vốn đã sạch).

## 5. 4-Layer Blacklist Model (Firestore)

| Tier | Source | Rule |
|---|---|---|
| 1 — Chính thống | Bộ TTTT, Bộ CA, NH, ChongLuaDao (seed Sheet) | Auto blacklist |
| 2 — Cộng đồng uy tín | Reporter có điểm uy tín cao | 2–3 báo cáo trùng → pending |
| 3 — Vô danh | Anonymous report | Pending, KHÔNG gắn "lừa đảo 100%" |
| 4 — Kháng nghị | Chủ tài khoản bị báo nhầm | Upload proof + Gemini OCR check validity → moderator resolve ≤24h |

Implement: tier-scoring logic trong Go + moderator console tối giản; **seed Tiers 2–4 với sample data** để demo.

## 6. Privacy & Security

- **Free-tier Gemini có thể dùng prompt để train** → masking PII trước send là **bắt buộc**, không phải gimmick. Pitch: *"Số tài khoản bị strip trước khi AI thấy — Google không bao giờ nhận được."*
- Backend RAM-only, zero-disk. Image không bao giờ lưu.
- Secrets (Gemini key, Firestore svc account) qua Cloud Run env/secrets — không commit git.
- Rate-limit cơ bản + size guard (image < 5MB) tại Gin.
- Optional: paid Flash tier cho commitment no-training (slide "real product").

## 7. Trade-offs Accepted

- Sheets chỉ read-only (Tier-1 view) → tránh write race; moderation sống trong Firestore.
- Masking server-side (Tesseract) → consistent fast, khớp skill Go + doc; thay vì client-side Tesseract.js (stronger story nhưng chậm trên thiết bị yếu).
- Next.js tách riêng → richer UI; rủi ro thời gian giảm bằng hard-cap 2 màn hình.

## 8. Risks → see `06-project-roadmap.md` §Risks
