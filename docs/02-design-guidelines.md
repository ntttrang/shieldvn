# ShieldVN — Design Guidelines

> Mobile-first. Đối tượng chính: người lớn tuổi, người ít tiếp xúc công nghệ, thường vào qua **Zalo in-app browser**. Ưu tiên **to, rõ, ít bước, kết quả trước**.
> Privacy-first AI anti-scam tool — UI tiếng Việt. Bản mockup trực quan: [`../assets/mockups/shieldvn-ui-mockup.html`](../assets/mockups/shieldvn-ui-mockup.html) (mở bằng trình duyệt).

## 1. Principles

1. **1-tap analyze** — mở app → tải ảnh/paste → 1 nút "Kiểm tra ngay".
2. **Verdict trước, chi tiết sau** — băng màu lớn + nhãn lớn đập vào mắt trước, evidence nằm dưới.
3. **Privacy明示** — banner cố định: *"Hệ thống tự che SĐT/STK/CCCD trước khi AI phân tích"*.
4. **Ngôn ngữ bình dân** — tránh thuật ngữ; "Nguy hiểm" thay "High risk", "Cẩn thận" thay "Suspicious".
5. **Một việc / màn hình** — không nav phức tạp, không hover-only, không thông báo lỗi tự ẩn.
6. **Hard rule: màu đèn giao thông 🟢🟡🔴 dành RIÊNG cho verdict.** Brand + CTA = **teal** (theo `app-icon.png` / `logo-icon.png`) — tách rõ khỏi 3 màu verdict, nút "Kiểm tra" không tranh chấp thị giác với kết quả RED.

## 2. Color — Design Tokens (CSS variables)

> **Nguồn palette:** trích từ `app-icon.png` (nền mint `#F0F8F0` + gradient khiên `#E6F7F5→#D4EFEF`) và `logo-icon.png` (viên khiên deep-teal `#006B6B`). Brand chính thức = **deep teal `#006B6B`**, không phải xanh dương. Trạng thái đèn giao thông (xanh/lục/vàng/đỏ) lấy ngay từ 3 chấm status của logo.

### 2.1 Brand / surface

| Token | Value | Role |
|---|---|---|
| `--brand` | `#006B6B` | Brand + **primary CTA** (deep teal — chữ trắng ~5.5:1 ✓) |
| `--brand-600` | `#00766C` | CTA hover/active (chữ trắng ~4.6:1 ✓) |
| `--brand-tint` | `#E6F7F5` | Privacy banner, focus halo, nền upload zone hover (mint gradient trong khiên) |
| `--ink` | `#0F172A` | Heading + body text |
| `--muted` | `#475569` | Secondary text (4.5:1 ✓) |
| `--bg` | `#F0F8F6` | App / page background (faint mint, theo nền app-icon) |
| `--surface` | `#FFFFFF` | Cards, app surface |
| `--line` | `#E2E8F0` | Borders, dividers |

### 2.2 Risk system — 3 verdicts (contrast-safe)

> **Note:** Hai cách đạt AA cho badge/băng:
> - **GREEN / RED** → fill đậm + **chữ trắng** (`#15803D` = 4.5:1, `#DC2626` = 4.8:1). Hex docs cũ (`#16A34A`, `#FF0000`) trượt AA với chữ trắng nên chỉ giữ làm accent dot.
> - **YELLOW** → fill **vàng gold `#FFC107`** (đúng vàng logo) + **chữ đậm `#5A3E00`** (~6:1 ✓) — kiểu biển báo cảnh báo. (Gold + chữ trắng chỉ 2.1:1, sai AA.)

| Level | Solid fill | Chữ trên fill | Soft bg tint | Border | Verdict label |
|---|---|---|---|---|---|
| 🟢 GREEN | `#15803D` (4.5:1 ✓) | trắng | `#E8F8EE` | `#BBF7D0` | AN TOÀN |
| 🟡 YELLOW | `#FFC107` (gold logo) | `#5A3E00` (~6:1 ✓) | `#FFFBEB` | `#FDE68A` | CẨN THẬN |
| 🔴 RED | `#DC2626` (4.8:1 ✓) | trắng | `#FEF2F2` | `#FECACA` | NGUY HIỂM |

Accent dots (chỉ trang trí, không phải text-bg) = màu logo sống động: GREEN `#00A651` · YELLOW `#FFC107` · RED `#FF3B30`.

**Quy tắc dùng màu:** badge/băng = fill đặc + icon SVG + chữ (trắng cho xanh/lục/đỏ, đậm cho vàng); vùng thẻ evidence = tint nhạt + border. Không bao giờ truyền nghĩa **chỉ bằng màu** — luôn kèm icon + chữ (rule `color-not-only`).

