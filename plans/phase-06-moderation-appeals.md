---
phase: 6
title: "Moderation & Appeals"
status: pending
effort: "3d"
priority: P3
dependencies: [5]
---

# Phase 6: Moderation & Appeals

## Overview
Minimal moderator console (web) to approve/reject reports + appeal flow (upload proof → Gemini OCR validity check → moderator resolve). Completes the "4-tier with moderation" story.

## Requirements
- Functional: moderator views pending list → approve/reject → update Firestore → reflect in lookup. Appeal: claimant submits **text claim** (report_id + note) → create appeal → moderator resolve. **[Validation S1]** NO image proof accepted (avoid storing claimant PII).
- Non-functional: separate console (route `/moderate`, basic token guard — no full auth needed for hackathon); resolve ≤24h is displayed SLA.

## Architecture
FE route `/moderate` (guarded by env token header). Endpoints: `GET /api/v1/moderation/pending`, `POST /api/v1/moderation/{id}/resolve`, `POST /api/v1/appeal` (JSON: report_id + note), `GET /api/v1/appeals`. Appeal = **text claim only** (no proof image) — **[Validation S1]** avoid storing claimant PII.

## Related Code Files
- Create: `backend/internal/handler/moderation_handler.go`, `backend/internal/handler/appeal_handler.go`
- Modify: `backend/internal/service/moderation_service.go` (resolve, promote to blacklist), `backend/internal/store/firestore_store.go`
- Create: `frontend/app/moderate/page.tsx`, `frontend/components/appeal-form.tsx`

## Implementation Steps
1. `moderation_handler`: list pending (from Phase 5), resolve {status: approved|rejected, reviewer}.
2. Resolve approved → promote entity into blacklist (Tier 2→blacklist) → lookup reflects.
3. `appeal_handler`: accept text claim (report_id + note) → create appeal record. **[Validation S1]** No image proof accepted.
4. FE `/moderate`: pending table + approve/reject buttons; appeals section.
5. Basic guard: compare header `X-Mod-Token` vs env (hackathon-grade; clearly note NOT production auth).
6. Seed: 1 demo appeal → moderator resolve → entity un-blacklist.

## Frontend
- `app/moderate/page.tsx` (pending table + approve/reject + appeals section); guard header `X-Mod-Token`.
- `appeal-form.tsx` — **text claim only**, NO image upload.

## Tests
- Manual TIER-03 (approve → reflect blacklist; reject → unlist); TIER-04 (text appeal → create → resolve).
- Verify no image proof upload path exists.

## Success Criteria
- [ ] Moderator approve → entity reflects blacklist; reject → remove (TIER-03 manual matrix)
- [ ] Appeal submit (text claim, no image) → create appeal; resolve updates status (TIER-04)
- [ ] Console guard token works
- [ ] No claimant PII leaks to logs (proof processed RAM-only)

## Observability

Moderation actions get `tracer.Start("moderation.resolve")` spans + resolve-latency metric; appeal submissions get an appeal-created counter (Phase 0 providers). Text claims are PII-adjacent — log only `status`/`report_id`, never claim text (Phase 0 privacy gate).

## Risk Assessment
- **Time short** → Phase 6 is P3; if slipped, keep Phase 5 (report+scoring) already enough to demo "4-tier". Cut console to 1 static list + resolve page.
- **Weak auth** → clearly this is hackathon guard; doc warns. Don't expose personal data.
- **Gemini OCR validity uncertain** → use as hint for moderator, no auto-reject.
