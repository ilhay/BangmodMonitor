package config

import "os"

type Config struct {
	HTTPPort      string
	ClickHouseDSN string
	MariaDSN      string
	NodeSecret    string
	JWTSecret     string

	// Stripe
	StripeSecretKey      string
	StripeWebhookSecret  string
	StripeStarterPriceID string
	StripeProPriceID     string
	StripeRegionPriceID  string
	AppBaseURL           string // e.g. https://app.bangmodmonitor.com
}

func Load() *Config {
	return &Config{
		HTTPPort:      getEnv("HTTP_PORT", "8080"),
		ClickHouseDSN: getEnv("CLICKHOUSE_DSN", "clickhouse://localhost:9000/bangmod"),
		MariaDSN:      getEnv("MARIADB_DSN", "bangmod:bangmod@tcp(localhost:3306)/bangmod?parseTime=true&charset=utf8mb4"),
		NodeSecret:    getEnv("NODE_SECRET", ""),
		JWTSecret:     getEnv("JWT_SECRET", "change-me-in-production"),

		StripeSecretKey:      getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret:  getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripeStarterPriceID: getEnv("STRIPE_STARTER_PRICE_ID", ""),
		StripeProPriceID:     getEnv("STRIPE_PRO_PRICE_ID", ""),
		StripeRegionPriceID:  getEnv("STRIPE_REGION_PRICE_ID", ""),
		AppBaseURL:           getEnv("APP_BASE_URL", "http://localhost:3000"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
