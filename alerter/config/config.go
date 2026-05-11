package config

import (
	"os"
	"strconv"
)

type Config struct {
	MariaDSN      string
	ClickHouseDSN string
	EvalInterval  int // seconds between rule evaluations
}

func Load() *Config {
	interval := 60
	if v := os.Getenv("EVAL_INTERVAL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			interval = n
		}
	}
	return &Config{
		MariaDSN:      getEnv("MARIADB_DSN", "bangmod:bangmod@tcp(localhost:3306)/bangmod?parseTime=true&charset=utf8mb4"),
		ClickHouseDSN: getEnv("CLICKHOUSE_DSN", "clickhouse://localhost:9000/bangmod"),
		EvalInterval:  interval,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
