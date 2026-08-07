# ShieldVN — Product Development Requirements (PDR)

> Trợ lý AI cảnh báo & phòng chống lừa đảo trực tuyến cho người Việt.
> Competition: **Google AI Riser Vietnam 2026** · Topic #5 Scam & Fraud · Deadline **2026-08-30**

## 1. Vision

Trở thành công cụ bỏ túi **Privacy-First** giúp người Việt nhận diện, phân tích và phòng tránh lừa đảo trực tuyến tại **"thời điểm vàng"** — ngay trước khi chuyển tiền hoặc bấm vào link lạ.

## 2. Problem

- Lừa đảo biến tướng liên tục: bill VietQR giả, tuyển CTV nạp tiền, mã QR dán đè, file `.apk` giả VNeID/Thuế.
- Công cụ tra cứu hiện tại bất tiện: nhập tay STK/SĐT/URL, **không đọc ngữ cảnh ảnh chụp màn hình**.
- Người dùng ngại gửi thông tin nhạy cảm (STK, CCCD, OTP) lên AI công cộng.

## 3. Target Users & Pain Points

| Persona | Pain |
|---|---|
| Người lớn tuổi | Không nhận biết bill/link giả; dễ hoảng khi bị hối thúc chuyển tiền |
| Sinh viên, người ít tiếp xúc công nghệ | Muốn kiểm tra nhanh nhưng không biết cách |
| Người giao dịch online thường xuyên | Cần xác nhận STK/lien kết nhanh, tin cậy |

## 4. Goals (MVP — hackathon)

1. Nhận ảnh chụp màn hình **hoặc** text/URL → ra nhận định **GREEN / YELLOW / RED** + bằng chứng tiếng Việt dễ hiểu, độ trễ < 2.5s (cold) / < 1.5s (warm).
2. **Privacy-First thật**: strip PII (SĐT/STK/CCCD) ở text (regex) và ảnh (OCR bbox mask) **trước** khi gửi Gemini.
3. Tra cứu **Blacklist 4 lớp** song song với AI, tích hợp vào bằng chứng.
4. "Warning card" chia sẻ qua Zalo/Facebook — viral loop.
5. Deploy Cloud Run (+deployment bonus).

## 5. Non-Goals (OUT of MVP)

- Native mobile app (Flutter/RN) — dùng PWA.
- Auto-update blacklist từ crawler chính phủ (dùng seed tay).
- Real payment integration, user accounts (guest-first).
- Heavy Playwright/k6 E2E + load test suite — defer (xem `08-test-plan.md`).

## 6. Competitive Analysis

| | **ShieldVN** | **nTrust** | **ChongLuaDao.vn** |
|---|---|---|---|
| Phân tích | Multi-modal AI (ảnh + text + QR + URL) | Tra cứu SĐT/STK, chặn cuộc gọi | Extension chặn URL |
| Đọc ngữ cảnh bill/tin nhắn | **Có** (font, layout, văn bản hối thúc) | Không | Không |
| PII protection | **Mask trước khi gửi AI** | Phụ thuộc input người dùng | Dựa URL |
| Viral | **Warning card** chia sẻ | Không | Báo cáo link xấu |

## 7. Success Metrics

| Metric | Target |
|---|---|
| Core flow (upload/paste → verdict) | hoạt động tin cậy, demo được |
| PII leak audit (text+image) | 0% PII plain-text tới Gemini |
| Latency warm / cold | < 1.5s / < 2.5s |
| Tier-1 blacklist recall trên sample | > 90% hit đã-seed |
| Demo video + pitch | < 2 phút, rõ ràng |

## 8. Scope Decisions (locked during brainstorm)

- Backend: Go + Gin · Frontend: Next.js (2 màn hình) · Masking: Tesseract OCR (`vie`) · DB: Firestore + Google Sheet (Tier-1 seed read-only) · AI: Gemini 2.5 Flash structured output · Deploy: Cloud Run.

## 9. Open Questions

1. Thêm Telegram bot (forward msg → verdict) để max convenience? (stretch)
2. Dùng paid Gemini Flash tier cho commitment "no-training" trên slide "real product", hay giữ free-tier + masking?
3. Defer hẳn Playwright/k6, hay giữ 1 Playwright smoke test?
