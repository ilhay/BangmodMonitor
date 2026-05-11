package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Maria struct {
	db *sql.DB
}

func NewMaria(ctx context.Context, dsn string) (*Maria, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mariadb open: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(3)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("mariadb ping: %w", err)
	}
	return &Maria{db: db}, nil
}

func (m *Maria) Migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS alert_rules (
			id             CHAR(36) NOT NULL PRIMARY KEY,
			name           VARCHAR(255) NOT NULL,
			host_id        CHAR(36),
			condition_type VARCHAR(32) NOT NULL,
			threshold      FLOAT NOT NULL DEFAULT 0,
			duration_sec   INT NOT NULL DEFAULT 60,
			target_url     VARCHAR(2048),
			channel        VARCHAR(32) NOT NULL,
			channel_config TEXT NOT NULL,
			enabled        TINYINT(1) NOT NULL DEFAULT 1,
			created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			KEY idx_enabled (enabled)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`CREATE TABLE IF NOT EXISTS alert_incidents (
			id          CHAR(36) NOT NULL PRIMARY KEY,
			rule_id     CHAR(36) NOT NULL,
			started_at  TIMESTAMP NOT NULL,
			resolved_at TIMESTAMP NULL DEFAULT NULL,
			details     TEXT,
			KEY idx_rule_open (rule_id, resolved_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	for _, s := range stmts {
		if _, err := m.db.ExecContext(ctx, s); err != nil {
			return fmt.Errorf("alert migrate: %w", err)
		}
	}
	return nil
}

type AlertRule struct {
	ID            string
	Name          string
	HostID        string
	ConditionType string // cpu_high | memory_high | probe_down | probe_slow
	Threshold     float64
	DurationSec   int
	TargetURL     string
	Channel       string
	ChannelConfig string // JSON
}

func (m *Maria) GetEnabledRules(ctx context.Context) ([]AlertRule, error) {
	rows, err := m.db.QueryContext(ctx,
		`SELECT id, name, COALESCE(host_id,''), condition_type, threshold, duration_sec,
		        COALESCE(target_url,''), channel, channel_config
		 FROM alert_rules WHERE enabled = 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []AlertRule
	for rows.Next() {
		var r AlertRule
		if err := rows.Scan(&r.ID, &r.Name, &r.HostID, &r.ConditionType,
			&r.Threshold, &r.DurationSec, &r.TargetURL, &r.Channel, &r.ChannelConfig); err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

// OpenIncident returns the open incident ID for a rule, or "" if none.
func (m *Maria) OpenIncident(ctx context.Context, ruleID string) (string, error) {
	var id string
	err := m.db.QueryRowContext(ctx,
		`SELECT id FROM alert_incidents WHERE rule_id = ? AND resolved_at IS NULL LIMIT 1`,
		ruleID,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return id, err
}

func (m *Maria) CreateIncident(ctx context.Context, id, ruleID, details string) error {
	_, err := m.db.ExecContext(ctx,
		`INSERT INTO alert_incidents (id, rule_id, started_at, details) VALUES (?, ?, NOW(), ?)`,
		id, ruleID, details,
	)
	return err
}

func (m *Maria) ResolveIncident(ctx context.Context, incidentID string) error {
	_, err := m.db.ExecContext(ctx,
		`UPDATE alert_incidents SET resolved_at = NOW() WHERE id = ?`,
		incidentID,
	)
	return err
}

func (m *Maria) Close() error {
	return m.db.Close()
}
