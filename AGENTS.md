# AGENTS.md — AI Agent Guidance & Codebase Reference for ShieldVN

> **ShieldVN** — Privacy-First AI Scam & Fraud Warning Assistant (Google AI Riser Vietnam 2026)  
> This file provides AI models, coding assistants, and human developers with an explicit reference for codebase architecture, coding standards, key flows, and development practices.

---

## 1. Project Overview & Architecture

ShieldVN is a **privacy-first scam detection platform** tailored for Vietnamese users. It accepts screenshots, pasted text, or suspicious URLs, redacts all PII (Personally Identifiable Information) on RAM before external transmission, executes parallel Gemini 2.5 Flash AI analysis & 4-tier Firestore blacklist lookups, and returns structured Vietnamese verdicts (**GREEN / YELLOW / RED**).

```
                      [ PWA Frontend: Next.js + Tailwind (2 Screens) ]
                                            │
                                            │ multipart/form-data (image) / JSON (text)
                                            ▼
                    [ Go 1.22 / Gin API Gateway (Google Cloud Run · RAM-only) ]
                                            │
                     ┌──────────────────────┴──────────────────────┐
                     ▼                                             ▼
            [ 1. Text PII Redaction ]                    [ 2. Image PII Masking ]
               Context-aware Regex                          Tesseract OCR (vie)
                     │                                      BBox Regex Masking
                     └──────────────────────┬──────────────────────┘
                                            │ Clean RAM Payload
                                            ▼
                               [ Parallel Goroutines ]
                                ├── (a) Gemini 2.5 Flash (Structured Output)
                                └── (b) 4-Tier Firestore & Sheets Blacklist Lookup
                                            │
                                            ▼
                                [ Verdict Merger Engine ]
                              (GREEN / YELLOW / RED + Evidence)
```

---

## 2. Directory Layout & Module Map

```
shieldvn/
├── backend/                       # Go 1.22 + Gin HTTP Server
│   ├── cmd/api/main.go            # Application entrypoint (setup, router, graceful shutdown)
│   ├── internal/
│   │   ├── config/                # Environment variables & runtime configuration
│   │   ├── handler/               # Thin HTTP handlers (Gin controllers, middleware)
│   │   ├── service/               # Core business logic (Gemini, PII, Blacklist, Moderation)
│   │   ├── store/                 # Storage abstraction (Cloud Firestore & Google Sheets API)
│   │   ├── model/                 # Domain models, DTO structs, JSON schemas
│   │   └── observability/         # Structured logging (slog), Request ID, OpenTelemetry
│   ├── Dockerfile                 # Multi-stage Alpine container (Go + Tesseract OCR vie)
│   ├── go.mod / go.sum            # Dependencies
│   └── .env.example               # Environment template
├── frontend/                      # Next.js (React + Tailwind PWA, 2 screens)
│   ├── app/                       # Next.js App Router (upload page, result page)
│   ├── components/                # UI Components (risk-card, evidence-list, warning-card, uploader)
│   ├── lib/                       # API client, client-side image compression, constants
│   └── public/                    # PWA manifest, icons, static assets
├── docs/                          # Core product specifications & architecture docs
│   ├── 01-project-overview-pdr.md # Product Development Requirements (PDR) & Goals
│   ├── 02-design-guidelines.md    # UI/UX design specifications & components
│   ├── 03-system-architecture.md  # Detailed system architecture & data flow
│   ├── 04-codebase-summary.md     # Codebase mapping & component responsibilities
│   ├── 05-code-standards.md       # Engineering standards, Go & TS conventions
│   ├── 06-project-roadmap.md      # Implementation roadmap & risk mitigation
│   ├── 07-deployment-guide.md     # GCP Cloud Run deployment instructions
│   ├── 08-test-plan.md            # Testing matrix & strategy
│   └── 09-ci-pipeline.md          # CI/CD workflow definition
├── plans/                         # Vertical slice implementation phase plans
│   ├── plan.md                    # Master plan & phase directory
│   └── phase-00..09-*.md          # Specific feature phase execution details
└── .agents/                       # Custom skills & agent workspace configurations
```

---

## 3. Core Coding Standards & Guidelines

### Architectural Principles
- **YAGNI · KISS · DRY**: Keep implementation minimal, direct, and non-redundant.
- **Vertical Slice Execution**: Features are implemented end-to-end per phase. Phase completion must leave the application in a demoable, working state.
- **File Length Hard Limit**: Keep files **< 200 lines**. Tightly decouple into helper files or modular packages whenever line counts grow.

