package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"

	"github.com/devert/monitor-agent/internal/api"
	"github.com/devert/monitor-agent/internal/config"
	"github.com/devert/monitor-agent/internal/database"
	"github.com/devert/monitor-agent/internal/docker"
	applogger "github.com/devert/monitor-agent/internal/logger"
	"github.com/devert/monitor-agent/internal/webhook"
)

// Version info — injected at build time via -ldflags from GitHub Actions.
var (
	version   = "dev"
	buildDate = "unknown"
	gitCommit = "unknown"
)


func main() {
	// ── Config ───────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: %v\n", err)
		os.Exit(1)
	}

	// ── Logger ───────────────────────────────────────────────────
	applogger.Init(cfg.LogLevel)

	log.Info().
		Str("server", cfg.ServerName).
		Str("version", version).
		Str("build_date", buildDate).
		Str("git_commit", gitCommit[:min(len(gitCommit), 7)]).
		Msg("🚀 Devert Monitor Agent starting")

	// ── Root context with graceful shutdown ──────────────────────
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// ── Database ─────────────────────────────────────────────────
	db, err := database.Connect(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("database: failed to connect")
	}
	defer db.Close()

	// ── Docker client ─────────────────────────────────────────────
	cli, err := client.NewClientWithOpts(
		client.WithHost("unix://"+cfg.DockerSocket),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("docker: failed to create client")
	}
	defer cli.Close()

	// Verify Docker connectivity
	if _, err := cli.Ping(ctx); err != nil {
		log.Fatal().Err(err).Str("socket", cfg.DockerSocket).Msg("docker: ping failed — is Docker running?")
	}
	log.Info().Msg("docker: connected successfully")

	// ── Webhook sender ────────────────────────────────────────────
	wh := webhook.New(cfg.N8NWebhookURL, cfg.N8NWebhookSecret, cfg.N8NWebhookEnabled)

	// ── Docker event listener (background goroutine) ──────────────
	eventListener := docker.NewEventListener(cli, cfg.ServerName, db, wh, cfg.Location)
	go eventListener.Listen(ctx)

	// ── HTTP server ───────────────────────────────────────────────
	router := api.NewRouter(cfg.APIToken, cfg.RateLimitRPM, cfg.ServerName, cli, db)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("addr", ":"+cfg.HTTPPort).Msg("http: server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("http: server error")
		}
	}()

	log.Info().Msg("✅ Devert Monitor Agent ready")

	// ── Wait for shutdown signal ──────────────────────────────────
	<-quit
	log.Info().Msg("shutdown: signal received, draining connections...")

	cancel() // Stop background goroutines

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("shutdown: HTTP server forced shutdown")
	}

	log.Info().Msg("shutdown: Devert Monitor Agent stopped gracefully")
}
