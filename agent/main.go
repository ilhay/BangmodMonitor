package main

import (
	"log"
	"os"
	"time"

	"github.com/bangmodmonitor/agent/collector"
	"github.com/bangmodmonitor/agent/config"
	"github.com/bangmodmonitor/agent/sender"
)

func main() {
	cfg := config.Load()

	if cfg.Token == "" {
		log.Fatal("AGENT_TOKEN is required. Set it via environment variable or --token flag.")
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	s := sender.New(cfg.APIURL, cfg.Token)
	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
	defer ticker.Stop()

	log.Printf("BangmodMonitor Agent started | host=%s region=%s interval=%ds api=%s",
		hostname, cfg.Region, cfg.Interval, cfg.APIURL)

	// Collect and send immediately on start
	collect(s, hostname)

	for range ticker.C {
		collect(s, hostname)
	}
}

func collect(s *sender.Sender, hostname string) {
	metrics, err := collector.Collect(hostname)
	if err != nil {
		log.Printf("collect error: %v", err)
		return
	}

	if err := s.Send(metrics); err != nil {
		log.Printf("send error: %v", err)
		return
	}

	log.Printf("metrics sent | host=%s cpu=%.1f%% mem=%.1f%%",
		hostname, metrics.CPU.UsagePercent, metrics.Memory.UsagePercent)
}
