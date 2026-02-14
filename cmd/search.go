package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:     "search [query]",
	Short:   "Search your manga collection by quote",
	Args:    cobra.ExactArgs(1),
	Example: `  mangasearch search "I sacrifice"`,
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]

		apiURL := fmt.Sprintf(
			"http://localhost:%d/search?q=%s",
			cfg.APIPort,
			url.QueryEscape(query),
		)

		resp, err := http.Get(apiURL)
		if err != nil {
			log.Fatalf("❌  Can't reach the API server. Is mangasearch running? (%v)", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			log.Fatalf("❌  Search failed: %s", string(body))
		}

		var results []map[string]interface{}
		if err := json.Unmarshal(body, &results); err != nil {
			log.Fatalf("❌  Bad response: %v", err)
		}

		if len(results) == 0 {
			fmt.Println("No results found.")
			os.Exit(0)
		}

		fmt.Printf("\nResults for \"%s\":\n\n", query)
		for i, r := range results {
			fmt.Printf(
				"  %d. %s — Chapter %v, Page %v\n     \"%v\"\n\n",
				i+1,
				r["series"],
				r["chapter"],
				r["page"],
				r["text"],
			)
		}
	},
}
