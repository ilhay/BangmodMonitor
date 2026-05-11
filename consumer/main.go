package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/bangmodmonitor/consumer/config"
	"github.com/bangmodmonitor/consumer/worker"
)

func main() {
	cfg := config.Load()

	opts, err := clickhouse.ParseDSN(cfg.ClickHouseDSN)
	if err != nil {
		log.Fatalf("clickhouse parse dsn: %v", err)
	}
	ch, err := clickhouse.Open(opts)
	if err != nil {
		log.Fatalf("clickhouse connect: %v", err)
	}
	if err := ch.Ping(context.Background()); err != nil {
		log.Fatalf("clickhouse ping: %v", err)
	}
	defer ch.Close()

	log.Printf("Consumer started | brokers=%v group=%s", cfg.RedpandaBrokers, cfg.GroupID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	agentWorker := worker.NewAgentWorker(cfg.RedpandaBrokers, cfg.GroupID, ch)
	probeWorker := worker.NewProbeWorker(cfg.RedpandaBrokers, cfg.GroupID, ch)

	go agentWorker.Run(ctx)
	go probeWorker.Run(ctx)

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	fmt.Printf("\nReceived signal %v — shutting down...\n", sig)
	cancel()
	log.Println("Consumer stopped")
}
