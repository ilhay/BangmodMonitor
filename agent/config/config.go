package config

import (
	"os"
	"strconv"
)

type Config struct {
	Token    string
	APIURL   string
	Interval int // seconds
	Region   string
}

func Load() *Config {
	interval := 30
	if v := os.Getenv("AGENT_INTERVAL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			interval = n
		}
	}
	return &Config{
		Token:    getEnv("AGENT_TOKEN", ""),
		APIURL:   getEnv("API_URL", "http://localhost:8080"),
		Interval: interval,
		Region:   getEnv("AGENT_REGION", "default"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
