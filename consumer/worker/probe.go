package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	kafka "github.com/segmentio/kafka-go"
)

// ProbeResultMsg mirrors api/mq.ProbeResultMsg.
type ProbeResultMsg struct {
	Timestamp  time.Time `json:"timestamp"`
	HostID     string    `json:"host_id"`
	TargetURL  string    `json:"target_url"`
	Region     string    `json:"region"`
	StatusCode uint16    `json:"status_code"`
	ResponseMS uint32    `json:"response_ms"`
	IsUp       uint8     `json:"is_up"`
}

// ProbeWorker consumes from bangmod.probe-results and batch-writes to ClickHouse.
type ProbeWorker struct {
	reader *kafka.Reader
	ch     clickhouse.Conn
}

func NewProbeWorker(brokers []string, groupID string, ch clickhouse.Conn) *ProbeWorker {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          "bangmod.probe-results",
		GroupID:        groupID + "-probe",
		MinBytes:       1,
		MaxBytes:       10_000_000,
		MaxWait:        200 * time.Millisecond,
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})
	return &ProbeWorker{reader: r, ch: ch}
}

func (w *ProbeWorker) Run(ctx context.Context) {
	log.Println("ProbeWorker: consuming from bangmod.probe-results")
	batch := make([]ProbeResultMsg, 0, 100)
	ticker := time.NewTicker(time.Second)
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

			var m ProbeResultMsg
			if err := json.Unmarshal(msg.Value, &m); err != nil {
				log.Printf("ProbeWorker: unmarshal error: %v", err)
				_ = w.reader.CommitMessages(ctx, msg)
				continue
			}
			batch = append(batch, m)
			if len(batch) >= 100 {
				w.flush(ctx, batch)
				batch = batch[:0]
			}
			_ = w.reader.CommitMessages(ctx, msg)
		}
	}
}

func (w *ProbeWorker) flush(ctx context.Context, batch []ProbeResultMsg) {
	if len(batch) == 0 {
		return
	}
	b, err := w.ch.PrepareBatch(ctx, `INSERT INTO bangmod.probe_results`)
	if err != nil {
		log.Printf("ProbeWorker: prepare batch: %v", err)
		return
	}
	for _, m := range batch {
		_ = b.Append(m.Timestamp, m.HostID, m.TargetURL, m.Region, m.StatusCode, m.ResponseMS, m.IsUp)
	}
	if err := b.Send(); err != nil {
		log.Printf("ProbeWorker: batch send: %v", err)
	} else {
		log.Printf("ProbeWorker: wrote %d probe results to ClickHouse", len(batch))
	}
}
