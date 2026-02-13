package main

import (
	"context"
	"fmt"
	"log"

	"mangasearch/internal/db"
	"mangasearch/internal/queue"
	"mangasearch/internal/search"
)

func main() {
	ctx := context.Background()

	database, err := db.New("postgres://manga:manga@localhost:5432/mangasearch?sslmode=disable")
	if err != nil {
		log.Fatal("postgres connect failed:", err)
	}
	defer database.Close()
	fmt.Println("postgres connected ✓")

	if err := database.InitSchema(); err != nil {
		log.Fatal("schema init failed:", err)
	}
	fmt.Println("schema ready ✓")

	esClient, err := search.New("http://localhost:9200")
	if err != nil {
		log.Fatal("elasticsearch connect failed:", err)
	}
	fmt.Println("elasticsearch connected ✓")

	if err := esClient.InitIndex(ctx); err != nil {
		log.Fatal("elasticsearch index init failed:", err)
	}
	fmt.Println("elasticsearch index ready ✓")

	q := queue.NewRedisQueue(4, database, esClient)

	results, err := esClient.Search(ctx, "burns")
	if err != nil {
		log.Fatal("search failed:", err)
	}
	for _, r := range results {
		fmt.Printf("%s / %s / %s → %s\n", r.Series, r.Chapter, r.Page, r.Text)
	}
}
