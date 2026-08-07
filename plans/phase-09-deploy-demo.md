---
phase: 9
title: "Deploy & Demo"
status: pending
effort: "3d"
priority: P1
dependencies: [7, 8]
---

# Phase 9: Deploy & Demo

## Overview
Deploy backend + frontend to **Cloud Run** (+deployment bonus), polish PWA, run manual test matrix, record demo video <2 minutes + pitch deck. Submit entry.

## Requirements
- Functional: 2 Cloud Run services (api + web) public; `/healthz`; CORS correct; PWA installable.
- Non-functional: scale-to-zero; secrets via Cloud Run env; Tier-1 Sheet public.

## Architecture
Backend image: multi-stage Alpine + tesseract + vie (see `docs/07-deployment-guide.md`). Frontend: Next.js standalone build → Cloud Run (or Vercel). Env: `GEMINI_API_KEY`, Firestore svc account, `SHEET_ID`, `ALLOWED_ORIGIN`.

## Related Code Files
- Create/finalize: `backend/Dockerfile`, `frontend/Dockerfile` (or Vercel config), `frontend/public/manifest.json` + icons
- Modify: env configs, CORS allow production URL
- Reference: `docs/07-deployment-guide.md` (gcloud commands, checklist), `docs/08-test-plan.md` (manual matrix)

## Implementation Steps
1. Backend Dockerfile multi-stage (builder + Alpine runtime + tesseract+vie). Test `docker build` + run local.
2. `gcloud run deploy shieldvn-api --source ./backend ...` (region asia-southeast1, 1Gi mem, scale 0–10, secrets).
3. Frontend build standalone; deploy Cloud Run (or Vercel); set `NEXT_PUBLIC_API_URL`.
4. Set `ALLOWED_ORIGIN` = real FE URL (CORS).
5. PWA: `manifest.json` + icons, standalone display; test install mobile.
6. Manual test matrix (`docs/08-test-plan.md` §3): AN-01..04, SEC-01..03, TIER-01..04, UI-01..03, PERF-01..02.
7. PII leak audit: 50 sample grep → 0 match.
8. Record demo video <2 minutes (flow: upload fake bill → RED → evidence → warning card). Pitch deck (problem, demo, Gemini usage, privacy, 4-tier, deploy).

## Frontend
- PWA `manifest.json` + icons, standalone display; production build; CORS allow prod URL.

## Tests
- Full manual matrix (AN/SEC/TIER/UI/PERF — `docs/08-test-plan.md`).
- PII-leak audit (50 sample grep → 0); warm/cold latency; cross-browser Chrome/Safari/Zalo.

## Success Criteria
- [ ] Backend + FE public on Cloud Run, `/healthz` 200
- [ ] CORS correct, secrets not in git
- [ ] PWA installable mobile
- [ ] Manual matrix 100% pass (AN + SEC)
- [ ] PII audit 0% leak
- [ ] Demo video <2 minutes + pitch deck ready

## Observability (ops capstone)

This phase owns the prod ops surface on top of the [Phase 0 foundation](./phase-00-observability-foundation.md): build Cloud Logging/Trace/Monitoring dashboards, alert policies, SLOs; verify log↔trace correlation end-to-end on Cloud Run (real exporters); smoke the real-exporter path that CI skips (no-op only). Verify no PII in any prod log line via the 50-sample audit.

## Risk Assessment
- **Tesseract in Cloud Run image** → test build carefully; `tesseract-ocr-data-vie` must be present. Cold start +few seconds (acceptable, <2.5s goal may relax for demo).
- **CORS/URL wrong** → checklist `docs/07` §9; verify before recording video.
- **Cold start latency** → set min-instances=0 keep cost 0; if demo needs snappy, warm beforehand.
- **[Validation S1] Rate limit demo day** → free-tier RPM/RPD may cap when many judges test; prepare **cached-demo verdict** (sample response) as fallback if capped.
- **Deadline slip** → if Phase 6/8 not done, deploy core (1–4+7) still submittable; Phase 5/6/8 bonus.
