package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type DB struct {
	Conn *sql.DB
}

func New(dsn string) (*DB, error) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("db.New: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("db.New ping: %w", err)
	}

	return &DB{Conn: conn}, nil
}

func (db *DB) Close() error {
	return db.Conn.Close()
}
