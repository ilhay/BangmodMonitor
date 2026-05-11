package config

import (
	"os"
	"strconv"
)

type Config struct {
	APIURL   string
	NodeSecret string
	Region   string
	Interval int // seconds
}

func Load() *Config {
	interval := 60
	if v := os.Getenv("PROBE_INTERVAL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			interval = n
		}
	}
	return &Config{
		APIURL:     getEnv("API_URL", "http://localhost:8080"),
		NodeSecret: getEnv("NODE_SECRET", ""),
		Region:     getEnv("PROBE_REGION", "default"),
		Interval:   interval,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
