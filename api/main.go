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

	probeHandler := handler.NewProbe(ch, cfg.NodeSecret)

	v1 := r.Group("/api/v1")
	{
		// Agent ingest — requires valid agent token
		v1.POST("/ingest", middleware.Auth(maria), handler.NewIngest(ch).Handle)

		// Probe node ingest — validated by NODE_SECRET header
		v1.POST("/probe", probeHandler.Ingest)

		// Dashboard queries — open in Phase 1, auth added in Phase 3
		v1.GET("/metrics/:hostId", handler.NewMetrics(ch).Recent)
		v1.GET("/probe/results", probeHandler.Recent)
	}

	log.Printf("API server listening on :%s", cfg.HTTPPort)
	if err := r.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatalf("server: %v", err)
	}
}