## 3. Typography

- **Font:** `Be Vietnam Pro` (Google Fonts) — 1 họ duy nhất, weight `400 / 500 / 600 / 700`. Render dấu tiếng Việt tốt nhất; thay font `Arial`/Geist mặc định trong `globals.css`. (Đọc thêm: docs `09-ci-pipeline` không liên quan; import ở `globals.css`.)
- **Type scale (mobile-first, phóng to cho người lớn tuổi):**

| Role | Size / weight | Dùng cho |
|---|---|---|
| Verdict label | **28px / 700** | "NGUY HIỂM" |
| H1 | 24px / 700 | Tiêu đề trang |
| Section H | 18px / 700 | "Tại sao?", "Bạn nên làm gì" |
| Body-lg | 18px / 500 | Guidance line dưới verdict |
| Body | 17px / 400 | Evidence, mô tả |
| Button | 17px / 600 | CTA |
| Label | 14px / 500 | Confidence %, microcopy |

Line-height body 1.5; số liệu/số % dùng `font-variant-numeric: tabular-nums`.

## 4. Spacing · Shape · Motion · Icons

- **Spacing scale:** `4 / 8 / 12 / 16 / 24 / 32 / 48` (nhịp 8px).
- **Radius:** `12px` card · `10px` button/input · `9999px` pill/badge.
- **Shadow (1 cấp duy nhất):** `0 1px 2px rgba(15,23,42,.06), 0 8px 24px rgba(15,23,42,.06)`.
- **Touch target:** ≥ **48×48px** mọi phần tử tương tác (§6). `cursor-pointer` mọi element click.
- **Motion:** 150–250ms ease-out (enter); exit ≈ 60% enter. `prefers-reduced-motion` → radar loading trở thành pulse tĩnh. Không animate `width/height/top/left`.
- **Icons:** **Lucide SVG** (`shield, lock, image-plus, triangle-alert, circle-alert, circle-check, share-2, arrow-left, rotate-ccw, x, help-circle`). **Không dùng emoji** làm icon cấu trúc — verdict = icon + màu (vd `triangle-alert` đỏ), không phải ký tự 🟢🟡🔴.

## 5. Components (spec)

| Component | Spec |
|---|---|
| **Primary button** | fill `--brand`, chữ trắng 17/600, 48px+ cao, full-width mobile, disabled @ opacity .5 + `cursor-not-allowed` |
| **Risk pill / băng verdict** | fill đặc verdict, chữ trắng, pill hoặc băng full-width, icon + label |
| **Upload zone** | dashed `--line` 2px, `--surface`, ~180px, tap → file picker (+ camera capture), có trạng thái thumbnail sau khi chọn |
| **Privacy banner** | nền `--brand-tint`, icon lock, cố định |
| **Evidence card** | `--surface`, 1px `--line`, radius 12, bullet list 17px, icon dấu chấm/trước mỗi mục |
| **Recommendation** | checkmark Lucide `check` + text; mục NEGATIVE (RED) dùng emphasis in hoa chữ đầu |
| **Confidence bar** | track `--line`, fill màu verdict, % bên trái (tabular-nums) |
| **Input/textarea** | 1px `--line`, focus ring 2px `--brand`/30, min 48px |

## 6. Screen 1 — Upload (`/`)

**Empty:** Header (brand trái + icon help phải) → H1 *"Kiểm tra lừa đảo trong 10 giây"* → **Upload zone** (action chính) → divider *"HOẶC"* → textarea *"Dán tin nhắn / link đáng ngờ"* → **Privacy banner** → **CTA "Kiểm tra ngay"** (disabled tới khi có input) → microcopy trust *"Đã kiểm tra 12.480 tin nhắn"*.

**Has-input:** Upload zone thu thành row thumbnail `[img] bill.jpg ✕`; CTA active `--brand`.

**Loading** (~3–8s, thay vùng form): radar quét trên khiên + *"Đang phân tích…"* + các bước tick dần: ✓ Đã che thông tin nhạy cảm → Đang đối chiếu danh sách → AI đang đọc ảnh. CTA spinner + disabled.

**Error:** card đỏ nhạt dưới CTA — nguyên nhân + đường phục hồi (`"Không kết nối được. Thử lại nhé."`) + nút **Thử lại**. Không auto-dismiss.

## 7. Screen 2 — Result (`/result`)

Cùng khung cho 3 verdict; chỉ băng màu + label + guidance đổi.

