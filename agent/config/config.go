package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

type Config struct {
	Token      string
	APIURL     string // HTTP fallback (Phase 1 compat)
	GRPCTarget string // gRPC target e.g. localhost:9090
	Interval   int
	Region     string
	WALDir     string
	UseGRPC    bool
}

func Load() *Config {
	interval := 30
	if v := os.Getenv("AGENT_INTERVAL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			interval = n
		}
	}
	grpcTarget := os.Getenv("GRPC_TARGET")
	return &Config{
		Token:      getEnv("AGENT_TOKEN", ""),
		APIURL:     getEnv("API_URL", "http://localhost:8080"),
		GRPCTarget: grpcTarget,
		Interval:   interval,
		Region:     getEnv("AGENT_REGION", "default"),
		WALDir:     getEnv("WAL_DIR", defaultWALDir()),
		UseGRPC:    grpcTarget != "",
	}
}

func defaultWALDir() string {
	if runtime.GOOS == "windows" {
		base := os.Getenv("APPDATA")
		if base == "" {
			base = os.TempDir()
		}
		return filepath.Join(base, "BangmodMonitor")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".bangmodmonitor")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
