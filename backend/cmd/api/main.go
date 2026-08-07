package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shieldvn-backend/internal/config"
	"shieldvn-backend/internal/handler"
	"shieldvn-backend/internal/handler/middleware"
	"shieldvn-backend/internal/observability"
	"shieldvn-backend/internal/service"
	"shieldvn-backend/internal/store"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	// Load configuration from environment.
	cfg := config.Load()

	// Step 1: Initialize structured logging (slog → JSON → stdout).
	observability.InitLogger(cfg.ParsedLogLevel())
	slog.Info("starting shieldvn-api",
		"port", cfg.Port,
		"log_level", cfg.LogLevel,
	)

	// Step 3: Initialize OpenTelemetry (GCP exporters on Cloud Run, no-op locally).
	ctx := context.Background()
	telemetryShutdown, err := observability.InitTelemetry(ctx, observability.TelemetryConfig{
		GCPProjectID: cfg.GCPProjectID,
		ServiceName:  "shieldvn-api",
	})
	if err != nil {
		slog.Error("failed to initialize telemetry", "error", err)
		os.Exit(1)
	}

	// Step 4: Initialize Gemini Service
	geminiSvc, err := service.NewGeminiService(ctx, cfg.GeminiAPIKey)
	if err != nil {
		slog.Error("failed to initialize gemini service", "error", err)
		os.Exit(1)
	}

	// Step 5: Initialize Firestore and Blacklist Service
	firestoreStore, err := store.NewFirestoreStore(ctx, cfg.GCPProjectID)
	if err != nil {
		slog.Warn("failed to initialize firestore store (ADC missing?), Tier-1 blacklist lookup will be disabled", "error", err)
		// We proceed without Firestore to allow local development of other features
		firestoreStore = nil
	} else {
		defer firestoreStore.Close()
	}

	blacklistSvc := service.NewBlacklistService(firestoreStore)

	// Build Gin router.
	router := gin.New()
	router.Use(gin.Recovery())

	// CORS — allow the configured frontend origin.
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.AllowedOrigin},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Middleware order: otelgin (creates root span) → requestid (stamps span).
	router.Use(otelgin.Middleware("shieldvn-api"))
	router.Use(middleware.RequestID())

	// Health check — no auth, no middleware overhead concern.
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes — Phase 01 fills in the real handler.
	v1 := router.Group("/api/v1")
	{
		v1.POST("/analyze", handler.AnalyzeHandler(geminiSvc, blacklistSvc))
	}

	// Start HTTP server.
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("shutting down", "signal", sig.String())

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server first, then flush telemetry.
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}
	if err := telemetryShutdown(shutdownCtx); err != nil {
		slog.Error("telemetry shutdown error", "error", err)
	}

	slog.Info("shieldvn-api stopped")
}