```
┌──────────────────────────────────┐
│  ← (back)  ShieldVN               │
├──────────────────────────────────┤
│ ▓▓ [icon]  NGUY HIỂM  … 92% ▓▓▓░ │  ← băng full-width (verdict fill)
│ ▓ Không chuyển tiền.          ▓   │     guidance 18/500 trắng
│ ▓ Độ tin cậy 92% ▓▓▓▓▓▓▓▓░░░ ▓   │     confidence bar
│ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ │
│  Tại sao lại vậy?                 │  Evidence card (bullet, tiếng Việt)
│   • Số TK 98XX… trong blacklist    │
│   • Ép đóng "phí mở kênh"          │
│   • Tên người nhận trùng 3 báo cáo │
│  Bạn nên làm gì                    │  Recommendations (checkmark)
│   ✓ DỪNG — không chuyển tiền        │
│   ✓ Chặn & báo cáo số này          │
│   ✓ Gọi 156 (CSKH ngân hàng)       │
│  [share] Tạo thiệp cảnh báo │ Kiểm tra tiếp │
│  (muted) Kết quả sai? Báo cáo lại  │  appeal link (non-destructive)
└──────────────────────────────────┘
```

- **🔴 RED — NGUY HIỂM**: *"Không chuyển tiền."*
- **🟡 YELLOW — CẨN THẬN**: *"Tìm hiểu thêm trước khi chuyển tiền."* — evidence theo "nghi vấn", recommendation "Xác minh lại tên người nhận qua tổng đài".
- **🟢 GREEN — AN TOÀN**: *"Chưa phát hiện dấu hiệu lừa đảo."* — evidence ngắn (đã kiểm tra gì), recommendation gọn 1 dòng ("Vẫn cẩn thận với giao dịch lạ").

## 8. Warning Card (viral)

Canvas → PNG, **1080×1350** (portrait, tối ưu MXH). Render thành bottom-sheet khi bấm "Tạo thiệp cảnh báo".

Layout: băng màu verdict + headline tóm tắt vụ + 1 evidence chính + độ tin cậy + dòng CTA *"Kiểm tra trước khi chuyển tiền"* + **QR về app** cạnh brand ShieldVN. Nút: **Lưu ảnh** (download) + **Chia sẻ** (Web Share API → Zalo/Facebook; fallback download).

## 9. Responsive & PWA

- **Mobile (375–430):** single-column full-bleed, padding 16.
- **Desktop (≥768):** nền `--bg`, cột app **căn giữa max-width 480px** thành `--surface` + shadow → giữ cảm giác app.
- **Zalo in-app browser:** test riêng; không fixed bottom bar đè chrome Zalo; mọi action chạm được.
- **PWA:** standalone display, installable, splash/icon `--brand`.

## 10. Accessibility (người lớn tuổi)

- Contrast AA tối thiểu (badge chữ trắng trên fill đậm — bảng §2.2 đã verify).
- Touch target ≥ 48×48px. Không hover-only.
- Icon luôn kèm text label cho action chính. Mọi icon có `aria-label`.
- Form: label hiện, lỗi gần field, error dùng `role="alert"`/`aria-live`, focus-auto trường lỗi đầu tiên.
- `prefers-reduced-motion` tắt radar; Dynamic Type không vỡ layout.

## 11. Out of Scope (MVP)

Dark mode · multi-language toggle (chỉ tiếng Việt) · onboarding flow · dashboard. **Nhưng** mọi màu triển khai bằng **semantic CSS token** để mở rộng dark mode sau mà không refactor.

## 12. Decisions (locked)

1. Emoji cấu trúc (🛡️🔒🟢🟡🔴) → **Lucide SVG**. Giữ ẩn dụ đèn giao thông = icon + màu.
2. **Brand = deep teal `#006B6B`** (theo `app-icon.png` / `logo-icon.png`), không phải xanh dương. Palette đèn giao thông dành riêng verdict; brand tách riêng. Bác gợi ý CTA cam (tranh cãi với YELLOW) và CTA xanh dương cũ (không khớp logo).
3. Badge GREEN/RED = fill đậm + chữ trắng; YELLOW = gold `#FFC107` + chữ đậm `#5A3E00` (biển cảnh báo). Tất cả đạt AA theo §10.
4. Font = **Be Vietnam Pro**.
5. Dark mode: out of scope MVP; token hóa để mở rộng. Stub Next.js hiện có `dark:` class → gỡ bỏ cho MVP.

## 13. Open Questions

- Stub Next.js hiện **chỉ text** (chưa upload ảnh). PDR/README coi **upload ảnh là use case "golden moment" chính**. Có đưa upload ảnh vào scope MVP web không? (ánh xạ Roadmap Phase 1, không phải Phase 0.)
- Đếm "Đã kiểm tra N tin nhắn" (social proof) — số thật từ Firestore hay hardcode seed cho MVP?
