package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bangmodmonitor/probe/checker"
	"github.com/bangmodmonitor/probe/config"
)

func main() {
	cfg := config.Load()

	if cfg.NodeSecret == "" {
		log.Fatal("NODE_SECRET is required")
	}

	httpTargets := loadEnvList("PROBE_TARGETS")
	tcpTargets := loadTCPTargets("PROBE_TCP_TARGETS") // format: host:port,host:port

	if len(httpTargets) == 0 && len(tcpTargets) == 0 {
		log.Fatal("No targets. Set PROBE_TARGETS and/or PROBE_TCP_TARGETS")
	}

	log.Printf("Probe node started | region=%s http=%d tcp=%d interval=%ds",
		cfg.Region, len(httpTargets), len(tcpTargets), cfg.Interval)

	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
	defer ticker.Stop()

	runChecks(cfg, httpTargets, tcpTargets)
	for range ticker.C {
		runChecks(cfg, httpTargets, tcpTargets)
	}
}

type tcpTarget struct {
	host string
	port int
}

func loadEnvList(key string) []string {
	raw := os.Getenv(key)
	if raw == "" {
		return nil
	}
	var out []string
	for _, s := range strings.Split(raw, ",") {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func loadTCPTargets(key string) []tcpTarget {
	var out []tcpTarget
	for _, s := range loadEnvList(key) {
		parts := strings.SplitN(s, ":", 2)
		if len(parts) != 2 {
			continue
		}
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		out = append(out, tcpTarget{host: parts[0], port: port})
	}
	return out
}

type probePayload struct {
	Region  string        `json:"region"`
	Results []probeResult `json:"results"`
}

type probeResult struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	ResponseMS int64  `json:"response_ms"`
	IsUp       bool   `json:"is_up"`
	Error      string `json:"error,omitempty"`
}

func runChecks(cfg *config.Config, httpTargets []string, tcpTargets []tcpTarget) {
	var results []probeResult

	for _, url := range httpTargets {
		r := checker.CheckHTTP(url)
		status := "UP"
		if !r.IsUp {
			status = "DOWN"
		}
		log.Printf("[%s] HTTP %s %s %dms", cfg.Region, url, status, r.ResponseMS)
		results = append(results, probeResult{
			URL: url, StatusCode: r.StatusCode,
			ResponseMS: r.ResponseMS, IsUp: r.IsUp, Error: r.Error,
		})
	}

	for _, t := range tcpTargets {
		r := checker.CheckTCP(t.host, t.port)
		url := fmt.Sprintf("tcp://%s:%d", t.host, t.port)
		status := "UP"
		if !r.IsUp {
			status = "DOWN"
		}
		log.Printf("[%s] TCP %s %s %dms", cfg.Region, url, status, r.ResponseMS)
		results = append(results, probeResult{
			URL: url, ResponseMS: r.ResponseMS, IsUp: r.IsUp, Error: r.Error,
		})
	}

	if len(results) == 0 {
		return
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
