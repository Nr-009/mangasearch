package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"github.com/spf13/cobra"
)

var rebuildCmd = &cobra.Command{
	Use:   "rebuild-index",
	Short: "Wipe and re-index everything from scratch",
	Long:  `Deletes all rows from Postgres, wipes the Elasticsearch index, then re-scans and re-indexes everything.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("⚠️  This wipes all indexed data and starts over.")
		fmt.Print("Continue? (y/N): ")

		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Println("Aborted.")
			return
		}

		apiURL := fmt.Sprintf("http://localhost:%d/rebuild", cfg.APIPort)

		resp, err := http.Post(apiURL, "application/json", nil)
		if err != nil {
			log.Fatalf("❌  Can't reach the API server. Is mangasearch running? (%v)", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			log.Fatalf("❌  Rebuild failed: %s", string(body))
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			log.Fatalf("❌  Bad response: %v", err)
		}

		log.Printf("[rebuild] ✓ %v files queued.", result["queued_jobs"])
		log.Println("[rebuild] Run `mangasearch status` to track progress.")
	},
}
