package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show indexed pages and queue length",
	Run: func(cmd *cobra.Command, args []string) {
		apiURL := fmt.Sprintf("http://localhost:%d/status", cfg.APIPort)

		resp, err := http.Get(apiURL)
		if err != nil {
			log.Fatalf("❌  Can't reach the API server. Is mangasearch running? (%v)", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			log.Fatalf("❌  Status check failed: %s", string(body))
		}

		var status map[string]interface{}
		if err := json.Unmarshal(body, &status); err != nil {
			log.Fatalf("❌  Bad response: %v", err)
		}

		fmt.Println("\nMangaSearch Status:")
		fmt.Println("─────────────────────")
		fmt.Printf("  ✓  Indexed   : %v\n", status["indexed"])
		fmt.Printf("  📥  In queue  : %v\n", status["in_queue"])
		fmt.Println("─────────────────────")
	},
}
