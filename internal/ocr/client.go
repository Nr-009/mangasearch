package ocr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	url             string
	macPrefix       string
	containerPrefix string
}

func NewClient(port int, macPrefix, containerPrefix string) *Client {
	return &Client{
		url:             fmt.Sprintf("http://127.0.0.1:%d/ocr", port),
		macPrefix:       macPrefix,
		containerPrefix: containerPrefix,
	}
}

type request struct {
	Path string `json:"path"`
}

type response struct {
	Text string `json:"text"`
}

func (c *Client) translatePath(path string) string {
	return strings.Replace(path, c.macPrefix, c.containerPrefix, 1)
}

func (c *Client) GetData(pathName string) (string, error) {
	containerPath := c.translatePath(pathName)

	data, err := json.Marshal(request{Path: containerPath})
	if err != nil {
		return "", err
	}

	resp, err := http.Post(c.url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result response
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Text, nil
}
