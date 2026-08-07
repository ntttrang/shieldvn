---
phase: 2
title: "Image Verdict (Gemini Vision)"
status: review
effort: "2d"
priority: P1
dependencies: [1]
branch: feat/phase-02-image-verdict-gemini-vision
---

# Phase 2: Image Verdict (Gemini Vision)

## Overview
Add image upload (multipart) → send image to Gemini vision → verdict. This is the **"wow" moment** of the demo. PII not masked yet (Phase 7) — note temporarily in UI.

## Requirements
- Functional: `POST /api/v1/analyze` accepts additional `image` (multipart, <5MB) → Gemini multi-modal → verdict + evidence (e.g. "mismatched font", "missing FT code").
- Non-functional: client compresses image <1.5MB before upload; payload limit 5MB.

## Architecture
FE compress (Canvas) ──multipart──► Gin (multipart parse) ──► `gemini_service.Analyze(ctx, prompt, imageBytes)` uses `genai.NewPartFromBytes(bytes, "image/jpeg")` + text part ──► JSON ──► FE.

## Related Code Files
- Modify: `backend/internal/handler/analyze_handler.go` (parse multipart), `backend/internal/service/gemini_service.go` (add image part), `backend/internal/model/analysis.go`
- Create: `frontend/components/uploader.tsx`, `frontend/lib/image-compress.ts`
- Modify: `frontend/app/page.tsx` (upload or paste mode)

## Implementation Steps
1. Handler: `c.Request.FormFile("image")` + `c.PostForm("text_prompt")`; validate type (jpg/png) + size.
2. `gemini_service`: build `[]*genai.Part` — image bytes + prompt text; keep structured output config.
3. FE `image-compress.ts`: downscale + to JPEG quality ~0.7, max ~1400px, target <1.5MB.
4. `uploader.tsx`: file picker + preview + compress + send multipart.
5. Sample test: fake VietQR bill → RED + font evidence; real bill → GREEN.
6. Refine system instruction for vision (require specifying signs on image in Vietnamese).

## Frontend
- `components/uploader.tsx` (file picker + preview); `lib/image-compress.ts` (Canvas, target <1.5MB); `app/page.tsx` toggle upload/paste.

## Tests
- Manual: fake bill → RED + font evidence; real bill → GREEN; image >5MB → compressed, no 413.
- Build pass; warm latency < 2.5s with image.

## Success Criteria
- [ ] Upload fake bill image → RED with specific evidence (font/layout)
- [ ] Upload real bill → GREEN (no false alarm)
- [ ] Image >5MB compressed by FE; no 413 error
- [ ] Warm latency < 2.5s with image

## Observability

First consumer of the [Phase 0 foundation](./phase-00-observability-foundation.md). The Gemini round-trip is already a child span via the instrumented HTTP client (otelhttp, wired in Phase 0 step 3). Add `verdict_level` + `gemini_model` attrs to the analyze log line via `observability.LogMeta`. Image bytes must never reach logs (privacy guardrail in Phase 0).

## Risk Assessment
- **Generic evidence** → prompt forces listing specific signs + location; test 5 bill samples.
- **Heavy image slow** → client compress mandatory; reject >5MB.
- **Note:** this phase SENDS RAW IMAGE to Gemini (not yet masked). Mark TODO Phase 7. Do not use images containing real PII when testing.
