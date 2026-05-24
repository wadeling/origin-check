package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL   string
	RedisURL      string
	EncryptionKey string
	APIAddr       string
	CORSOrigin    string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:   envOr("DATABASE_URL", "postgres://origin:origin@localhost:5432/origin_check?sslmode=disable"),
		RedisURL:      envOr("REDIS_URL", "redis://localhost:6379/0"),
		EncryptionKey: envOr("ENCRYPTION_KEY", "change-me-to-32-byte-secret-key!!"),
		APIAddr:       envOr("API_ADDR", ":8080"),
		CORSOrigin:    envOr("CORS_ORIGIN", "http://localhost:3000"),
	}

	if len(cfg.EncryptionKey) < 32 {
		return nil, fmt.Errorf("ENCRYPTION_KEY must be at least 32 bytes")
	}

	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func Env(key string) string {
	return os.Getenv(key)
}
