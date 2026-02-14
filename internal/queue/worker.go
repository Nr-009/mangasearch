package queue

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"mangasearch/internal/db"
	"mangasearch/internal/ocr"
	"mangasearch/internal/search"
)

func parsePath(path string) (series, chapter, page string, err error) {
	parts := strings.Split(filepath.ToSlash(path), "/")
	if len(parts) < 3 {
		return "", "", "", fmt.Errorf("path too short: %q", path)
	}
	series = parts[len(parts)-3]
	chapter = parts[len(parts)-2]
	page = parts[len(parts)-1]
	return series, chapter, page, nil
}

func process(dataPath string, database *db.DB, esClient *search.Client, ocrClient *ocr.Client, id int) error {
	series, chapter, page, err := parsePath(dataPath)
	if err != nil {
		return fmt.Errorf("parsePath: %w", err)
	}

	text, err := ocrClient.GetData(dataPath)
	if err != nil {
		fmt.Printf("[worker %d] ocr error: %v\n", id, err)
		return err
	}
	fmt.Printf("[worker %d] OCR done — %s / %s / %s\n", id, series, chapter, page)

	if err := database.SavePage(context.Background(), series, chapter, page, dataPath, text); err != nil {
		return fmt.Errorf("SavePage: %w", err)
	}
	fmt.Printf("[worker %d] ✓ saved %s / %s / %s\n", id, series, chapter, page)

	if err := esClient.IndexPage(context.Background(), series, chapter, page, dataPath, text); err != nil {
		return fmt.Errorf("IndexPage: %w", err)
	}
	fmt.Printf("[worker %d] ✓ indexed %s / %s / %s\n", id, series, chapter, page)

	return nil
}
