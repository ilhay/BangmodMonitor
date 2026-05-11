package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	kafka "github.com/segmentio/kafka-go"
)

// AgentMetricMsg mirrors api/mq.AgentMetricMsg — duplicated to avoid
// cross-module import. Keep in sync with the API's mq package.
type AgentMetricMsg struct {
	Timestamp  time.Time `json:"timestamp"`
	HostID     string    `json:"host_id"`
	Hostname   string    `json:"hostname"`
	Region     string    `json:"region"`
	CPUPercent float32   `json:"cpu_percent"`
	CPUCores   uint8     `json:"cpu_cores"`
	MemTotal   uint64    `json:"mem_total"`
	MemUsed    uint64    `json:"mem_used"`
	MemPercent float32   `json:"mem_percent"`
	Disks      []DiskMsg `json:"disks,omitempty"`
	Networks   []NetMsg  `json:"networks,omitempty"`
}

type DiskMsg struct {
	Path       string  `json:"path"`
	TotalBytes uint64  `json:"total_bytes"`
	UsedBytes  uint64  `json:"used_bytes"`
	UsePct     float32 `json:"use_pct"`
}

type NetMsg struct {
	Interface   string `json:"interface"`
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
}

// AgentWorker consumes from bangmod.agent-metrics and batch-writes to ClickHouse.
type AgentWorker struct {
	reader *kafka.Reader
	ch     clickhouse.Conn
}

func NewAgentWorker(brokers []string, groupID string, ch clickhouse.Conn) *AgentWorker {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          "bangmod.agent-metrics",
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10_000_000,
		MaxWait:        100 * time.Millisecond,
		CommitInterval: time.Second, // auto-commit every second
		StartOffset:    kafka.LastOffset,
	})
	return &AgentWorker{reader: r, ch: ch}
}

func (w *AgentWorker) Run(ctx context.Context) {
	log.Println("AgentWorker: consuming from bangmod.agent-metrics")
	batch := make([]AgentMetricMsg, 0, 200)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				w.flush(ctx, batch)
			}
			w.reader.Close()
			return
		case <-ticker.C:
			if len(batch) > 0 {
				w.flush(ctx, batch)
				batch = batch[:0]
			}
		default:
			msg, err := w.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				continue
			}

			var m AgentMetricMsg
			if err := json.Unmarshal(msg.Value, &m); err != nil {
				log.Printf("AgentWorker: unmarshal error: %v", err)
				_ = w.reader.CommitMessages(ctx, msg)
				continue
			}
			batch = append(batch, m)

			if len(batch) >= 200 {
				w.flush(ctx, batch)
				batch = batch[:0]
			}
			_ = w.reader.CommitMessages(ctx, msg)
		}
	}
}

func (w *AgentWorker) flush(ctx context.Context, batch []AgentMetricMsg) {
	if len(batch) == 0 {
		return
	}

	agentBatch, err := w.ch.PrepareBatch(ctx, `INSERT INTO bangmod.agent_metrics`)
	if err != nil {
		log.Printf("AgentWorker: prepare batch: %v", err)
		return
	}
	diskBatch, _ := w.ch.PrepareBatch(ctx, `INSERT INTO bangmod.disk_metrics`)
	netBatch, _ := w.ch.PrepareBatch(ctx, `INSERT INTO bangmod.net_metrics`)

	for _, m := range batch {
		_ = agentBatch.Append(m.Timestamp, m.HostID, m.Hostname, m.Region,
			m.CPUPercent, m.CPUCores, m.MemTotal, m.MemUsed, m.MemPercent)
		for _, d := range m.Disks {
			_ = diskBatch.Append(m.Timestamp, m.HostID, d.Path, d.TotalBytes, d.UsedBytes, d.UsePct)
		}
		for _, n := range m.Networks {
			_ = netBatch.Append(m.Timestamp, m.HostID, n.Interface,
				n.BytesSent, n.BytesRecv, n.PacketsSent, n.PacketsRecv)
		}
	}

	if err := agentBatch.Send(); err != nil {
		log.Printf("AgentWorker: agent batch send: %v", err)
	} else {
		log.Printf("AgentWorker: wrote %d metrics to ClickHouse", len(batch))
	}
	_ = diskBatch.Send()
	_ = netBatch.Send()
}
