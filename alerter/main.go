package main

import (
	"context"
	"log"
	"time"

	"github.com/bangmodmonitor/alerter/config"
	"github.com/bangmodmonitor/alerter/evaluator"
	"github.com/bangmodmonitor/alerter/storage"
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

	eval := evaluator.New(maria, ch)
	ticker := time.NewTicker(time.Duration(cfg.EvalInterval) * time.Second)
	defer ticker.Stop()

	log.Printf("Alert engine started | eval_interval=%ds", cfg.EvalInterval)

	// Run immediately on start
	eval.Run(ctx)

	for range ticker.C {
		eval.Run(ctx)
	}
}
