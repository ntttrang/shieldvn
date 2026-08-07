---
title: "ShieldVN MVP — Scam Detection App"
description: "Implementation plan for ShieldVN — Privacy-First AI scam detector (Go/Gin + Next.js + Gemini + Tesseract + Firestore/Sheet). Phase 0 observability foundation + 9 vertical-slice feature phases; safe floor at Phase 3."
status: pending
priority: P1
branch: ""
tags: [hackathon, gemini, scam-detection, go, nextjs]
blockedBy: []
blocks: []
created: "2026-08-01T09:52:24.343Z"
createdBy: "ck:plan"
source: skill
---

# ShieldVN MVP — Scam Detection App

## Overview

ShieldVN — scam-warning AI assistant for Google AI Riser Vietnam 2026 (deadline 2026-08-30, solo ~4 weeks). Accepts screenshots / text / URL → verdict **GREEN/YELLOW/RED** + Vietnamese-language evidence, strips PII before sending to Gemini. Source of truth: `docs/01..08` + `plans/reports/from-brainstorm-to-planner-260731-1622-shieldvn-mvp-design-report.md`.

**Stack (locked):** Go 1.22 + Gin (backend) · Next.js + Tailwind PWA (frontend, 2 screens) · Gemini 2.5 Flash structured output · Tesseract OCR `vie` (image PII mask) · Firestore (4-tier/moderation/appeals) + Google Sheet (Tier-1 seed read-only) · Cloud Run.

## Golden Rule — Vertical Slice & Safe Floor

**Each phase ends = demoable. No later phase blocks an earlier one.** Stopping at Phase 3 (~day 5) still yields a winning core demo. Phases 4–9 are upside. If time is short, **Priority = Phases 1→3 mandatory, then 7 (masking) + 9 (deploy), then 5/6/8.** Phase 0 (observability foundation) is plumbing — built in hybrid/retrofit timing alongside the feature work; it never blocks the safe floor.

## Phases

| Phase | Name | Effort | Status |
|-------|------|--------|--------|
| 0 | [Observability Foundation (slog + request-id + OTel)](./phase-00-observability-foundation.md) | 10h | Pending |
| 1 | [Project Skeleton & Text Verdict](./phase-01-project-skeleton-text-verdict.md) | 1d | Pending |
| 2 | [Image Verdict (Gemini Vision)](./phase-02-image-verdict-gemini-vision.md) | 2d | Review (PR) |
| 3 | [Tier-1 Blacklist Lookup](./phase-03-tier-1-blacklist-lookup.md) | 2d | Pending |
| 4 | [Text PII Redaction](./phase-04-text-pii-redaction.md) | 1d | Pending |
| 5 | [4-Tier Model & Reporting](./phase-05-4-tier-model-reporting.md) | 3d | Pending |
| 6 | [Moderation & Appeals](./phase-06-moderation-appeals.md) | 3d | Pending |
| 7 | [Image PII Masking](./phase-07-image-pii-masking.md) | 1d | Pending |
| 8 | [Viral Warning Card](./phase-08-viral-warning-card.md) | 1d | Pending |
| 9 | [Deploy & Demo](./phase-09-deploy-demo.md) | 3d | Pending |

**Safe floor:** Phases 1–3. **Privacy-critical:** 4 & 7. **Differentiator:** 5–6 (4-tier). **Viral:** 8. **Ship:** 9. **Foundation (non-blocking):** Phase 0.

## Foundation & Instrumentation

