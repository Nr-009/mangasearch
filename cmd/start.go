package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"mangasearch/internal/api"
	"mangasearch/internal/db"
	"mangasearch/internal/ocr"
	"mangasearch/internal/queue"
	"mangasearch/internal/search"
	"mangasearch/internal/startup"
	"mangasearch/internal/watcher"

	"github.com/spf13/cobra"
)

var withDocker bool

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Boot everything and start watching for new files",
	Long:  `Boots Docker, inits all services, runs initial scan, starts API server and file watcher. Blocks until Ctrl+C.`,
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
		log.Printf("✓  postgres connected")

		esAddr := fmt.Sprintf("http://localhost:%d", cfg.ESPort)
		esClient, err := search.New(esAddr)
		if err != nil {
			log.Fatalf("❌  elasticsearch: %v", err)
		}
		if err := esClient.InitIndex(ctx); err != nil {
			log.Fatalf("❌  elasticsearch index: %v", err)
		}
		log.Printf("✓  elasticsearch connected")

		ocrClient := ocr.NewClient(cfg.OCRPort, cfg.MangaFolder, cfg.MangaFolderContainer)
		log.Printf("✓  ocr client configured")

		redisClient := queue.NewRedisQueue(cfg.Workers, cfg.RedisAddr, dbClient, esClient, ocrClient)
		log.Printf("✓  redis connected")

		watcherClient := watcher.NewWatcher(cfg.MangaFolder)
		server := api.NewServer(cfg, dbClient, esClient, ocrClient, redisClient, watcherClient)

		log.Println("[start] Running initial scan...")
		if _, err := server.RunScan(); err != nil {
			log.Fatalf("❌  initial scan: %v", err)
		}
		log.Println("[start] Initial scan done.")

		server.StartWatcher()
		log.Println("[start] Watcher running.")

		go func() {
			log.Println("[start] API server starting...")
			if err := server.Run(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("[start] API server died: %v", err)
			}
		}()

		log.Println("[start] Everything is up. Ctrl+C to stop.")

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit

		log.Println("\n[start] Shutting down...")

		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutCtx); err != nil {
			log.Printf("[start] Forced shutdown: %v", err)
		}

		server.StopWatcher()

		if withDocker {
			log.Println("[start] Bringing Docker down...")
			c := exec.Command("docker", "compose", "down")
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			c.Run()
		}

		log.Println("[start] Bye.")
	},
}

func init() {
	startCmd.Flags().BoolVar(&withDocker, "with-docker", false, "run docker compose down on exit")
}
