# ShieldVN — Project Roadmap

> Deadline **2026-08-30** (~4 tuần từ 2026-07-31). Solo build. Nguyên tắc: **mỗi phase kết thúc = demo được; không để phase sau chặn phase trước.**

## Build Order (vertical slice — safe floor tại Phase 2)

| # | Phase | Days | Deliverable | Trạng thái |
|---|-------|------|-------------|-----------|
| 0 | Skeleton + text verdict | 1 | `go mod init`, Gin, Gemini text call + responseSchema, Next.js stub | ⬜ Pending |
| 1 | Image verdict | 2–3 | Upload ảnh → Gemini vision → GREEN/YELLOW/RED | ⬜ Pending |
| 2 | **Tier-1 lookup** | 4–5 | Seed Sheet + Firestore lookup song song, merge vào evidence | ⬜ Pending |
| 3 | Text PII redact | 6 | Regex strip SĐT/STK/CCCD trước send | ⬜ Pending |
| 4 | 4-tier model + report | 7–9 | Firestore schema, "report" flow, tier-scoring, seed T2–T4 | ⬜ Pending |
| 5 | Moderation + appeal | 10–12 | Moderator console, approve/reject, appeal submit | ⬜ Pending |
| 6 | Image PII mask | 13 | Tesseract TSV → bbox → mask regions | ⬜ Pending |
| 7 | Warning card | 14 | Canvas render PNG shareable Zalo/FB | ⬜ Pending |
| 8 | Deploy + polish | 15–17 | Cloud Run, PWA, manual test matrix | ⬜ Pending |
| 9 | Demo + submit | 18–20 | Video <2 phút + pitch deck + nộp bài | ⬜ Pending |

**Safe floor:** dừng tại Phase 2 (~ngày 5) vẫn có demo core thắng được. Phase 3–9 là upside.

## Milestones

- **M1 — Core demoable (ngày ~5):** text+image → verdict + blacklist hit. **Bắt buộc đạt.**
- **M2 — Privacy + 4-tier (ngày ~13):** masking + reporting + moderation.
- **M3 — Shipped (ngày ~17):** deploy Cloud Run + PWA.
- **M4 — Submit (ngày ~20):** video + hồ sơ.

## Risks & Mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| Solo + full scope 4 tuần | Cao | Vertical slice; safe floor Phase 2 |
| Tesseract VN accuracy | TB | Ảnh digital → accuracy cao; fallback PaddleOCR sidecar |
| Next.js tốn thời gian | TB | Hard-cap 2 màn, Tailwind only, không design system |
| Free-tier Gemini train data | TB (trust) | Masking pre-send; nhắc paid tier ở slide product |
| Heavy E2E suite | TB | Defer; manual matrix + Go unit tests |

## Out-of-Scope (MVP)

Native app · crawler auto-update · user accounts/payment · Playwright/k6 full suite.

## Changelog

- **2026-07-31** — Roadmap tạo từ brainstorm. Scope locked: full doc + 4-tier, vertical-slice order.
