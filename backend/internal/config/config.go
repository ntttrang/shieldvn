package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	Port            string
	AllowedOrigin   string
	GeminiAPIKey    string
	GeminiModelName string
	LogLevel        string
	GCPProjectID    string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	// Attempt to load .env file, ignore error if it doesn't exist
	_ = godotenv.Load()

	return &Config{
		Port:            getEnv("PORT", "8080"),
		AllowedOrigin:   getEnv("ALLOWED_ORIGIN", "http://localhost:3000"),
		GeminiAPIKey:    os.Getenv("GEMINI_API_KEY"),
		GeminiModelName: getEnv("GEMINI_MODEL_NAME", "gemini-2.5-flash-lite"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		GCPProjectID:    os.Getenv("GOOGLE_CLOUD_PROJECT"),
	}
}

// ParsedLogLevel converts the string log level to slog.Level.
// Defaults to info on unrecognized input and logs a warning.
func (c *Config) ParsedLogLevel() slog.Level {
	switch strings.ToLower(c.LogLevel) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		slog.Warn("unrecognized LOG_LEVEL, defaulting to info", "raw", c.LogLevel)
		return slog.LevelInfo
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
