package startup

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"
	"mangasearch/internal/config"
)

func Boot(ctx context.Context, cfg *config.Config) error {
	if err := startDocker(); err != nil {
		return fmt.Errorf("docker compose: %w", err)
	}
	if err := waitForPostgres(ctx, cfg); err != nil {
		return fmt.Errorf("postgres not ready: %w", err)
	}
	if err := waitForRedis(ctx, cfg); err != nil {
		return fmt.Errorf("redis not ready: %w", err)
	}
	if err := waitForElasticsearch(ctx, cfg); err != nil {
		return fmt.Errorf("elasticsearch not ready: %w", err)
	}
	if err := waitForOCRServer(ctx, cfg); err != nil {
		return fmt.Errorf("ocr server not ready: %w", err)
	}
	return nil
}

func startDocker() error {
	fmt.Println("starting docker compose...")
	cmd := exec.Command("docker", "compose", "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func waitForPostgres(ctx context.Context, cfg *config.Config) error {
	fmt.Println("waiting for postgres...")
	return retry(ctx, 30, 2*time.Second, func() error {
		cmd := exec.Command("docker", "exec", "mangasearch-postgres-1",
			"pg_isready", "-U", "manga", "-d", "mangasearch")
		return cmd.Run()
	})
}

func waitForRedis(ctx context.Context, cfg *config.Config) error {
	fmt.Println("waiting for redis...")
	return retry(ctx, 30, 2*time.Second, func() error {
		cmd := exec.Command("docker", "exec", "mangasearch-redis-1",
			"redis-cli", "ping")
		return cmd.Run()
	})
}

func waitForElasticsearch(ctx context.Context, cfg *config.Config) error {
	fmt.Println("waiting for elasticsearch...")
	url := fmt.Sprintf("http://localhost:%d/_cluster/health", cfg.ESPort)
	return retry(ctx, 60, 3*time.Second, func() error {
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("status %d", resp.StatusCode)
		}
		return nil
	})
}

func waitForOCRServer(ctx context.Context, cfg *config.Config) error {
	fmt.Println("waiting for ocr server...")
	url := fmt.Sprintf("http://localhost:%d/health", cfg.OCRPort)
	if err := retry(ctx, 60, 3*time.Second, func() error {
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	}); err != nil {
		return err
	}
	fmt.Println("ocr server up — waiting for workers to stabilize...")
	time.Sleep(time.Duration(cfg.Workers) * 3 * time.Second)
	return nil
}

func retry(ctx context.Context, attempts int, delay time.Duration, fn func() error) error {
	for i := 0; i < attempts; i++ {
		if err := fn(); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return fmt.Errorf("timed out after %d attempts", attempts)
}
