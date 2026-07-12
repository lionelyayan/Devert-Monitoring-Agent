package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// DB wraps a pgxpool connection pool.
type DB struct {
	Pool *pgxpool.Pool
}

// Connect establishes a PostgreSQL connection pool and runs database migrations.
func Connect(ctx context.Context, dsn string) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("database: failed to parse DSN: %w", err)
	}

	cfg.MaxConns = 10
	cfg.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("database: failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database: ping failed: %w", err)
	}

	db := &DB{Pool: pool}

	if err := db.migrate(ctx); err != nil {
		return nil, fmt.Errorf("database: migration failed: %w", err)
	}

	log.Info().Msg("database: connected and migrated successfully")
	return db, nil
}

// Close releases the pool connections.
func (db *DB) Close() {
	db.Pool.Close()
}

// migrate runs the inline schema migration (idempotent CREATE IF NOT EXISTS).
func (db *DB) migrate(ctx context.Context) error {
	schema := `
CREATE TABLE IF NOT EXISTS event_logs (
    id          BIGSERIAL PRIMARY KEY,
    server      VARCHAR(255)    NOT NULL DEFAULT '',
    container   VARCHAR(255)    NOT NULL DEFAULT '',
    image       VARCHAR(255)    NOT NULL DEFAULT '',
    action      VARCHAR(100)    NOT NULL,
    event_type  VARCHAR(50)     NOT NULL DEFAULT 'docker',
    status      VARCHAR(50)     NOT NULL DEFAULT '',
    message     TEXT            NOT NULL DEFAULT '',
    payload     JSONB           NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_event_logs_server     ON event_logs (server);
CREATE INDEX IF NOT EXISTS idx_event_logs_container  ON event_logs (container);
CREATE INDEX IF NOT EXISTS idx_event_logs_action     ON event_logs (action);
CREATE INDEX IF NOT EXISTS idx_event_logs_event_type ON event_logs (event_type);
CREATE INDEX IF NOT EXISTS idx_event_logs_created_at ON event_logs (created_at DESC);
`
	if _, err := db.Pool.Exec(ctx, schema); err != nil {
		return fmt.Errorf("database: migration failed: %w", err)
	}
	log.Debug().Msg("database: schema migration applied")
	return nil
}
