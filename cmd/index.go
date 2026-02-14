package cmd

import (
	"context"
	"fmt"
	"log"
	"mangasearch/internal/api"
	"mangasearch/internal/db"
	"mangasearch/internal/ocr"
	"mangasearch/internal/queue"
	"mangasearch/internal/search"
	"mangasearch/internal/startup"
	"mangasearch/internal/watcher"
	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "One-time scan and index of your manga folder",
	Long:  `Boots all services, scans the manga folder, indexes new or modified files, then exits.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		if err := startup.Boot(ctx, cfg); err != nil {
			log.Fatalf("❌  boot: %v", err)
		}

		dbClient, err := db.New(cfg.PostgresDSN)
		if err != nil {
			log.Fatalf("❌  postgres: %v", err)
		}
		if err := dbClient.CreateSchema(ctx); err != nil {
			log.Fatalf("❌  schema: %v", err)
		}

		esAddr := fmt.Sprintf("http://localhost:%d", cfg.ESPort)
		esClient, err := search.New(esAddr)
		if err != nil {
			log.Fatalf("❌  elasticsearch: %v", err)
		}
		if err := esClient.InitIndex(ctx); err != nil {
			log.Fatalf("❌  elasticsearch index: %v", err)
		}

		ocrClient := ocr.NewClient(cfg.OCRPort, cfg.MangaFolder, cfg.MangaFolderContainer)
		redisClient := queue.NewRedisQueue(cfg.Workers, cfg.RedisAddr, dbClient, esClient, ocrClient)
		watcherClient := watcher.NewWatcher(cfg.MangaFolder)
		server := api.NewServer(cfg, dbClient, esClient, ocrClient, redisClient, watcherClient)

		log.Println("[index] Scanning...")
		pushed, err := server.RunScan()
		if err != nil {
			log.Fatalf("❌  scan failed: %v", err)
		}

		if pushed == 0 {
			log.Println("[index] Nothing new. All caught up.")
			return
		}

		log.Printf("[index] ✓ Done. %d files indexed.", pushed)
	},
}
