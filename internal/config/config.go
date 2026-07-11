package config

import (
	"os"
	"time"
)

type Config struct {
	HTTPAddr        string
	DatabaseURL     string
	ExternalBaseURL string
	ShutdownTimeout time.Duration
}

func Load() Config {
	return Config{
		HTTPAddr:        getenv("HTTP_ADDR", ":8081"),
		DatabaseURL:     getenv("DATABASE_URL", "postgres://app:app@localhost:55433/app?sslmode=disable"),
		ExternalBaseURL: getenv("EXTERNAL_BASE_URL", "http://localhost:8081"),
		ShutdownTimeout: 5 * time.Second,
	}
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
