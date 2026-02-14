package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"mangasearch/internal/config"
	"mangasearch/internal/db"
	"mangasearch/internal/ocr"
	"mangasearch/internal/queue"
	"mangasearch/internal/search"
	"mangasearch/internal/startup"
	"mangasearch/internal/watcher"
	"os"
	"strings"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatal("config load failed:", err)
	}
	fmt.Println("config loaded ✓")

	if err := startup.Boot(ctx, cfg); err != nil {
		log.Fatal("boot failed:", err)
	}
	fmt.Println("all services ready ✓")

	database, err := db.New(cfg.PostgresDSN)
	if err != nil {
		log.Fatal("postgres connect failed:", err)
	}
	defer database.Close()
	fmt.Println("postgres connected ✓")

	if err := database.CreateSchema(ctx); err != nil {
		log.Fatal("schema init failed:", err)
	}
	fmt.Println("schema ready ✓")

	esClient, err := search.New(fmt.Sprintf("http://localhost:%d", cfg.ESPort))
	if err != nil {
		log.Fatal("elasticsearch connect failed:", err)
	}
	fmt.Println("elasticsearch connected ✓")

	if err := esClient.InitIndex(ctx); err != nil {
		log.Fatal("elasticsearch index init failed:", err)
	}
	fmt.Println("elasticsearch index ready ✓")

	ocrClient := ocr.NewClient(cfg.OCRPort, cfg.MangaFolder, cfg.MangaFolderContainer)
	q := queue.NewRedisQueue(cfg.Workers, cfg.RedisAddr, database, esClient, ocrClient)

	w := watcher.NewWatcher(cfg.MangaFolder)

	onCompare := func(toIndex []string, toDelete []string) {
		fmt.Printf("watcher: %d new/modified, %d deleted\n", len(toIndex), len(toDelete))
		for _, path := range toDelete {
			if err := database.DeletePage(ctx, path); err != nil {
				log.Printf("delete failed for %s: %v", path, err)
			}
		}
		q.Start(toIndex)
	}

	fmt.Println("scanning manga folder...")
	if err := w.Scan(ctx, database, onCompare); err != nil {
		log.Fatal("initial scan failed:", err)
	}
	fmt.Println("initial scan done ✓")

	w.Start(ctx, database, cfg.WatcherInterval, onCompare)
	fmt.Println("watcher running ✓")

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("\n--- MangaSearch ---")
	fmt.Println("commands: search <query> | quit")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "quit" {
			w.Stop()
			fmt.Println("bye")
			break
		}

		if strings.HasPrefix(input, "search ") {
			query := strings.TrimPrefix(input, "search ")
			results, err := esClient.Search(ctx, query)
			if err != nil {
				fmt.Println("search error:", err)
				continue
			}
			if len(results) == 0 {
				fmt.Println("no results found")
				continue
			}
			for _, r := range results {
				fmt.Printf("%s / %s / %s → %s\n", r.Series, r.Chapter, r.Page, r.Text)
			}
			continue
		}

		fmt.Println("unknown command — try: search <query> | quit")
	}
}
