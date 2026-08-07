---
phase: 3
title: "Tier-1 Blacklist Lookup"
status: pending
effort: "2d"
priority: P1
dependencies: [2]
---

# Phase 3: Tier-1 Blacklist Lookup

## Overview
Look up **Tier-1 seed** (Google Sheet read-only) in parallel with Gemini, using entities (bank account number/phone number) Gemini extracts, merge results into evidence. **End = full core demo (safe floor reached).**

## Requirements
- Functional: after Gemini returns `extracted_entities`, query Firestore/Sheet to check whether bank account number/phone number is in Tier-1; if yes → escalate verdict + add evidence "bank account number in Tier 1 blacklist".
- Non-functional: lookup runs // with Gemini (goroutine) to avoid increasing latency; Sheet public read.

## Architecture
Gin orchestrator: goroutine A = Gemini vision; goroutine B = wait for `extracted_entities` then query blacklist. `errgroup`/`sync.WaitGroup` merge → merge. (Note: B depends on A's output, so 2 stages: Gemini first, then lookup — or launch lookup after entities available. Keep simple: run entities→lookup sequentially, parallel only if computing other metadata.)

**Tier-1 storage:** import seed from Sheet into Firestore collection `blacklist_tier1` (entity_type, entity_value) at boot or via one-time script. Sheet stays public for judges to view.

## Related Code Files
- Create: `backend/internal/store/firestore_store.go`, `backend/internal/store/sheets_store.go`, `backend/internal/service/blacklist_service.go`
- Modify: `backend/internal/handler/analyze_handler.go` (orchestration + merge), `backend/internal/model/analysis.go` (add blacklist hit field)
- Create: `scripts/seed_tier1.go` (import Sheet → Firestore), sample Tier-1 Sheet (public scam bank account number/phone number/URL)

## Implementation Steps
1. Create Google Sheet Tier-1 (columns: entity_type, entity_value, source, note); share read-only.
2. `sheets_store.go`: read range → map entity.
3. `firestore_store.go`: client init from svc account; collection `blacklist_tier1`; `IsListed(entityType, value) (bool, source)`.
4. `seed_tier1.go`: read Sheet → upsert Firestore (run once).
5. `blacklist_service.go`: accept `extracted_entities` → lookup each entity → return hits.
6. Handler merge: if Tier-1 hit → verdict at least YELLOW, evidence "bank account number … in Tier 1 blacklist (source: …)".
7. Test: seeded bank account number → escalate; clean bank account number → no change.

## Frontend
- `risk-card` / `evidence-list` render blacklist-hit evidence ("bank account number … in Tier 1 blacklist — source …"). Minimal UI change.

## Tests
- `blacklist_service_test.go` (mocked store): seeded bank account number → escalate; clean bank account number → no change.
- Seed script idempotent (re-run no duplicate).

## Success Criteria
- [ ] Bank account number in Tier-1 seed → verdict escalate + evidence with clear source
- [ ] Lookup doesn't increase perceived latency (>200ms) or runs in background
- [ ] Sheet public accessible (judges-visible)
- [ ] Seed script idempotent

## Observability

Wraps the Firestore/blacklist lookup in a `tracer.Start("firestore.lookup")` child span (Phase 0 providers) and logs `latency_ms` + `tier` attr via `observability.LogMeta` — makes the >200ms latency budget observable. Lookup values (bank account number/phone number) are PII: never logged (Phase 0 privacy gate).

## Risk Assessment
- **Race entities→lookup** → run lookup AFTER Gemini returns entities (simple, correct). No need for // if not significant.
- **Firestore creds local** → use svc account JSON in `.env` (DO NOT commit); guide in `docs/07-deployment-guide.md`.
- **Safe floor reached here** — if time runs out, this is still a winning demo.
