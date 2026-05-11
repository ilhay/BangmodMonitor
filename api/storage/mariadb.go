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
		`CREATE TABLE IF NOT EXISTS orgs (
			id         CHAR(36) NOT NULL PRIMARY KEY,
			name       VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`CREATE TABLE IF NOT EXISTS users (
			id            CHAR(36) NOT NULL PRIMARY KEY,
			org_id        CHAR(36) NOT NULL,
			email         VARCHAR(255) NOT NULL,
			password_hash VARCHAR(72) NOT NULL,
			role          VARCHAR(32) NOT NULL DEFAULT 'admin',
			created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uniq_email (email),
			KEY idx_org (org_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`CREATE TABLE IF NOT EXISTS hosts (
			id          CHAR(36) NOT NULL PRIMARY KEY,
			org_id      CHAR(36),
			name        TEXT NOT NULL,
			token_hash  CHAR(64) NOT NULL,
			region      VARCHAR(64) NOT NULL DEFAULT 'default',
			created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uniq_token_hash (token_hash),
			KEY idx_org (org_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`CREATE TABLE IF NOT EXISTS probe_targets (
			id            CHAR(36) NOT NULL PRIMARY KEY,
			host_id       CHAR(36) NOT NULL,
			url           VARCHAR(2048) NOT NULL,
			interval_sec  INT NOT NULL DEFAULT 60,
			enabled       TINYINT(1) NOT NULL DEFAULT 1,
			created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			KEY idx_host (host_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		// Add org_id column to hosts if missing (safe on existing installs)
		`ALTER TABLE hosts ADD COLUMN IF NOT EXISTS org_id CHAR(36) AFTER id`,
	}
	for _, s := range stmts {
		if _, err := m.db.ExecContext(ctx, s); err != nil {
			return fmt.Errorf("mariadb migrate: %w", err)
		}
	}
	return nil
}

// ── Org ──────────────────────────────────────────────────────────────────────

func (m *Maria) CreateOrg(ctx context.Context, id, name string) error {
	_, err := m.db.ExecContext(ctx,
		`INSERT INTO orgs (id, name) VALUES (?, ?)`, id, name)
	return err
}

// ── User ─────────────────────────────────────────────────────────────────────

type User struct {
	ID           string
	OrgID        string
	Email        string
	PasswordHash string
	Role         string
}

func (m *Maria) CreateUser(ctx context.Context, id, orgID, email, passwordHash, role string) error {
	_, err := m.db.ExecContext(ctx,
		`INSERT INTO users (id, org_id, email, password_hash, role) VALUES (?, ?, ?, ?, ?)`,
		id, orgID, email, passwordHash, role)
	return err
}

func (m *Maria) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	u := &User{}
	err := m.db.QueryRowContext(ctx,
		`SELECT id, org_id, email, password_hash, role FROM users WHERE email = ? LIMIT 1`, email,
	).Scan(&u.ID, &u.OrgID, &u.Email, &u.PasswordHash, &u.Role)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

// ── Host ─────────────────────────────────────────────────────────────────────

type Host struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	Name      string    `json:"name"`
	Region    string    `json:"region"`
	CreatedAt time.Time `json:"created_at"`
}

func (m *Maria) CreateHost(ctx context.Context, id, orgID, name, tokenHash, region string) error {
	_, err := m.db.ExecContext(ctx,
		`INSERT INTO hosts (id, org_id, name, token_hash, region) VALUES (?, ?, ?, ?, ?)`,
		id, orgID, name, tokenHash, region)
	return err
}

func (m *Maria) ListHosts(ctx context.Context, orgID string) ([]Host, error) {
	rows, err := m.db.QueryContext(ctx,
		`SELECT id, COALESCE(org_id,''), name, region, created_at FROM hosts WHERE org_id = ? ORDER BY created_at DESC`,
		orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []Host
	for rows.Next() {
		var h Host
		if err := rows.Scan(&h.ID, &h.OrgID, &h.Name, &h.Region, &h.CreatedAt); err != nil {
			return nil, err
		}
		hosts = append(hosts, h)
	}
	return hosts, rows.Err()
}

func (m *Maria) GetHost(ctx context.Context, id, orgID string) (*Host, error) {
	h := &Host{}
	err := m.db.QueryRowContext(ctx,
		`SELECT id, COALESCE(org_id,''), name, region, created_at FROM hosts WHERE id = ? AND org_id = ?`,
		id, orgID,
	).Scan(&h.ID, &h.OrgID, &h.Name, &h.Region, &h.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return h, err
}

func (m *Maria) DeleteHost(ctx context.Context, id, orgID string) error {
	res, err := m.db.ExecContext(ctx,
		`DELETE FROM hosts WHERE id = ? AND org_id = ?`, id, orgID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("host not found or not owned by org")
	}
	return nil
}

func (m *Maria) RotateToken(ctx context.Context, id, orgID, newTokenHash string) error {
	res, err := m.db.ExecContext(ctx,
		`UPDATE hosts SET token_hash = ? WHERE id = ? AND org_id = ?`,
		newTokenHash, id, orgID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("host not found or not owned by org")
	}
	return nil
}

// ValidateToken checks a SHA-256 hashed agent token. Returns hostID and orgID.
func (m *Maria) ValidateToken(ctx context.Context, tokenHash string) (hostID, orgID string, ok bool) {
	var hid, oid string
	err := m.db.QueryRowContext(ctx,
		`SELECT id, COALESCE(org_id,'') FROM hosts WHERE token_hash = ? LIMIT 1`, tokenHash,
	).Scan(&hid, &oid)
	if err != nil {
		return "", "", false
	}
	return hid, oid, true
}

func (m *Maria) Close() error {
	return m.db.Close()
}
