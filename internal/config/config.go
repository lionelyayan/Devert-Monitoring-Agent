package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Server identity
	ServerName string
	Timezone   string
	Location   *time.Location

	// HTTP API
	HTTPPort      string
	APIToken      string
	RateLimitRPM  int

	// PostgreSQL
	PostgresDSN string

	// Docker
	DockerSocket string

	// n8n Webhook
	N8NWebhookURL     string
	N8NWebhookSecret  string
	N8NWebhookEnabled bool

	// Logging
	LogLevel string

	// Monitoring intervals
	ServerPollInterval   time.Duration
	ResourcePollInterval time.Duration
	ServicePollInterval  time.Duration
}

// Load reads configuration from a .env file (if present) and environment variables.
func Load() (*Config, error) {
	// Load .env file if it exists — ignore error if not found (env vars may be set directly)
	_ = godotenv.Load()

	loc, err := loadLocation()
	if err != nil {
		return nil, fmt.Errorf("config: invalid TIMEZONE: %w", err)
	}

	cfg := &Config{
		ServerName: getEnv("SERVER_NAME", "devert-server"),
		Timezone:   getEnv("TIMEZONE", "UTC"),
		Location:   loc,

		HTTPPort:     getEnv("HTTP_PORT", "8080"),
		APIToken:     getEnv("API_TOKEN", ""),
		RateLimitRPM: getEnvInt("RATE_LIMIT_RPM", 100),

		PostgresDSN: getEnv("POSTGRES_DSN", ""),

		DockerSocket: getEnv("DOCKER_SOCKET", "/var/run/docker.sock"),

		N8NWebhookURL:     getEnv("N8N_WEBHOOK_URL", ""),
		N8NWebhookSecret:  getEnv("N8N_WEBHOOK_SECRET", ""),
		N8NWebhookEnabled: getEnvBool("N8N_WEBHOOK_ENABLED", true),

		LogLevel: getEnv("LOG_LEVEL", "info"),

		ServerPollInterval:   getEnvDuration("SERVER_POLL_INTERVAL", 30) * time.Second,
		ResourcePollInterval: getEnvDuration("RESOURCE_POLL_INTERVAL", 10) * time.Second,
		ServicePollInterval:  getEnvDuration("SERVICE_POLL_INTERVAL", 60) * time.Second,
	}

	if cfg.APIToken == "" {
		return nil, fmt.Errorf("config: API_TOKEN must be set")
	}
	if cfg.PostgresDSN == "" {
		return nil, fmt.Errorf("config: POSTGRES_DSN must be set")
	}

	return cfg, nil
}

func loadLocation() (*time.Location, error) {
	tz := getEnv("TIMEZONE", "UTC")
	return time.LoadLocation(tz)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return fallback
}

func getEnvDuration(key string, fallbackSeconds int) time.Duration {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return time.Duration(i)
		}
	}
	return time.Duration(fallbackSeconds)
}
