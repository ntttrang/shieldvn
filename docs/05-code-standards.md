# ShieldVN — Code Standards

> YAGNI · KISS · DRY. File < 200 dòng (modularize khi vượt). Tên file kebab-case cho JS/TS, snake_case cho Go.

## 1. Repo Layout (intended)

```text
shieldvn-app/
├── backend/                      # Go (Gin)
│   ├── cmd/api/main.go           # entrypoint (≤80 dòng)
│   ├── internal/
│   │   ├── handler/              # Gin handlers — thin, gọi service
│   │   ├── service/              # business logic (gemini, pii, blacklist, moderation)
│   │   ├── store/                # Firestore + Sheets clients
│   │   ├── model/                # domain structs + DTOs
│   │   └── config/               # env loading
│   ├── Dockerfile                # multi-stage (Alpine + tesseract + vie)
│   ├── go.mod
│   └── go.sum
├── frontend/                     # Next.js (React + Tailwind, PWA)
│   ├── app/                      # 2 routes: upload page, result page
│   ├── components/               # risk-card, evidence-list, warning-card, uploader
│   ├── lib/                      # api client, image-compress, constants
│   └── public/                   # manifest, icons
├── docs/                         # bộ docs này
├── plans/                        # plans + reports
└── .github/workflows/            # CI (lint, build)
```

## 2. Go Conventions

- Package theo layer (`handler`, `service`, `store`, `model`). Handler mỏng — chỉ parse + gọi service.
- Error: `if err != nil { return fmt.Errorf("pii redact: %w", err) }`. Wrap với context. Không swallow.
- Context truyền qua mọi layer (`ctx context.Context` là tham số đầu).
- Structs DTO tách biệt domain model; tag JSON `snake_case`.
- Không log PII. Log chỉ metadata (request id, latency, verdict level).
- Tests: file `_test.go` cùng package. Table-driven. `testify` OK.

```go
// Service signature mẫu
func (s *AnalysisService) Analyze(ctx context.Context, in AnalysisInput) (AnalysisResult, error)
```

## 3. Frontend Conventions (Next.js)

- Function components + hooks. TypeScript strict.
- 2 route duy nhất MVP: `/` (upload), `/result` (hoặc stream inline). Không routing phức tạp.
- Tailwind utility-first; không CSS-in-JS. Mobile-first.
- Component file < 200 dòng; tách `risk-card`, `evidence-list`, `warning-card`, `uploader`.
- API client ở `lib/api.ts` — single source of truth gọi backend.
- Nén ảnh client-side (`lib/image-compress.ts`) trước upload (< 1.5MB).

## 4. Naming

- Files: kebab-case (FE) — `risk-card.tsx`; snake_case (Go) — `pii_sanitizer.go`.
- Identifiers: Go PascalCase (exported) / camelCase; TS camelCase, types PascalCase.
- Tên tự документа — dài OK (`pii_image_masker.go` tốt hơn `mask.go`).

## 5. Comments

- Giải thích **why**, không giải thích **what**. Không reference plan/phase/finding code.
- Tốt: `// Tesseract TSV cho bbox word; regex match phone/STK/CCCD rồi mask`
- Tệ: `// per Phase 6 requirement`

## 6. Git / Commits

- Conventional: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`. Không AI references.
- Branch off `main`. Commit theo phase (xem memory: commit-per-phase).
- **Không commit** secrets / `.env` / API keys / Firestore svc account.

## 7. Quality Gates (pre-push)

- Go: `go build ./...` + `go vet ./...` + `go test ./...` pass.
- FE: `npm run build` (tsc strict) pass.
- Không fix-that-breaks-tests. Test failing → fix root cause, không mock để qua.

## 8. Testing Philosophy (hackathon)

- Ưu tiên Go unit test cho logic deterministic: PII regex, tier-scoring, response-schema parse.
- Integration: 1 smoke test gọi Gemini với prompt cố định.
- E2E heavy (Playwright/k6) → defer (xem `08-test-plan.md`). Manual matrix cho demo.
