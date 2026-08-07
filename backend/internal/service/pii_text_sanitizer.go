package service

import (
    "regexp"
    "strings"
)

// PIISanitizer provides context‑aware redaction of personal identifiers from text.
type PIISanitizer struct {
    phoneRegex   *regexp.Regexp
    cccd12Regex  *regexp.Regexp
    cccd9Regex   *regexp.Regexp
    bankRegex    *regexp.Regexp
    bankKeywords []string
}

// NewPIISanitizer constructs a sanitizer with compiled regexes.
func NewPIISanitizer() *PIISanitizer {
    return &PIISanitizer{
        phoneRegex:   regexp.MustCompile(`0\d{9,10}`),
        cccd12Regex: regexp.MustCompile(`\b\d{12}\b`),
        cccd9Regex:  regexp.MustCompile(`\b\d{9}\b`),
        bankRegex:   regexp.MustCompile(`\b\d{8,19}\b`),
        bankKeywords: []string{"stk", "tk", "tài khoản"},
    }
}

// Anonymize redacts phone numbers, national IDs, and optionally bank account numbers.
func (s *PIISanitizer) Anonymize(text string) string {
    // Redact phone numbers.
    sanitized := s.phoneRegex.ReplaceAllString(text, "[PHONE_REDACTED]")
    // Redact national ID (12‑digit) and legacy 9‑digit.
    sanitized = s.cccd12Regex.ReplaceAllString(sanitized, "[CCCD_REDACTED]")
    sanitized = s.cccd9Regex.ReplaceAllString(sanitized, "[CCCD_REDACTED]")

    // Determine if banking context exists.
    lower := strings.ToLower(sanitized)
    hasBankContext := false
    for _, kw := range s.bankKeywords {
        if strings.Contains(lower, kw) {
            hasBankContext = true
            break
        }
    }
    if hasBankContext {
        sanitized = s.bankRegex.ReplaceAllString(sanitized, "[ACCOUNT_REDACTED]")
    }
    return sanitized
}
