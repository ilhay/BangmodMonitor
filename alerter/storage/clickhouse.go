package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type CH struct {
	conn clickhouse.Conn
}

func NewCH(ctx context.Context, dsn string) (*CH, error) {
	opts, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("clickhouse parse dsn: %w", err)
	}
	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("clickhouse connect: %w", err)
	}
	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("clickhouse ping: %w", err)
	}
	return &CH{conn: conn}, nil
}

// AvgCPU returns average CPU usage for a host over the last durationSec seconds.
// Returns -1 if no data.
func (ch *CH) AvgCPU(ctx context.Context, hostID string, durationSec int) (float64, error) {
	var avg float64
	var count uint64
	err := ch.conn.QueryRow(ctx,
		`SELECT avg(cpu_percent), count() FROM bangmod.agent_metrics
		 WHERE host_id = ? AND timestamp >= now() - INTERVAL ? SECOND`,
		hostID, durationSec,
	).Scan(&avg, &count)
	if err != nil || count == 0 {
		return -1, err
	}
	return avg, nil
}

// AvgMemory returns average memory usage for a host over the last durationSec seconds.
func (ch *CH) AvgMemory(ctx context.Context, hostID string, durationSec int) (float64, error) {
	var avg float64
	var count uint64
	err := ch.conn.QueryRow(ctx,
		`SELECT avg(mem_percent), count() FROM bangmod.agent_metrics
		 WHERE host_id = ? AND timestamp >= now() - INTERVAL ? SECOND`,
		hostID, durationSec,
	).Scan(&avg, &count)
	if err != nil || count == 0 {
		return -1, err
	}
	return avg, nil
}

type ProbeStats struct {
	Total     uint64
	DownCount uint64
	AvgRespMS float64
	LastCheck time.Time
}

// ProbeStats returns probe statistics for a URL over the last durationSec seconds.
func (ch *CH) ProbeStats(ctx context.Context, targetURL string, durationSec int) (*ProbeStats, error) {
	var s ProbeStats
	err := ch.conn.QueryRow(ctx,
		`SELECT count(), countIf(is_up = 0), avg(response_ms), max(timestamp)
		 FROM bangmod.probe_results
		 WHERE target_url = ? AND timestamp >= now() - INTERVAL ? SECOND`,
		targetURL, durationSec,
	).Scan(&s.Total, &s.DownCount, &s.AvgRespMS, &s.LastCheck)
	if err != nil || s.Total == 0 {
		return nil, err
	}
	return &s, nil
}

func (ch *CH) Close() error {
	return ch.conn.Close()
}
