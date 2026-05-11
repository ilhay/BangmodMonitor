package config

import (
	"os"
	"strings"
)

type Config struct {
	RedpandaBrokers []string
	ClickHouseDSN   string
	GroupID         string
	BatchSize       int
}

func Load() *Config {
	batchSize := 500
	return &Config{
		RedpandaBrokers: splitCSV(getEnv("REDPANDA_BROKERS", "localhost:19092")),
		ClickHouseDSN:   getEnv("CLICKHOUSE_DSN", "clickhouse://localhost:9000/bangmod"),
		GroupID:         getEnv("CONSUMER_GROUP", "bangmod-clickhouse-writer"),
		BatchSize:       batchSize,
	}
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
