# ShieldVN — Test Plan

> Hackathon-realistic. **Ưu tiên Go unit test (deterministic) + manual matrix (demo).** Defer Playwright/k6 full suite — không đáng đổi thời gian build (flag open).

## 1. Strategy

| Layer | Phạm vi | Cách |
|---|---|---|
| Unit (Go) | PII regex, tier-scoring, response-schema parse, TSV bbox parse | `go test`, table-driven, `testify` |
| Integration (Go) | 1 smoke Gemini call (prompt cố định), Firestore read/write 1 bản ghi | `go test -tags=integration`, skip trong CI nếu no creds |
| API contract | `/api/v1/analyze` shape đúng `ScamAnalysisResponse` | `httptest` |
| Manual matrix | toàn bộ kịch bản demo — xem §3 | checklist tay trên staging |
| E2E heavy (Playwright/k6) | **DEFER** — open question giữ 1 smoke test (tùy thời gian) | optional |

**Exit criteria (pre-submit):** 100% manual matrix An+Nhiễm pass · 0% PII lọt log · latency warm < 1.5s.

## 2. Unit Test Cases (Go — must have)

| ID | Mô tả | Input | Expected |
|---|---|---|---|
| U-PII-01 | Redact phone text | `"Gọi 0912345678"` | `"Gọi [PHONE_REDACTED]"` |
| U-PII-02 | Redact STK text | `"STK 1903123456789"` | `"STK [ACCOUNT_REDACTED]"` |
| U-PII-03 | Redact CCCD | `"CCCD 079200012345"` | masked |
| U-PII-04 | Không redact số ngắn vô hại | `"mã OTP 6 số"` | unchanged |
| U-MASK-01 | TSV parse → bbox đúng | TSV sample | correct `[]image.Rectangle` |
| U-MASK-02 | Mask region sinh ảnh có box đen | image + rects | box pixels = black |
| U-TIER-01 | Tier 1 auto-blacklist | entity trong seed | `isBlacklisted=true` |
| U-TIER-02 | Tier 2 cần 2–3 trùng | 2 reports | `status=pending` |
| U-TIER-03 | Tier 3 anonymous không label 100% | anon report | `pending`, no definite label |
| U-SCHEMA-01 | Parse Gemini JSON đúng | fixture JSON | `AnalysisResult` struct filled |
| U-SCHEMA-02 | Reject malformed | `{}` | error, không crash |

## 3. Manual Test Matrix (demo-critical)

### Nhóm AN — Phân tích
| ID | Kịch bản | Expected |
|---|---|---|
| AN-01 | Upload bill VietQR giả (font lệch) | RED + evidence "phông chữ lệch" |
| AN-02 | Paste tin CTV "nạp tiền chốt đơn" + link TG | YELLOW/RED + "không ứng tiền" |
| AN-03 | Upload bill VCB thật | GREEN (không báo nhầm) |
| AN-04 | Paste URL phishing đã seed | RED + "URL trong blacklist" |

### Nhóm SEC — Privacy
| ID | Kịch bản | Expected |
|---|---|---|
| SEC-01 | Text có SĐT+STK | log/payload tới Gemini = placeholder, 0 plain PII |
| SEC-02 | Ảnh chứa STK/CCCD | region mask trước send; Gemini không đọc được vùng che |
| SEC-03 | Upload 10 lần liên tiếp | zero disk write, RAM ổn định |

### Nhóm TIER — Blacklist
| ID | Kịch bản | Expected |
|---|---|---|
| TIER-01 | STK trong Tier-1 seed | verdict escalate + evidence "Tier 1" |
| TIER-02 | Submit report (Tier 2–3) | vào pending, điểm uy tín cập nhật |
| TIER-03 | Moderator approve/reject | status đổi, reflect ở lookup |
| TIER-04 | Appeal submit + proof | tạo appeal, moderator resolve |

### Nhóm UI
| ID | Kịch bản | Expected |
|---|---|---|
| UI-01 | Ảnh 10MB | client nén < 1.5MB, không 413 |
| UI-02 | Tạo warning card (RED) | PNG 1080×1350, có QR |
| UI-03 | Mobile Safari/Chrome + Zalo browser | layout không bể, nút ≥48px |

### Nhóm PERF
| ID | Mục tiêu |
|---|---|
| PERF-01 | Warm latency < 1.5s (20 req liên tiếp) |
| PERF-02 | Cold start < 2.5s |

## 4. PII Leak Audit (pre-submit)

- Bật log payload tới Gemini ở staging → grep 50 sample cho `0\d{9}`, `\d{8,19}` → **0 match**.
- Verify image gửi đi có box đen vùng PII (visual spot-check 5 ảnh).

## 5. Deferred / Open

- **Playwright E2E + k6 load**: defer. Nếu còn thời gian (Phase 8+) → thêm 1 Playwright smoke (AN-01 end-to-end) + 1 k6 50-VU smoke. Không bắt buộc exit.
- Cross-browser đầy đủ: Chrome Android + Safari iOS + Zalo in-app (manual).

## 6. Tools

- Go: `testing`, `httptest`, `testify`.
- Manual: checklist trên staging URL.
- (Optional) Playwright, k6 — chỉ nếu thời gian cho phép.
