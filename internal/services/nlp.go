package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"geo-search/internal/models"
)

var nlpClient = &http.Client{
	Timeout: 2 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	},
}

func CheckNLPHealth(ctx context.Context, nlpURL string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, nlpURL+"/health", nil)
	if err != nil {
		return false
	}

	resp, err := nlpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func ParseQuery(ctx context.Context, nlpURL, text, city string) (*models.NLPResponse, error) {
	reqBody, err := json.Marshal(map[string]string{
		"text": text,
		"city": city,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, nlpURL+"/parse", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := nlpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("nlp request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nlp returned status %d: %s", resp.StatusCode, body)
	}

	var result models.NLPResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
