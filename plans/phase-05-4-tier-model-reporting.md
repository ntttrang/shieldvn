---
phase: 5
title: "4-Tier Model & Reporting"
status: pending
effort: "3d"
priority: P2
dependencies: [4]
---

# Phase 5: 4-Tier Model & Reporting

## Overview
Firestore data model for **4-tier blacklist** + "Report" endpoint. Tier-scoring logic. Seed Tier 2–4 with sample data for demo. This is the **differentiator** of ShieldVN.

## Requirements
- Functional: user "Report additional info" from result screen → create report in Firestore. Tier-scoring: T1 auto, T2 needs 2–3 matches → pending, T3 anonymous pending (no "100%" label), T4 appeal.
- Non-functional: Firestore atomic writes (transaction for count); don't block analyze flow.
- **[Validation S1]** Pending T2/T3 must surface to user verdict: **YELLOW + note "community reports awaiting review"** (`blacklist_service` returns pending status → handler maps). No auto RED for unreviewed reports.

## Architecture
Firestore collections: `reports/{id}` (entity_type, entity_value, tier, reporter_id?, status, evidence_url?, created_at), `moderation/{id}`, `appeals/{id}`. `moderation_service.go` holds tier rules (see `docs/03-system-architecture.md` §5). Lookup in Phase 3 extended: check approved reports too.

## Related Code Files
- Create: `backend/internal/model/report.go`, `backend/internal/service/moderation_service.go`, `backend/internal/handler/report_handler.go`
- Modify: `backend/internal/store/firestore_store.go` (reports CRUD + transaction count), `backend/internal/service/blacklist_service.go` (merge approved Tier 2–4)
- Modify: `frontend/app/page.tsx` + `frontend/components/` ("Report additional info" button → entity/note/anonymous form)
- Create: `scripts/seed_tier234.go` (sample reports T2–T4)

## Implementation Steps
1. `model/report.go`: structs Report, Moderation, Appeal + status enum.
2. `firestore_store`: `CreateReport`, `CountReports(entity)` (transaction), `ListPending()`.
3. `moderation_service.go`: `Classify(report)` → tier rule; `ScoreEntity(entity)` → status (clean/pending/blacklisted) based on count + tier.
4. `report_handler.go`: `POST /api/v1/report` (guest-first, optional reporter handle).
5. FE: minimal report form (entity auto-fill from verdict, note, anonymous checkbox).
6. Seed T2 (2–3 reports on same bank account number), T3 (1 anon), T4 (1 appeal) to demo the flow.
7. Extend blacklist lookup: entity with enough T2 reports reaching threshold → escalate like T1.

## Frontend
- "Report additional info" form (entity auto-fill from verdict, note, anonymous checkbox).
- Result displays YELLOW + note "community reports awaiting review" when pending T2/T3 exists.

## Tests
- `moderation_service_test.go`: U-TIER-02 (2–3 matches → pending), U-TIER-03 (anon pending, no "100%" label).
- Firestore count transaction test (3 // reports → count = 3).

## Success Criteria
- [ ] T2: 2–3 matching reports → pending status (U-TIER-02 pass)
- [ ] T3 anon: pending, no "100%" label (U-TIER-03)
- [ ] Report created, count atomic
- [ ] Demo seed T2–T4 runs on UI/lookup

## Observability

Custom metrics land here against the Phase 0 meter provider: `shieldvn.verdicts` counter (attrs: `verdict_level`, `tier`) and a report-created counter. Pending→YELLOW+note transitions get a `status=pending` attr. No entity values in metrics — metadata only.

## Risk Assessment
- **Count race** → Firestore transaction. Mitigate: test 3 // reports.
- **Too complex** → keep MVP: tier rule hardcoded thresholds (2–3), no complex reputation engine (defer to Phase 6 if time remains).
- **False report on real shop** → T2/T3 only pending, no auto RED; needs moderator (Phase 6).
