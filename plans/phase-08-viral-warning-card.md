---
phase: 8
title: "Viral Warning Card"
status: pending
effort: "1d"
priority: P3
dependencies: [2]
---

# Phase 8: Viral Warning Card

## Overview
"Generate quick warning card" button → Canvas renders PNG infographic (risk color + summary + QR to app) → share/download Zalo/Facebook. **Viral loop** to attract active users.

## Requirements
- Functional: on RED/YELLOW screen, click → generate 1080×1350 image → "Save image" + "Share" (Web Share API; fallback download).
- Non-functional: client-side Canvas render (no backend); QR encodes app link.

## Architecture
FE-only: `warning-card.tsx` takes verdict data → draws Canvas (color band, title, 1 main evidence, QR) → `canvas.toBlob('image/png')` → `navigator.share({files})` or `<a download>`. See `docs/02-design-guidelines.md` §4.

## Related Code Files
- Create: `frontend/components/warning-card.tsx`, `frontend/lib/qrcode.ts` (uses `qrcode` npm lib)
- Modify: `frontend/app/page.tsx` ("Generate warning card" button in result actions)

## Implementation Steps
1. Install `qrcode` npm (client QR gen).
2. `warning-card.tsx`: Canvas 1080×1350 — color band (RED/YELLOW), large title ("SCAM WARNING"), 1 main evidence line, QR corner (app link), app logo/name.
3. `toBlob` → Web Share API (`navigator.canShare({files})`); fallback `<a download>`.
4. Button appears on result (RED/YELLOW).
5. Test share to Zalo/FB (Web Share fallback download on desktop).

## Frontend
- `components/warning-card.tsx` (Canvas 1080×1350, risk color band, QR via `qrcode` lib).
- Web Share API (`navigator.share({files})`) + fallback `<a download>`; button in result actions (RED/YELLOW).

## Tests
- Manual UI-02: PNG correct risk color + has QR; render <1s.
- Share/download on Chrome/Safari + Zalo in-app; render Vietnamese diacritics.

## Success Criteria
- [ ] Click → generate 1080×1350 PNG correct risk color (UI-02 manual)
- [ ] Has QR to app
- [ ] Share/download works on mobile Chrome/Safari + Zalo browser
- [ ] Render < 1s

## Observability

Frontend-only (client-side Canvas); no backend instrumentation. Client-side error tracking is deferred (YAGNI) per [Phase 0 scope](./phase-00-observability-foundation.md).

## Risk Assessment
- **Web Share doesn't support file** (Zalo browser) → fallback download always available; test on Zalo in-app.
- **Canvas text VN font** → use system font; test render Vietnamese diacritics.
- **Priority P3** — if slipped, it's a viral nice-to-have feature, doesn't block core demo.
