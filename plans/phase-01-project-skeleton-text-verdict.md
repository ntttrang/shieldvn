---
phase: 1
title: "Project Skeleton & Text Verdict"
status: pending
effort: "1d"
priority: P1
dependencies: []
---

# Phase 1: Project Skeleton & Text Verdict

## Overview
Stand up Go/Gin backend + Next.js frontend stub, wire **Gemini text analysis** with structured JSON, render a minimal GREEN/YELLOW/RED verdict. End of phase: **paste text → verdict** running E2E.

## Requirements
- Functional: `POST /api/v1/analyze` accepts `text_prompt` → calls Gemini 2.5 Flash → returns `ScamAnalysisResponse` JSON.
- Non-functional: cold start < 2.5s, warm < 1.5s; FE displays color badge + evidence list.
- **[Validation S1]** Gemini JSON: use `responseSchema` + Go struct unmarshal + **retry once** on malformed (no crash). Confirm `ResponseSchema` API + SDK version via `docs-seeker`.

## Architecture
FE (Next.js) ──JSON──► Gin handler ──► GeminiService (genai SDK, responseSchema) ──► JSON ──► FE render. No PII/DB in this phase.

## Related Code Files
- Create: `backend/go.mod`, `backend/cmd/api/main.go`, `backend/internal/config/config.go`, `backend/internal/handler/analyze_handler.go`, `backend/internal/service/gemini_service.go`, `backend/internal/model/analysis.go`
- Create: `frontend/` (Next.js init), `frontend/app/page.tsx`, `frontend/components/risk-card.tsx`, `frontend/components/evidence-list.tsx`, `frontend/lib/api.ts`
- Create: `backend/.env.example` (GEMINI_API_KEY, ALLOWED_ORIGIN), `backend/Makefile`/run script

## Implementation Steps
1. `go mod init shieldvn-backend`; install `gin`, `google.golang.org/genai`.
2. `config.go` load env (godotenv or os.Getenv).
3. `model/analysis.go` define `AnalysisInput`, `AnalysisResult` + JSON tags matching `ScamAnalysisResponse`.
4. `gemini_service.go`: `Analyze(ctx, prompt)` → `GenerateContent` with `ResponseMIMEType: application/json` + `responseSchema` (risk_score/confidence/detected_patterns/evidence/recommendations/extracted_entities). System instruction = VN scam scenario checklist (see `docs/03-system-architecture.md` §4).
5. `analyze_handler.go`: parse JSON → call service → return JSON; `GET /healthz`.
6. `main.go`: wire Gin router, CORS (ALLOWED_ORIGIN), port 8080.
7. Next.js init (`npx create-next-app`, Tailwind). `lib/api.ts` calls `/api/v1/analyze`. `page.tsx`: textarea + "Check" button → `risk-card` + `evidence-list`.
8. Local smoke: paste a VN job-scam sample ("Recruit CTV, deposit money, close orders") → expect YELLOW/RED.

## Frontend
- `create-next-app` + Tailwind; `app/page.tsx` (textarea + "Check" button); `components/risk-card.tsx`; `components/evidence-list.tsx`; `lib/api.ts` (`POST /api/v1/analyze`).
- Privacy banner placeholder (finalize text in Phase 4).

## Tests
- Build gates: `go build ./...` + `npm run build` pass.
- Manual smoke: paste a VN job-scam sample ("Recruit CTV, deposit money, close orders") → YELLOW/RED.

## Success Criteria
- [ ] `POST /api/v1/analyze` returns JSON matching schema for text input
- [ ] FE displays GREEN/YELLOW/RED badge + evidence from text
- [ ] `/healthz` 200; CORS correct for localhost:3000
- [ ] `go build ./...` + `npm run build` pass

## Risk Assessment
- **Gemini JSON doesn't conform to schema** → use `responseSchema` (structured output) not hand-parse; fallback retry once. Mitigate: test fixed prompt first.
- **Weak system instruction** → refine VN scam checklist with 3–5 samples (fake bill, part-time/collaborator recruiter, VNeID impersonation).
