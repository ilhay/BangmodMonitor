package main

import (
	"context"
	"log"
	"net/http"

	"github.com/bangmodmonitor/api/billing"
	"github.com/bangmodmonitor/api/cache"
	"github.com/bangmodmonitor/api/config"
	"github.com/bangmodmonitor/api/grpcserver"
	"github.com/bangmodmonitor/api/handler"
	"github.com/bangmodmonitor/api/middleware"
	"github.com/bangmodmonitor/api/mq"
	"github.com/bangmodmonitor/api/storage"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	maria, err := storage.NewMaria(ctx, cfg.MariaDSN)
	if err != nil {
		log.Fatalf("mariadb: %v", err)
	}
	defer maria.Close()

	if err := maria.Migrate(ctx); err != nil {
		log.Fatalf("mariadb migrate: %v", err)
	}

	ch, err := storage.NewCH(ctx, cfg.ClickHouseDSN)
	if err != nil {
		log.Fatalf("clickhouse: %v", err)
	}
	defer ch.Close()

	if err := ch.Migrate(ctx); err != nil {
		log.Fatalf("clickhouse migrate: %v", err)
	}

	// Redis token cache (optional — gracefully disabled if REDIS_ADDR is empty)
	tokenCache := cache.NewTokenCache(cfg.RedisAddr)
	if tokenCache.Enabled() {
		if err := tokenCache.Ping(ctx); err != nil {
			log.Printf("WARNING: Redis ping failed (%v) — falling back to DB-only token validation", err)
		} else {
			log.Printf("Redis token cache: connected to %s", cfg.RedisAddr)
		}
	} else {
		log.Println("Redis token cache: disabled (set REDIS_ADDR to enable)")
	}

	// Stripe
	stripeSvc := billing.NewStripe(cfg.StripeSecretKey, cfg.StripeWebhookSecret, cfg.StripeRegionPriceID)
	if stripeSvc.Enabled() {
		log.Println("Stripe integration: enabled")
	} else {
		log.Println("Stripe integration: disabled (set STRIPE_SECRET_KEY to enable)")
	}

	// Redpanda producer (optional — disabled when REDPANDA_BROKERS is empty)
	producer, err := mq.NewProducer(cfg.RedpandaBrokers)
	if err != nil {
		log.Fatalf("redpanda producer: %v", err)
	}
	defer producer.Close()
	if producer.Enabled() {
		log.Printf("Redpanda producer: connected to %v", cfg.RedpandaBrokers)
	} else {
		log.Println("Redpanda producer: disabled (set REDPANDA_BROKERS to enable)")
	}

	// gRPC server (runs alongside HTTP on a separate port)
	if cfg.GRPCPort != "" {
		grpcSrv := grpcserver.New(ch, maria, tokenCache, producer, cfg.NodeSecret)
		grpcserver.Start(grpcSrv, ":"+cfg.GRPCPort)
	}

	authHandler    := handler.NewAuth(maria, cfg.JWTSecret, stripeSvc)
	probeHandler   := handler.NewProbe(ch, cfg.NodeSecret, producer)
	hostHandler    := handler.NewHost(maria, ch, tokenCache)
	billingHandler := handler.NewBilling(maria, stripeSvc, cfg.AppBaseURL, cfg.StripeStarterPriceID, cfg.StripeProPriceID)
	adminHandler   := handler.NewAdmin(maria)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":       "ok",
			"redis_cache":  tokenCache.Enabled(),
			"grpc_port":    cfg.GRPCPort,
		})
	})

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		v1.POST("/billing/webhook", billingHandler.Webhook)

		// Agent ingest — Bearer agent token (HTTP fallback for Phase 5 migration)
		v1.POST("/ingest", middleware.Auth(maria, tokenCache), handler.NewIngest(ch, producer).Handle)

		// Probe node ingest — NODE_SECRET header
		v1.POST("/probe", probeHandler.Ingest)
		v1.GET("/probe/results", probeHandler.Recent)

		customer := v1.Group("/")
		customer.Use(middleware.RequireAuth(cfg.JWTSecret))
		{
			customer.GET("/me", authHandler.Me)

			hosts := customer.Group("/hosts")
			{
				hosts.GET("", hostHandler.List)
				hosts.POST("", hostHandler.Create)
				hosts.DELETE("/:id", hostHandler.Delete)
				hosts.POST("/:id/rotate", hostHandler.RotateToken)
				hosts.GET("/:id/metrics", hostHandler.Metrics)
			}

			bill := customer.Group("/billing")
			{
				bill.GET("/overview", billingHandler.Overview)
				bill.POST("/checkout", billingHandler.Checkout)
				bill.POST("/portal", billingHandler.Portal)
				bill.GET("/regions", billingHandler.GetRegions)
				bill.POST("/regions/:region", billingHandler.AddRegion)
				bill.DELETE("/regions/:region", billingHandler.RemoveRegion)
				bill.GET("/invoices", billingHandler.Invoices)
			}

			admin := customer.Group("/admin")
			{
				admin.GET("/orgs", adminHandler.ListOrgs)
				admin.GET("/stats", adminHandler.Stats)
			}
		}
	}

	log.Printf("HTTP server listening on :%s", cfg.HTTPPort)
	if err := r.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatalf("server: %v", err)
	}
}
