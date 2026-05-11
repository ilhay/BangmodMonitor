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

	pg, err := storage.NewPG(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pg.Close()

	if err := pg.Migrate(ctx); err != nil {
		log.Fatalf("postgres migrate: %v", err)
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

	v1 := r.Group("/api/v1")
	{
		// Agent ingest — requires valid agent token
		v1.POST("/ingest", middleware.Auth(pg), handler.NewIngest(ch).Handle)

		// Dashboard queries — open in Phase 1, will add auth in Phase 3
		v1.GET("/metrics/:hostId", handler.NewMetrics(ch).Recent)
	}

	log.Printf("API server listening on :%s", cfg.HTTPPort)
	if err := r.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatalf("server: %v", err)
	}
}
