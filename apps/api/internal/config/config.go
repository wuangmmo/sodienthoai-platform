package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port          string
	DatabaseURL   string
	RedisAddr     string
	OpenSearchURL string
	AppEnv        string
	AdminAPIToken string
	ReporterHashSecret string
}

func Load() (Config, error) {
	cfg := Config{
		Port:          envOrDefault("API_PORT", "8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		OpenSearchURL: os.Getenv("OPENSEARCH_URL"),
		AppEnv:        envOrDefault("APP_ENV", "development"),
		AdminAPIToken: os.Getenv("ADMIN_API_TOKEN"),
		ReporterHashSecret: os.Getenv("REPORTER_HASH_SECRET"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.RedisAddr == "" {
		return Config{}, fmt.Errorf("REDIS_ADDR is required")
	}
	if cfg.OpenSearchURL == "" {
		return Config{}, fmt.Errorf("OPENSEARCH_URL is required")
	}
	if cfg.AppEnv == "production" && len(cfg.AdminAPIToken) < 32 {
		return Config{}, fmt.Errorf("ADMIN_API_TOKEN must be at least 32 characters in production")
	}
	if cfg.AppEnv == "production" && len(cfg.ReporterHashSecret) < 32 {
		return Config{}, fmt.Errorf("REPORTER_HASH_SECRET must be at least 32 characters in production")
	}
	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
