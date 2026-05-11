package config

import "os"

type Config struct {
	HTTPPort      string
	ClickHouseDSN string
	MariaDSN      string
}

func Load() *Config {
	return &Config{
		HTTPPort:      getEnv("HTTP_PORT", "8080"),
		ClickHouseDSN: getEnv("CLICKHOUSE_DSN", "clickhouse://localhost:9000/bangmod"),
		MariaDSN:      getEnv("MARIADB_DSN", "bangmod:bangmod@tcp(localhost:3306)/bangmod?parseTime=true&charset=utf8mb4"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
