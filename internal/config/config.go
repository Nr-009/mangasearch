package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
	"github.com/joho/godotenv"
)

type Config struct {
	MangaFolder          string
	MangaFolderContainer string
	Workers              int
	OCRPort              int
	ESPort               int
	APIPort              int
	PostgresDSN          string
	RedisAddr            string
	WatcherInterval      time.Duration
}

func Load(envPath string) (*Config, error) {
	if err := godotenv.Load(envPath); err != nil {
		return nil, fmt.Errorf("config.Load: %w", err)
	}

	cfg := &Config{}
	cfg.MangaFolder = os.Getenv("MANGA_FOLDER")
	cfg.MangaFolderContainer = os.Getenv("MANGA_FOLDER_CONTAINER")
	if cfg.MangaFolder == "" {
		return nil, fmt.Errorf("MANGA_FOLDER is required in .env")
	}

	var err error
	cfg.Workers, err = parseInt("WORKERS", 1)
	if err != nil {
		return nil, err
	}

	cfg.OCRPort, err = parseInt("OCR_PORT", 5001)
	if err != nil {
		return nil, err
	}

	cfg.ESPort, err = parseInt("ES_PORT", 9200)
	if err != nil {
		return nil, err
	}

	cfg.APIPort, err = parseInt("API_PORT", 8080)
	if err != nil {
		return nil, err
	}

	postgresPort, err := parseInt("POSTGRES_PORT", 5432)
	if err != nil {
		return nil, err
	}
	cfg.PostgresDSN = fmt.Sprintf("postgres://%s:%s@localhost:%d/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		postgresPort,
		os.Getenv("POSTGRES_DB"),
	)

	redisPort, err := parseInt("REDIS_PORT", 6379)
	if err != nil {
		return nil, err
	}
	cfg.RedisAddr = fmt.Sprintf("localhost:%d", redisPort)

	cfg.WatcherInterval, err = parseDuration("WATCHER_INTERVAL", 30*time.Minute)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func parseInt(key string, defaultVal int) (int, error) {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal, nil
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("%s invalid: %w", key, err)
	}
	return n, nil
}

func parseDuration(key string, defaultVal time.Duration) (time.Duration, error) {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal, nil
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return 0, fmt.Errorf("%s invalid: %w", key, err)
	}
	return d, nil
}
