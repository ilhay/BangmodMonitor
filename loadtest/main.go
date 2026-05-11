// BangmodMonitor load tester — simulates N concurrent agents pushing via gRPC.
// Usage: GRPC_TARGET=localhost:9090 AGENT_TOKEN=xxx AGENTS=1000 go run .
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	bangmodv1 "github.com/bangmodmonitor/gen/bangmod/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	target := getEnv("GRPC_TARGET", "localhost:9090")
	token := getEnv("AGENT_TOKEN", "test-token-phase1")
	agents, _ := strconv.Atoi(getEnv("AGENTS", "100"))
	duration, _ := strconv.Atoi(getEnv("DURATION_SEC", "30"))
	interval, _ := strconv.Atoi(getEnv("INTERVAL_SEC", "5"))

	conn, err := grpc.NewClient(target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	client := bangmodv1.NewMetricsServiceClient(conn)

	fmt.Printf("BangmodMonitor Load Test\n")
	fmt.Printf("  Target  : %s\n", target)
	fmt.Printf("  Agents  : %d\n", agents)
	fmt.Printf("  Interval: %ds\n", interval)
	fmt.Printf("  Duration: %ds\n", duration)
	fmt.Println()

	var (
		sent   atomic.Int64
		errors atomic.Int64
		totalLatency atomic.Int64
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(duration)*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < agents; i++ {
		wg.Add(1)
		go func(agentNum int) {
			defer wg.Done()
			ticker := time.NewTicker(time.Duration(interval) * time.Second)
			defer ticker.Stop()

			// Stagger start to avoid thundering herd
			time.Sleep(time.Duration(rand.Intn(interval*1000)) * time.Millisecond)

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					start := time.Now()
					_, err := client.IngestMetrics(ctx, &bangmodv1.IngestRequest{
						Token: token,
						Metrics: &bangmodv1.AgentMetrics{
							Timestamp: time.Now().Unix(),
							Hostname:  fmt.Sprintf("loadtest-agent-%d", agentNum),
							Cpu: &bangmodv1.CpuMetric{
								UsagePercent: 20.0 + rand.Float64()*60,
								Cores:        4,
							},
							Memory: &bangmodv1.MemoryMetric{
								TotalBytes:   8 * 1024 * 1024 * 1024,
								UsedBytes:    uint64(4+rand.Intn(4)) * 1024 * 1024 * 1024,
								UsagePercent: 50.0 + rand.Float64()*40,
							},
						},
					})
					latency := time.Since(start).Milliseconds()
					if err != nil {
						errors.Add(1)
					} else {
						sent.Add(1)
						totalLatency.Add(latency)
					}
				}
			}
		}(i)
	}

	// Progress reporter
	go func() {
		tick := time.NewTicker(5 * time.Second)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				s := sent.Load()
				e := errors.Load()
				var avgMs int64
				if s > 0 {
					avgMs = totalLatency.Load() / s
				}
				fmt.Printf("  sent=%d errors=%d avg_latency=%dms\n", s, e, avgMs)
			}
		}
	}()

	wg.Wait()

	s := sent.Load()
	e := errors.Load()
	var avgMs int64
	if s > 0 {
		avgMs = totalLatency.Load() / s
	}

	fmt.Printf("\n=== Results ===\n")
	fmt.Printf("  Sent        : %d\n", s)
	fmt.Printf("  Errors      : %d\n", e)
	fmt.Printf("  Throughput  : %.1f msg/s\n", float64(s)/float64(duration))
	fmt.Printf("  Avg latency : %dms\n", avgMs)
	fmt.Printf("  Error rate  : %.2f%%\n", float64(e)/float64(s+e)*100)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
