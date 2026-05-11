package config

import "os"

type Config struct {
	HTTPPort          string
	ClickHouseDSN     string
	PostgresDSN       string
}

func Load() *Config {
	return &Config{
		HTTPPort:      getEnv("HTTP_PORT", "8080"),
		ClickHouseDSN: getEnv("CLICKHOUSE_DSN", "clickhouse://localhost:9000/bangmod"),
		PostgresDSN:   getEnv("POSTGRES_DSN", "postgres://bangmod:bangmod@localhost:5432/bangmod?sslmode=disable"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
