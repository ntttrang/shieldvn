---
phase: 7
title: "Image PII Masking"
status: pending
effort: "1d"
priority: P2
dependencies: [4]
---

# Phase 7: Image PII Masking (Tesseract)

## Overview
Mask PII regions (phone number/bank account number/national ID) on image **before** sending to Gemini, using **Tesseract OCR** free (`vie.traineddata`) → TSV bbox → regex match → mask. Completes the privacy claim for images. Fills the `MaskImageRegions` gap (old doc didn't say where bbox comes from).

## Requirements
- Functional: upload image → OCR TSV → find word matching phone/bank account number/national ID → draw black box (image/draw) → send clean JPEG to Gemini.
- Non-functional: Tesseract in Docker image (`apk add tesseract-ocr tesseract-ocr-data-vie`); sufficient accuracy on digital text images; <3s added to pipeline.

## Architecture
Handler: image bytes → `pii_image_masker.Mask(bytes)` → clean bytes → Gemini. `Mask`: `exec.Command("tesseract", imgPath, "stdout", "-l", "vie", "tsv")` → parse TSV (cols: text, conf, left, top, width, height) → regex match → `image/draw` box → encode JPEG.

## Related Code Files
- Create: `backend/internal/service/pii_image_masker.go`, `backend/internal/service/pii_image_masker_test.go`, `backend/internal/service/tsv_parser.go`
- Modify: `backend/internal/handler/analyze_handler.go` (call mask before Gemini), `backend/Dockerfile` (add tesseract + vie)
- Reference: `docs/first_idea.md` `MaskImageRegions` (fix: bbox from Tesseract, not a parameter)

## Implementation Steps
1. `tsv_parser.go`: parse stdout TSV → `[]Word{Text, Rect}`. Unit test with TSV sample (U-MASK-01).
2. `pii_image_masker.go`: decode image → run tesseract (write image to tmp or stdin pipe) → parse → regex match (phone/bank account number/national ID, reuse Phase 4 regex) → `draw.Draw` black box → encode JPEG. No permanent disk write (tmp cleaned after).
3. Unit test: input image+rects → box pixels black (U-MASK-02).
4. Wire handler: mask before `gemini_service.Analyze`. Remove Phase 2 TODO.
5. Dockerfile: `RUN apk --no-cache add tesseract-ocr tesseract-ocr-data-vie` (see `docs/07-deployment-guide.md`).
6. Audit: image with bank account number/national ID → image sent to Gemini has black box over that region (SEC-02 manual).

## Frontend
- None new. Ensure raw image NOT displayed; UI only shows result.

## Tests
- `tsv_parser_test.go` (U-MASK-01); `pii_image_masker_test.go` (U-MASK-02 black box).
- **10-sample bbox-accuracy check** (bar <80% → PaddleOCR); visual spot-check 5 images box correct PII region.

## Success Criteria
- [ ] Image containing bank account number/national ID → that region boxed black before reaching Gemini (SEC-02)
- [ ] Tesseract runs in Docker image local + Cloud Run
- [ ] Verdict quality preserved (enough context outside masked region)
- [ ] TSV parse + mask unit test pass

## Observability

Wraps the OCR+mask pipeline in a `tracer.Start("image.mask")` span and logs `latency_ms` (Phase 0 providers) — makes the <3s added-latency budget observable. Masked-vs-unmasked image bytes never logged.

## Risk Assessment
- **Poor VN accuracy** → **bar [Validation S1]: if <80% bbox-correct on 10 VN screenshot samples → add PaddleOCR sidecar**. Digital text images (banking/Zalo) usually high accuracy.
- **Tesseract cgo/binary** → use CLI `exec.Command` (no cgo), just need binary in image.
- **Box misaligned** → use exact bbox from TSV; visual test 5 images.