- **[Phase 0 — Observability Foundation](./phase-00-observability-foundation.md)** (formerly the standalone `260801-1520-GH-02-observability-logging-tracing-foundation` plan) is the observability plumbing: `slog` structured logging + request-ID middleware + OTel providers/GCP exporters (no-op fallback off-Cloud-Run). **Soft reference, NOT a hard `blockedBy`:** it is built in hybrid/retrofit timing and never blocks the safe floor — feature phases may ship with plain logging first per the vertical-slice rule, then adopt the foundation incrementally.
- **Phases 2–8 instrument against Phase 0 inline** (Gemini/Firestore spans, verdict metrics). See the [instrumentation map](./phase-00-observability-foundation.md#instrumentation-map) for which phase adds what.
- **Phase 9 owns the ops capstone:** dashboards, alert policies, SLOs, and prod verification of log↔trace correlation.

## Key Decisions (from brainstorm — do not re-debate)

- Sheets = **read-only** Tier-1 seed (avoid write race); moderation lives in Firestore.
- Masking **server-side** Tesseract (consistent fast, matches Go skill). Fallback PaddleOCR if VN accuracy is poor.
- Free-tier Gemini can train → masking is **mandatory**, not a gimmick.
- Frontend Next.js (user choice) hard-cap 2 screens.
- Defer Playwright/k6 full suite → manual matrix + Go unit tests.

## Open Questions

1. Telegram bot entrypoint (forward message → verdict)? Stretch.
2. ~~Paid Gemini tier for the "no-training" slide?~~ → **Resolved (Validation S1):** free-tier + masking; cached-demo fallback. No paid tier.
3. Keep 1 Playwright smoke test or defer entirely?

## Validation Log

### Verification Results
- Tier: Standard (greenfield). Claims checked: ~20.
- Verified: `docs/01..08` + `first_idea.md` exist ✅; Tesseract `tsv` mode real; Firestore + `google.golang.org/genai` SDKs exist.
- Failed: 0. Unverified: exact genai Go SDK `ResponseSchema` field name + version → confirm in Phase 1 via `docs-seeker`.

### Session 1 — 2026-08-01
**Trigger:** Post-plan `/ck:plan validate` (default mode). **Questions asked:** 6.

#### Questions & Answers
1. [Risk] Gemini malformed JSON → **responseSchema + parse-retry** (Go unmarshal + retry 1× on bad JSON).
2. [Assumption] Bank account number regex aggressiveness → **Context-aware redaction** (long-digit runs only when preceded by "STK"/"tk"/phone-number context tokens; always redact phone + national ID; preserve amounts/Ref-IDs).
3. [Architecture] Tier 2/3 pending user-facing → **YELLOW + "community reports awaiting review" note**.
4. [Security] Appeal PII surface → **Text-claim appeals only, no image proof upload**.
5. [Risk] Free-tier rate limits on demo day → **Free-tier + cached-demo fallback**.
6. [Assumption] Tesseract VN fallback trigger → **Bar: <80% bbox-correct on 10 samples → PaddleOCR**.

#### Confirmed Decisions
- Gemini: structured output + Go unmarshal + 1 retry. | PII text: context-aware.
- Verdict UX: pending T2/T3 → transparent YELLOW + note. | Appeals: text-only (no claimant PII stored).
- AI tier: free-tier + masking, cached fallback for judging. | OCR: Tesseract w/ quantitative fallback bar.

#### Action Items
- [ ] Phase 1: confirm genai Go SDK `ResponseSchema` API + version (docs-seeker); implement parse-retry wrapper.
- [ ] Phase 4: context-aware redaction regex + tests preserving amounts/Ref-IDs.
- [ ] Phase 5: `blacklist_service` returns pending status → handler maps to YELLOW + note.
- [ ] Phase 6: remove image-proof upload; appeals = text claim + report_id only.
- [ ] Phase 7: add 10-sample bbox-accuracy check + PaddleOCR trigger at <80%.
- [ ] Phase 9: prepare cached demo verdict for rate-limit fallback.

#### Impact on Phases
- Phase 1 (+parse-retry, SDK confirm) · Phase 4 (context-aware) · Phase 5 (pending→YELLOW+note) · Phase 6 (text-only appeals) · Phase 7 (OCR bar) · Phase 9 (cached-demo fallback).

### Whole-Plan Consistency Sweep
- Re-read `plan.md` + `phase-01..09`. Propagated all 6 decisions. Removed stale "appeal + Gemini OCR proof" from Phase 6 (aligned text-only).
- No unresolved contradictions remain.

### Session 2 — 2026-08-01
**Trigger:** User — "Why I don't see plan for frontend and testing?"
**Root cause:** 9 phases cut vertically-by-feature → FE + tests embedded per slice, not visible as standalone.
**Decision:** Keep vertical slices (preserve per-phase demo / safe floor); add explicit **Frontend** + **Tests** subsections to every phase for visibility.
**Impact:** All 9 phases now surface FE deliverables + tests inline. No structural recut.
