package db

import "context"

func (db *DB) CreateSchema(ctx context.Context) error {
	_, err := db.Conn.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS pages (
			path       TEXT PRIMARY KEY,
			series     TEXT        NOT NULL,
			chapter    TEXT        NOT NULL,
			page       TEXT        NOT NULL,
			text       TEXT        NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}
