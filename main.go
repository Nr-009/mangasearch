package main

import (
	"log"

	"mangasearch/cmd"
	"mangasearch/internal/config"
)

func main() {
	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatalf("❌  config: %v", err)
	}
	cmd.Execute(cfg)
}
