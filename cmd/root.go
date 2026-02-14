package cmd

import (
	"os"
	"mangasearch/internal/config"
	"github.com/spf13/cobra"
)

var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:   "mangasearch",
	Short: "Search your manga collection by quote",
	Long: `MangaSearch indexes your manga collection and lets you search for quotes.

  mangasearch start                boot everything and watch for new files
  mangasearch index                one-time scan and index
  mangasearch search "I sacrifice" find that panel
  mangasearch status               see what's indexed and in queue
  mangasearch rebuild-index        wipe and re-index everything`,
}

func Execute(c *config.Config) {
	cfg = c
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(indexCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(rebuildCmd)
}
