package db

import (
	"context"
	"time"
)

func (db *DB) SavePage(ctx context.Context, series, chapter, page, path, text string) error {
	_, err := db.Conn.ExecContext(ctx, `
		INSERT INTO pages (path, series, chapter, page, text, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (path) DO UPDATE SET
			text       = EXCLUDED.text,
			created_at = NOW()
	`, path, series, chapter, page, text)
	return err
}

func (db *DB) LoadSnapshots(ctx context.Context) (map[string]time.Time, error) {
	rows, err := db.Conn.QueryContext(ctx, `SELECT path, created_at FROM pages`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	snapshots := make(map[string]time.Time)
	for rows.Next() {
		var path string
		var createdAt time.Time
		if err := rows.Scan(&path, &createdAt); err != nil {
			return nil, err
		}
		snapshots[path] = createdAt
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return snapshots, nil
}

func (db *DB) DeletePage(ctx context.Context, path string) error {
	_, err := db.Conn.ExecContext(ctx, `DELETE FROM pages WHERE path = $1`, path)
	return err
}

func (db *DB) DeleteAllPages(ctx context.Context) error {
	_, err := db.Conn.ExecContext(ctx, `DELETE FROM pages`)
	return err
}

func (db *DB) CountPages(ctx context.Context) (int, error) {
	var count int
	row := db.Conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM pages`)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
