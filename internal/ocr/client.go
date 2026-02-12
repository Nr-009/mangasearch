package ocr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type request struct {
	Path string `json:"path"`
}

type response struct {
	Text string `json:"text"`
}

func GetData(pathName string) (string, error) {
	data, err := json.Marshal(request{Path: pathName})
	if err != nil {
		return "", err
	}

	resp, err := http.Post("http://127.0.0.1:5000/ocr", "application/json", bytes.NewBuffer(data))
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
