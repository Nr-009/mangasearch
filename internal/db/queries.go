package db

func (db *DB) SavePage(series, chapter, page, path, text string) error {
	_, err := db.Conn.Exec(`
		INSERT INTO pages (series, chapter, page, path, text)
		VALUES ($1, $2, $3, $4, $5)
	`, series, chapter, page, path, text)
	return err
}
