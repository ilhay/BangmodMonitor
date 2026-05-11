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

func (ch *CH) Migrate(ctx context.Context) error {
	queries := []string{
		`CREATE DATABASE IF NOT EXISTS bangmod`,
		`CREATE TABLE IF NOT EXISTS bangmod.agent_metrics (
			timestamp     DateTime,
			host_id       String,
			hostname      String,
			region        String,
			cpu_percent   Float32,
			cpu_cores     UInt8,
			mem_total     UInt64,
			mem_used      UInt64,
			mem_percent   Float32
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMM(timestamp)
		ORDER BY (host_id, timestamp)
		TTL timestamp + INTERVAL 30 DAY`,

		`CREATE TABLE IF NOT EXISTS bangmod.disk_metrics (
			timestamp   DateTime,
			host_id     String,
			path        String,
			total_bytes UInt64,
			used_bytes  UInt64,
			use_percent Float32
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMM(timestamp)
		ORDER BY (host_id, path, timestamp)
		TTL timestamp + INTERVAL 30 DAY`,

		`CREATE TABLE IF NOT EXISTS bangmod.net_metrics (
			timestamp    DateTime,
			host_id      String,
			interface    String,
			bytes_sent   UInt64,
			bytes_recv   UInt64,
			packets_sent UInt64,
			packets_recv UInt64
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMM(timestamp)
		ORDER BY (host_id, interface, timestamp)
		TTL timestamp + INTERVAL 30 DAY`,

		`CREATE TABLE IF NOT EXISTS bangmod.probe_results (
			timestamp       DateTime,
			host_id         String,
			target_url      String,
			region          String,
			status_code     UInt16,
			response_ms     UInt32,
			is_up           UInt8
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMM(timestamp)
		ORDER BY (host_id, region, timestamp)
		TTL timestamp + INTERVAL 30 DAY`,

		// ── Materialized views for fast dashboard queries (Phase 5) ──────────

		// Per-minute rollup: avg CPU + memory per host
		`CREATE MATERIALIZED VIEW IF NOT EXISTS bangmod.agent_metrics_1m
		ENGINE = AggregatingMergeTree()
		PARTITION BY toYYYYMM(ts_minute)
		ORDER BY (host_id, ts_minute)
		TTL ts_minute + INTERVAL 90 DAY
		AS SELECT
			host_id,
			toStartOfMinute(timestamp) AS ts_minute,
			avgState(cpu_percent)      AS cpu_avg_state,
			avgState(mem_percent)      AS mem_avg_state,
			maxState(cpu_percent)      AS cpu_max_state,
			maxState(mem_percent)      AS mem_max_state
		FROM bangmod.agent_metrics
		GROUP BY host_id, ts_minute`,

		// Per-hour rollup: for longer time ranges
		`CREATE MATERIALIZED VIEW IF NOT EXISTS bangmod.agent_metrics_1h
		ENGINE = AggregatingMergeTree()
		PARTITION BY toYYYYMM(ts_hour)
		ORDER BY (host_id, ts_hour)
		TTL ts_hour + INTERVAL 365 DAY
		AS SELECT
			host_id,
			toStartOfHour(timestamp) AS ts_hour,
			avgState(cpu_percent)    AS cpu_avg_state,
			avgState(mem_percent)    AS mem_avg_state,
			maxState(cpu_percent)    AS cpu_max_state,
			maxState(mem_percent)    AS mem_max_state
		FROM bangmod.agent_metrics
		GROUP BY host_id, ts_hour`,

		// Per-minute probe uptime rollup
		`CREATE MATERIALIZED VIEW IF NOT EXISTS bangmod.probe_results_1m
		ENGINE = AggregatingMergeTree()
		PARTITION BY toYYYYMM(ts_minute)
		ORDER BY (target_url, region, ts_minute)
		TTL ts_minute + INTERVAL 90 DAY
		AS SELECT
			target_url,
			region,
			toStartOfMinute(timestamp) AS ts_minute,
			avgState(toFloat32(is_up)) AS uptime_state,
			avgState(response_ms)      AS resp_ms_avg_state
		FROM bangmod.probe_results
		GROUP BY target_url, region, ts_minute`,
	}

	for _, q := range queries {
		if err := ch.conn.Exec(ctx, q); err != nil {
			return fmt.Errorf("clickhouse migrate: %w", err)
		}
	}
	return nil
}

type AgentMetricRow struct {
	Timestamp  time.Time
	HostID     string
	Hostname   string
	Region     string
	CPUPercent float32
	CPUCores   uint8
	MemTotal   uint64
	MemUsed    uint64
	MemPercent float32
}

func (ch *CH) InsertAgentMetric(ctx context.Context, r AgentMetricRow) error {
	return ch.conn.Exec(ctx,
		`INSERT INTO bangmod.agent_metrics VALUES (?,?,?,?,?,?,?,?,?)`,
		r.Timestamp, r.HostID, r.Hostname, r.Region,
		r.CPUPercent, r.CPUCores, r.MemTotal, r.MemUsed, r.MemPercent,
	)
}

