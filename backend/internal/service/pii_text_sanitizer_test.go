package service

import "testing"

func TestPIISanitizer_Anonymize(t *testing.T) {
    s := NewPIISanitizer()
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"phone", "Call me at 0912345678 please", "Call me at [PHONE_REDACTED] please"},
        {"national id 12", "My CCCD is 123456789012", "My CCCD is [CCCD_REDACTED]"},
        {"national id 9", "CCCD: 123456789", "CCCD: [CCCD_REDACTED]"},
        {"bank with context", "STK 12345678901234 transfer", "STK [ACCOUNT_REDACTED] transfer"},
        {"bank without context", "Account number 12345678 is shown", "Account number 12345678 is shown"},
        {"short number not redacted", "Your OTP is 123456", "Your OTP is 123456"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := s.Anonymize(tt.input)
            if got != tt.expected {
                t.Fatalf("expected %q, got %q", tt.expected, got)
            }
        })
    }
}
