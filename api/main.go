package main

import (
	"context"
	"log"
	"net/http"

	"github.com/bangmodmonitor/api/config"
	"github.com/bangmodmonitor/api/handler"
	"github.com/bangmodmonitor/api/middleware"
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

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	authHandler  := handler.NewAuth(maria, cfg.JWTSecret)
	probeHandler := handler.NewProbe(ch, cfg.NodeSecret)
	hostHandler  := handler.NewHost(maria, ch)

	v1 := r.Group("/api/v1")
	{
		// Public auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Agent ingest — Bearer agent token
		v1.POST("/ingest", middleware.Auth(maria), handler.NewIngest(ch).Handle)

		// Probe node ingest — NODE_SECRET header
		v1.POST("/probe", probeHandler.Ingest)
		v1.GET("/probe/results", probeHandler.Recent)

		// Customer routes — JWT required
		customer := v1.Group("/")
		customer.Use(middleware.RequireAuth(cfg.JWTSecret))
		{
			customer.GET("/me", authHandler.Me)

			// Host management
			hosts := customer.Group("/hosts")
			{
				hosts.GET("", hostHandler.List)
				hosts.POST("", hostHandler.Create)
				hosts.DELETE("/:id", hostHandler.Delete)
				hosts.POST("/:id/rotate", hostHandler.RotateToken)
				hosts.GET("/:id/metrics", hostHandler.Metrics)
			}
		}
	}

	log.Printf("API server listening on :%s", cfg.HTTPPort)
	if err := r.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatalf("server: %v", err)
	}
}
