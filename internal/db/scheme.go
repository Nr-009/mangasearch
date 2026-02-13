package db

func (db *DB) InitSchema() error {
	_, err := db.Conn.Exec(`
		CREATE TABLE IF NOT EXISTS pages (
			id         SERIAL PRIMARY KEY,
			series     TEXT        NOT NULL,
			chapter    TEXT        NOT NULL,
			page       TEXT        NOT NULL,
			path       TEXT        NOT NULL,
			text       TEXT        NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}