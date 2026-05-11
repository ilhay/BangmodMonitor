package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bangmodmonitor/probe/checker"
	"github.com/bangmodmonitor/probe/config"
)

func main() {
	cfg := config.Load()

	if cfg.NodeSecret == "" {
		log.Fatal("NODE_SECRET is required")
	}

	// Phase 1: targets loaded from env (comma-separated URLs)
	// Phase 3+: pulled from API based on customer subscriptions
	targets := loadTargets()
	if len(targets) == 0 {
		log.Fatal("No probe targets. Set PROBE_TARGETS=https://example.com,https://example2.com")
	}

	log.Printf("Probe node started | region=%s targets=%d interval=%ds",
		cfg.Region, len(targets), cfg.Interval)

	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
	defer ticker.Stop()

	runChecks(cfg, targets)
	for range ticker.C {
		runChecks(cfg, targets)
	}
}

func loadTargets() []string {
	raw := os.Getenv("PROBE_TARGETS")
	if raw == "" {
		return nil
	}
	var targets []string
	for _, t := range splitCSV(raw) {
		if t != "" {
			targets = append(targets, t)
		}
	}
	return targets
}

func splitCSV(s string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

type probePayload struct {
	Region     string         `json:"region"`
	Results    []probeResult  `json:"results"`
}

type probeResult struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	ResponseMS int64  `json:"response_ms"`
	IsUp       bool   `json:"is_up"`
	Error      string `json:"error,omitempty"`
}

func runChecks(cfg *config.Config, targets []string) {
	var results []probeResult
	for _, url := range targets {
		r := checker.CheckHTTP(url)
		results = append(results, probeResult{
			URL:        r.URL,
			StatusCode: r.StatusCode,
			ResponseMS: r.ResponseMS,
			IsUp:       r.IsUp,
			Error:      r.Error,
		})
		status := "UP"
		if !r.IsUp {
			status = "DOWN"
		}
		log.Printf("[%s] %s %s %dms", cfg.Region, url, status, r.ResponseMS)
	}

	payload, _ := json.Marshal(probePayload{Region: cfg.Region, Results: results})
	req, _ := http.NewRequest(http.MethodPost, cfg.APIURL+"/api/v1/probe", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Node-Secret", cfg.NodeSecret)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("probe report error: %v", err)
		return
	}
	defer resp.Body.Close()
}
