package watcher

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"
	"mangasearch/internal/db"
)

const defaultFolder = "/manga"

type Watcher struct {
	filesFound map[string]time.Time
	mainFolder string
	stopCh     chan struct{}
}

func NewWatcher(mainFolder string) *Watcher {
	if mainFolder == "" {
		mainFolder = defaultFolder
	}
	return &Watcher{
		filesFound: make(map[string]time.Time),
		mainFolder: mainFolder,
		stopCh:     make(chan struct{}),
	}
}

func (w *Watcher) updateFiles() {
	type result struct {
		path    string
		modTime time.Time
	}

	results := make(chan result, 256)
	var wg sync.WaitGroup

	var traverse func(dir string)
	traverse = func(dir string) {
		defer wg.Done()
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, entry := range entries {
			fullPath := filepath.Join(dir, entry.Name())
			if entry.IsDir() {
				wg.Add(1)
				go traverse(fullPath)
			} else if isImageFile(entry.Name()) {
				info, err := entry.Info()
				if err != nil {
					continue
				}
				results <- result{path: fullPath, modTime: info.ModTime()}
			}
		}
	}

	wg.Add(1)
	go traverse(w.mainFolder)
	go func() {
		wg.Wait()
		close(results)
	}()

	w.filesFound = make(map[string]time.Time)
	for r := range results {
		w.filesFound[r.path] = r.modTime
	}
}

func (w *Watcher) Compare(ctx context.Context, database *db.DB) (toIndex []string, toDelete []string, err error) {
	w.updateFiles()

	savedSnapshots, err := database.LoadSnapshots(ctx)
	if err != nil {
		return nil, nil, err
	}

	for path, modTime := range w.filesFound {
		savedTime, exists := savedSnapshots[path]
		if !exists {
			toIndex = append(toIndex, path)
		} else if modTime.After(savedTime) {
			toIndex = append(toIndex, path)
		}
	}

	for path := range savedSnapshots {
		if _, exists := w.filesFound[path]; !exists {
			toDelete = append(toDelete, path)
		}
	}

	return toIndex, toDelete, nil
}

func (w *Watcher) Start(ctx context.Context, database *db.DB, interval time.Duration, onCompare func(toIndex []string, toDelete []string)) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				toIndex, toDelete, err := w.Compare(ctx, database)
				if err != nil {
					continue
				}
				onCompare(toIndex, toDelete)
			case <-w.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (w *Watcher) Scan(ctx context.Context, database *db.DB, onCompare func(toIndex []string, toDelete []string)) error {
	toIndex, toDelete, err := w.Compare(ctx, database)
	if err != nil {
		return err
	}
	onCompare(toIndex, toDelete)
	return nil
}

func (w *Watcher) Stop() {
	close(w.stopCh)
}

func isImageFile(name string) bool {
	ext := filepath.Ext(name)
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png"
}
