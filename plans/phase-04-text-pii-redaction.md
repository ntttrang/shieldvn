---
phase: 4
title: "Text PII Redaction"
status: pending
effort: "1d"
priority: P2
dependencies: [3]
---

# Phase 4: Text PII Redaction

## Overview
Strip PII (phone number/bank account number/national ID) from text input **before** sending to Gemini using regex. The text part of the privacy claim. Fast, deterministic, unit-tested.

## Requirements
- Functional: `text_prompt` through **context-aware** sanitizer → always redact phone number + national ID; bank account number only redact when context present ("STK"/"tk"/account-number keywords…) to **preserve amounts/Ref-ID**. Replace placeholder before Gemini. **[Validation S1]**
- Non-functional: 0% plain-text PII in payload to Gemini (audit); don't break narrative needed for analysis.

## Architecture
Handler: raw text → `pii_text_sanitizer.Anonymize(text)` → clean text → Gemini. Regex: phone `0\d{9,10}`, bank account number `\b\d{8,19}\b`, national ID `\b\d{12}\b` (and old 9-digit national ID needs care to avoid false-positive with short bank account number).

## Related Code Files
- Create: `backend/internal/service/pii_text_sanitizer.go`, `backend/internal/service/pii_text_sanitizer_test.go`
- Modify: `backend/internal/handler/analyze_handler.go` (call sanitizer before Gemini)

## Implementation Steps
1. `pii_text_sanitizer.go`: struct with compiled regex; `Anonymize(text) string` replaces → `[PHONE_REDACTED]`, `[ACCOUNT_REDACTED]`, `[CCCD_REDACTED]`. (See `docs/first_idea.md` sample code.)
2. Table-driven unit tests: phone, bank account number, national ID, and **do not** redact harmless short numbers (6-digit OTP, small amounts). See `docs/08-test-plan.md` U-PII-*.
3. Wire into handler before `gemini_service.Analyze`.
4. Audit: log (staging) payload to Gemini, grep `0\d{9}`/`\d{8,19}` → 0 match on 10 samples.
5. Consider: if Gemini needs to know "has bank account number" to score, pass extra flag `has_account_number: true` (don't pass the value).

## Frontend
- None new. Finalize privacy banner text: *"The system automatically masks phone number/bank account number/national ID before AI analyzes"*.

## Tests
- `pii_text_sanitizer_test.go` table-driven: U-PII-01..04 + edge cases (preserve amounts/Ref-ID).
- Audit: log staging payload, grep `0\d{9}` / `\d{8,19}` → 0 match / 10 samples.

## Success Criteria
- [ ] Unit test U-PII-01..04 pass
- [ ] 0% phone number/bank account number/national ID plain-text to Gemini (audit)
- [ ] Verdict quality not degraded (narrative still has enough context)

## Observability

Redaction is privacy-sensitive by design. The Phase 0 logger gate guarantees no PII (raw phone/bank account number/national ID) is ever logged — only metadata keys. Optional: a `status=redacted` counter on the analyze span; do not log redaction content. The staging-payload audit grep complements this as a PII-leak check.

## Risk Assessment
- **False-positive redact amounts/OTP** → bank account number regex `\d{8,19}` can catch large amounts; mitigate: use context (token "STK"/"tk" preceding) or 8-19 range + boundary `\b`; test edge cases.
- **Lost context** → pass has_* entity flag instead of value.