type DiskRow struct {
	Timestamp  time.Time
	HostID     string
	Path       string
	TotalBytes uint64
	UsedBytes  uint64
	UsePct     float32
}

func (ch *CH) InsertDiskMetrics(ctx context.Context, rows []DiskRow) error {
	batch, err := ch.conn.PrepareBatch(ctx, `INSERT INTO bangmod.disk_metrics`)
	if err != nil {
		return err
	}
	for _, r := range rows {
		if err := batch.Append(r.Timestamp, r.HostID, r.Path, r.TotalBytes, r.UsedBytes, r.UsePct); err != nil {
			return err
		}
	}
	return batch.Send()
}

type NetRow struct {
	Timestamp   time.Time
	HostID      string
	Interface   string
	BytesSent   uint64
	BytesRecv   uint64
	PacketsSent uint64
	PacketsRecv uint64
}

func (ch *CH) InsertNetMetrics(ctx context.Context, rows []NetRow) error {
	batch, err := ch.conn.PrepareBatch(ctx, `INSERT INTO bangmod.net_metrics`)
	if err != nil {
		return err
	}
	for _, r := range rows {
		if err := batch.Append(r.Timestamp, r.HostID, r.Interface, r.BytesSent, r.BytesRecv, r.PacketsSent, r.PacketsRecv); err != nil {
			return err
		}
	}
	return batch.Send()
}

type ProbeRow struct {
	Timestamp  time.Time
	HostID     string
	TargetURL  string
	Region     string
	StatusCode uint16
	ResponseMS uint32
	IsUp       uint8
}

func (ch *CH) InsertProbeResult(ctx context.Context, r ProbeRow) error {
	return ch.conn.Exec(ctx,
		`INSERT INTO bangmod.probe_results VALUES (?,?,?,?,?,?,?)`,
		r.Timestamp, r.HostID, r.TargetURL, r.Region, r.StatusCode, r.ResponseMS, r.IsUp,
	)
}

type MetricPoint struct {
	Timestamp  time.Time `json:"timestamp"`
	CPUPercent float32   `json:"cpu_percent"`
	MemPercent float32   `json:"mem_percent"`
}

func (ch *CH) QueryRecentMetrics(ctx context.Context, hostID string, limit int) ([]MetricPoint, error) {
	rows, err := ch.conn.Query(ctx,
		`SELECT timestamp, cpu_percent, mem_percent
		 FROM bangmod.agent_metrics
		 WHERE host_id = ?
		 ORDER BY timestamp DESC
		 LIMIT ?`,
		hostID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []MetricPoint
	for rows.Next() {
		var p MetricPoint
		if err := rows.Scan(&p.Timestamp, &p.CPUPercent, &p.MemPercent); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

type ProbeResult struct {
	Timestamp  time.Time `json:"timestamp"`
	TargetURL  string    `json:"url"`
	Region     string    `json:"region"`
	StatusCode uint16    `json:"status_code"`
	ResponseMS uint32    `json:"response_ms"`
	IsUp       bool      `json:"is_up"`
}

func (ch *CH) QueryProbeResults(ctx context.Context, url, region string, limit int) ([]ProbeResult, error) {
	query := `SELECT timestamp, target_url, region, status_code, response_ms, is_up
			  FROM bangmod.probe_results
			  WHERE 1=1`
	args := []any{}
	if url != "" {
		query += ` AND target_url = ?`
		args = append(args, url)
	}
	if region != "" {
		query += ` AND region = ?`
		args = append(args, region)
	}
	query += ` ORDER BY timestamp DESC LIMIT ?`
	args = append(args, limit)

	rows, err := ch.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ProbeResult
	for rows.Next() {
		var r ProbeResult
		var isUp uint8
		if err := rows.Scan(&r.Timestamp, &r.TargetURL, &r.Region, &r.StatusCode, &r.ResponseMS, &isUp); err != nil {
			return nil, err
		}
		r.IsUp = isUp == 1
		results = append(results, r)
	}
	return results, rows.Err()
}

// QueryRollupMetrics queries the per-minute materialized view for fast dashboard rendering.
// Falls back to raw table if the rollup has no data yet.
func (ch *CH) QueryRollupMetrics(ctx context.Context, hostID string, limit int) ([]MetricPoint, error) {
	rows, err := ch.conn.Query(ctx,
		`SELECT ts_minute, avgMerge(cpu_avg_state), avgMerge(mem_avg_state)
		 FROM bangmod.agent_metrics_1m
		 WHERE host_id = ?
		 GROUP BY ts_minute
		 ORDER BY ts_minute DESC
		 LIMIT ?`,
		hostID, limit,
	)
	if err != nil {
		return ch.QueryRecentMetrics(ctx, hostID, limit) // fallback
	}
	defer rows.Close()

	var points []MetricPoint
	for rows.Next() {
		var p MetricPoint
		if err := rows.Scan(&p.Timestamp, &p.CPUPercent, &p.MemPercent); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	if len(points) == 0 {
		return ch.QueryRecentMetrics(ctx, hostID, limit) // fallback to raw
	}
	return points, rows.Err()
}

func (ch *CH) Close() error {
	return ch.conn.Close()
}
