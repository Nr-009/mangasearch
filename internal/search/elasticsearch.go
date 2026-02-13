package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
)

const indexName = "manga_pages"

const mapping = `{
  "mappings": {
    "properties": {
      "series":  { "type": "keyword" },
      "chapter": { "type": "keyword" },
      "page":    { "type": "keyword" },
      "path":    { "type": "keyword" },
      "text":    { "type": "text" }
    }
  }
}`

type Client struct {
	es *elasticsearch.Client
}

func New(address string) (*Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{address},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch.New: %w", err)
	}
	return &Client{es: es}, nil
}

func (c *Client) InitIndex(ctx context.Context) error {
	res, err := c.es.Indices.Exists([]string{indexName})
	if err != nil {
		return fmt.Errorf("InitIndex exists check: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		return nil
	}

	res, err = c.es.Indices.Create(
		indexName,
		c.es.Indices.Create.WithBody(bytes.NewReader([]byte(mapping))),
		c.es.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("InitIndex create: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("InitIndex create response: %s", res.String())
	}
	return nil
}

func (c *Client) IndexPage(ctx context.Context, series, chapter, page, path, text string) error {
	doc := map[string]string{
		"series":  series,
		"chapter": chapter,
		"page":    page,
		"path":    path,
		"text":    text,
	}

	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("IndexPage marshal: %w", err)
	}

	res, err := c.es.Index(
		indexName,
		bytes.NewReader(body),
		c.es.Index.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("IndexPage index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("IndexPage response: %s", res.String())
	}
	return nil
}

type SearchResult struct {
	Series  string `json:"series"`
	Chapter string `json:"chapter"`
	Page    string `json:"page"`
	Path    string `json:"path"`
	Text    string `json:"text"`
}

func (c *Client) Search(ctx context.Context, query string) ([]SearchResult, error) {
	body, err := json.Marshal(map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"text": query,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("Search marshal: %w", err)
	}

	res, err := c.es.Search(
		c.es.Search.WithContext(ctx),
		c.es.Search.WithIndex(indexName),
		c.es.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, fmt.Errorf("Search request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Search response: %s", res.String())
	}

	var response struct {
		Hits struct {
			Hits []struct {
				Source SearchResult `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("Search decode: %w", err)
	}

	results := make([]SearchResult, 0, len(response.Hits.Hits))
	for _, hit := range response.Hits.Hits {
		results = append(results, hit.Source)
	}

	return results, nil
}
