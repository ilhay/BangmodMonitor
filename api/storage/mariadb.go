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
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("mariadb ping: %w", err)
	}
	return &Maria{db: db}, nil
}

func (m *Maria) Migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS hosts (
			id           CHAR(36) NOT NULL PRIMARY KEY,
			name         VARCHAR(255) NOT NULL,
			token_hash   CHAR(64) NOT NULL,
			region       VARCHAR(64) NOT NULL DEFAULT 'default',
			created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uniq_token_hash (token_hash)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`CREATE TABLE IF NOT EXISTS probe_targets (
			id            CHAR(36) NOT NULL PRIMARY KEY,
			host_id       CHAR(36) NOT NULL,
			url           VARCHAR(2048) NOT NULL,
			interval_sec  INT NOT NULL DEFAULT 60,
			enabled       TINYINT(1) NOT NULL DEFAULT 1,
			created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			KEY idx_host (host_id),
			CONSTRAINT fk_probe_host FOREIGN KEY (host_id) REFERENCES hosts(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	for _, s := range stmts {
		if _, err := m.db.ExecContext(ctx, s); err != nil {
			return fmt.Errorf("mariadb migrate: %w", err)
		}
	}
	return nil
}

func (m *Maria) ValidateToken(ctx context.Context, tokenHash string) (hostID string, ok bool) {
	var id string
	err := m.db.QueryRowContext(ctx,
		`SELECT id FROM hosts WHERE token_hash = ? LIMIT 1`, tokenHash,
	).Scan(&id)
	if err != nil {
		return "", false
	}
	return id, true
}

func (m *Maria) Close() error {
	return m.db.Close()
}