### Naming Conventions
- **Go Backend**:
  - File names: `snake_case.go` (e.g., `analysis_service.go`, `requestid_test.go`).
  - Identifiers: `PascalCase` for exported symbols; `camelCase` for internal variables/functions.
  - Descriptive names over short abbreviations (`pii_sanitizer.go` over `mask.go`).
- **TypeScript / Next.js Frontend**:
  - File names: `kebab-case.tsx` or `kebab-case.ts` (e.g., `risk-card.tsx`, `image-compress.ts`).
  - Identifiers: `camelCase` for variables/functions; `PascalCase` for React components/Types.

### Error Handling & Context
- **Context First**: Always pass `ctx context.Context` as the first parameter across layers (`handler` → `service` → `store`).
- **Error Wrapping**: Wrap errors with meaningful context using `fmt.Errorf("component operation: %w", err)`. **Never swallow errors silently**.
- **No Defensive Hacks**: Fix root causes of failures. Do not mask errors by returning dummy zero-value fallbacks or commenting out broken code.

### Logging & Privacy Guardrails
- **Structured Logging**: Use `slog` with JSON handler (`slog.Info`, `slog.Error`).
- **ZERO PII LOGGING (STRICT MANDATE)**:
  - **NEVER** log bank account numbers (STK), phone numbers, national IDs (CCCD), user names, or unmasked text inputs.
  - Log **only** structural metadata: `request_id`, HTTP status, execution latency, error messages, and final verdict level (`GREEN/YELLOW/RED`).

---

## 4. Key Technical Workflows

### 1. PII Sanitization Engine
- **Text PII Redaction**:
  - Context-aware regex matching for bank accounts (STK), phone numbers, and national IDs (CCCD).
  - Must preserve transfer amounts, transaction reference IDs, and generic numerical data.
- **Image PII Masking**:
  - Processed in-memory (RAM-only, zero disk persistence).
  - Tesseract OCR CLI (`vie` traineddata) TSV mode generates word bounding boxes (`left`, `top`, `width`, `height`).
  - Regex detects sensitive tokens; Go `image/draw` draws solid rectangle masks before payload transmission to Gemini.

### 2. Structured Gemini 2.5 Flash Contract
- Go SDK: `google.golang.org/genai`.
- Enforces JSON output mode (`ResponseMIMEType: "application/json"`) with structured schema:
  ```json
  {
    "risk_score": "GREEN|YELLOW|RED",
    "confidence_score": 0.95,
    "detected_patterns": ["impending urgency", "fake QR code"],
    "evidence": ["Vietnamese evidence description"],
    "recommendations": ["Do not transfer money"],
    "extracted_entities": {
      "bank_account": "1234567890",
      "phone_number": "0912345678",
      "url": "https://suspicious-domain.com"
    }
  }
  ```
- Backend incorporates a 1-retry fallback mechanism for JSON unmarshaling resilience.

### 3. 4-Tier Blacklist Scoring Engine
- **Tier 1 (Official)**: Authority & bank blacklists (auto-blacklist `RED`). Seeded via read-only Google Sheets API for transparent judging.
- **Tier 2 (Trusted Community)**: High-reputation community reporters (threshold hit -> `YELLOW` pending).
- **Tier 3 (Anonymous)**: User reports (marked as pending; never auto-labels 100% scam).
- **Tier 4 (Appeals)**: Account holder dispute resolution flow (text claims only, no image PII storage).

### 4. Observability & Telemetry
- `slog` structured JSON logs sent to stdout.
- Gin middleware injects `X-Request-ID` into request headers and OpenTelemetry span context.
- OpenTelemetry exporter connects to GCP Cloud Trace / Cloud Logging when running on Cloud Run, falling back to clean no-op providers locally or in CI without requiring GCP credentials.

---

## 5. Development Setup & Verification

### Running Backend Locally
```bash
cd backend
cp .env.example .env          # Set GEMINI_API_KEY and configuration
go mod tidy
go run ./cmd/api              # Starts server on :8080
```

### Running Frontend Locally
```bash
cd frontend
npm install
echo "NEXT_PUBLIC_API_URL=http://localhost:8080" > .env.local
npm run dev                   # Starts dev server on :3000
```

### Quality Gates (Pre-Push Checks)
Before pushing or marking any task complete, verify that the project passes all checks:

```bash
# Go Backend Verification
cd backend
go build ./...
go vet ./...
go test ./...

# Next.js Frontend Verification
cd frontend
npm run build
```

---

## 6. Commit & Documentation Policy

- **Commit Messages**: Follow Conventional Commits format (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`).  
  - *Example*: `feat: implement PII regex text sanitizer`
  - *Rule*: **Do NOT** include LLM model references or internal task phase codes in commit messages.
- **Comments**: Explain **why** decisions were made, not **what** the code does. Avoid referring to internal prompt phase numbers in source comments.
