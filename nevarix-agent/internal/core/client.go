package core

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

// Health checks GET {core}/api/v1/health.
func Health(ctx context.Context, coreURL string) error {
	_, err := do(ctx, http.MethodGet, coreURL+"/api/v1/health", nil)
	return err
}

// Agents returns the raw GET {core}/api/v1/agents JSON.
func Agents(ctx context.Context, coreURL string) ([]byte, error) {
	return do(ctx, http.MethodGet, coreURL+"/api/v1/agents", nil)
}

// Scan asks Core to scan its default CIDR (POST {core}/api/v1/scan with empty body).
func Scan(ctx context.Context, coreURL string) ([]byte, error) {
	return do(ctx, http.MethodPost, coreURL+"/api/v1/scan", bytes.NewReader([]byte("{}")))
}

func do(ctx context.Context, method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return raw, fmt.Errorf("core returned status=%d", resp.StatusCode)
	}
	return raw, nil
}
