package main

import (
	"log"
	"os"
	"time"

	bangmodv1 "github.com/bangmodmonitor/gen/bangmod/v1"
	"github.com/bangmodmonitor/agent/collector"
	"github.com/bangmodmonitor/agent/config"
	"github.com/bangmodmonitor/agent/sender"
	"github.com/bangmodmonitor/agent/wal"
)

func main() {
	cfg := config.Load()

	if cfg.Token == "" {
		log.Fatal("AGENT_TOKEN is required")
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	w, walErr := wal.New(cfg.WALDir)
	if walErr != nil {
		log.Printf("WARNING: WAL init failed (%v) — no offline buffering", walErr)
	}

	var grpcSender *sender.GRPCSender
	if cfg.UseGRPC {
		grpcSender, err = sender.NewGRPC(cfg.GRPCTarget, cfg.Token)
		if err != nil {
			log.Fatalf("gRPC sender: %v", err)
		}
		defer grpcSender.Close()
		log.Printf("Protocol: gRPC → %s", cfg.GRPCTarget)
	} else {
		log.Printf("Protocol: HTTP → %s", cfg.APIURL)
	}

	httpSender := sender.New(cfg.APIURL, cfg.Token)
	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
	defer ticker.Stop()

	log.Printf("BangmodMonitor Agent started | host=%s region=%s interval=%ds wal=%s",
		hostname, cfg.Region, cfg.Interval, cfg.WALDir)

	collect(cfg, hostname, grpcSender, httpSender, w)
	for range ticker.C {
		collect(cfg, hostname, grpcSender, httpSender, w)
	}
}

func collect(cfg *config.Config, hostname string,
	grpcSender *sender.GRPCSender, httpSender *sender.Sender, w *wal.WAL) {

	m, err := collector.Collect(hostname)
	if err != nil {
		log.Printf("collect error: %v", err)
		return
	}

	if cfg.UseGRPC && grpcSender != nil {
		if w != nil {
			replayGRPCWAL(w, grpcSender)
		}
		if err := grpcSender.Send(toProto(m)); err != nil {
			log.Printf("gRPC send error: %v — buffering to WAL", err)
			if w != nil {
				_ = w.Append(m)
			}
			return
		}
	} else {
		if w != nil {
			replayHTTPWAL(w, httpSender)
		}
		if err := httpSender.Send(m); err != nil {
			log.Printf("HTTP send error: %v — buffering to WAL", err)
			if w != nil {
				_ = w.Append(m)
			}
			return
		}
	}

	log.Printf("metrics sent | host=%s cpu=%.1f%% mem=%.1f%%",
		hostname, m.CPU.UsagePercent, m.Memory.UsagePercent)
}

func replayGRPCWAL(w *wal.WAL, s *sender.GRPCSender) {
	entries, err := wal.Drain[collector.Metrics](w)
	if err != nil || len(entries) == 0 {
		return
	}
	log.Printf("WAL replay: %d buffered entries", len(entries))
	for _, m := range entries {
		if err := s.Send(toProto(&m)); err != nil {
			log.Printf("WAL replay failed: %v — re-buffering", err)
			_ = w.Append(m)
			return
		}
	}
	log.Printf("WAL replay complete")
}

func replayHTTPWAL(w *wal.WAL, s *sender.Sender) {
	entries, err := wal.Drain[collector.Metrics](w)
	if err != nil || len(entries) == 0 {
		return
	}
	log.Printf("WAL replay: %d buffered entries", len(entries))
	for _, m := range entries {
		if err := s.Send(&m); err != nil {
			log.Printf("WAL replay failed: %v — re-buffering", err)
			_ = w.Append(m)
			return
		}
	}
}

func toProto(m *collector.Metrics) *bangmodv1.AgentMetrics {
	p := &bangmodv1.AgentMetrics{
		Timestamp: m.Timestamp,
		Hostname:  m.Hostname,
		Cpu: &bangmodv1.CpuMetric{
			UsagePercent: m.CPU.UsagePercent,
			Cores:        int32(m.CPU.Cores),
		},
		Memory: &bangmodv1.MemoryMetric{
			TotalBytes:   m.Memory.TotalBytes,
			UsedBytes:    m.Memory.UsedBytes,
			UsagePercent: m.Memory.UsagePercent,
		},
		Apps: &bangmodv1.AppStatus{
			Nginx: m.Apps.Nginx, Apache: m.Apps.Apache,
			Mysql: m.Apps.MySQL, Mariadb: m.Apps.MariaDB,
			Postgresql: m.Apps.PostgreSQL,
		},
	}
	for _, d := range m.Disks {
		p.Disks = append(p.Disks, &bangmodv1.DiskMetric{
			Path: d.Path, TotalBytes: d.TotalBytes,
			UsedBytes: d.UsedBytes, UsagePercent: d.UsagePercent,
		})
	}
	for _, n := range m.Networks {
		p.Networks = append(p.Networks, &bangmodv1.NetworkMetric{
			Interface:   n.Interface,
			BytesSent:   n.BytesSent, BytesRecv: n.BytesRecv,
			PacketsSent: n.PacketsSent, PacketsRecv: n.PacketsRecv,
		})
	}
	return p
}
