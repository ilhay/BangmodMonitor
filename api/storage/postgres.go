package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PG struct {
	pool *pgxpool.Pool
}

func NewPG(ctx context.Context, dsn string) (*PG, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pg connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pg ping: %w", err)
	}
	return &PG{pool: pool}, nil
}

func (pg *PG) Migrate(ctx context.Context) error {
	_, err := pg.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS hosts (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name        TEXT NOT NULL,
			token_hash  TEXT NOT NULL UNIQUE,
			region      TEXT NOT NULL DEFAULT 'default',
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS probe_targets (
			id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			host_id    UUID REFERENCES hosts(id) ON DELETE CASCADE,
			url        TEXT NOT NULL,
			interval   INT NOT NULL DEFAULT 60,
			enabled    BOOL NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	return err
}

// ValidateToken checks if a bearer token (plain) matches a stored SHA-256 hash.
// Phase 1: simple lookup by hash. Phase 5: move to Redis.
func (pg *PG) ValidateToken(ctx context.Context, tokenHash string) (hostID string, ok bool) {
	var id string
	err := pg.pool.QueryRow(ctx,
		`SELECT id FROM hosts WHERE token_hash = $1`, tokenHash,
	).Scan(&id)
	if err != nil {
		return "", false
	}
	return id, true
}

func (pg *PG) Close() {
	pg.pool.Close()
}
