package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"geo-search/internal/models"
)

func ParseQuery(nlpURL, text, city string) (*models.NLPResponse, error) {
	reqBody, err := json.Marshal(map[string]string{
		"text": text,
		"city": city,
	})
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Post(nlpURL+"/parse", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("nlp request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
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
