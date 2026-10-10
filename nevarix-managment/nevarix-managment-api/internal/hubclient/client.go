// Package hubclient talks to nevarix-hub, the system of record (PostgreSQL).
package hubclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/ArminDashti/nevarix-api/internal/metrics"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func New(baseURL, token string) *Client {
	return &Client{
		baseURL:    baseURL,
		token:      token,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// AgentLatest is the Hub's latest snapshot for one agent.
type AgentLatest struct {
	AgentID              string               `json:"agentId"`
	Hostname             string               `json:"hostname"`
	CPUPercent           float64              `json:"cpuPercent"`
	Cores                []metrics.CoreSample `json:"cores"`
	MemoryTotalBytes     uint64               `json:"memoryTotalBytes"`
	MemoryUsedBytes      uint64               `json:"memoryUsedBytes"`
	MemoryAvailableBytes uint64               `json:"memoryAvailableBytes"`
	MemoryUsedPercent    float64              `json:"memoryUsedPercent"`
	Disks                []metrics.DiskSample `json:"disks"`
	BandwidthRxBps       float64              `json:"bandwidthRxBps"`
	BandwidthTxBps       float64              `json:"bandwidthTxBps"`
	UptimeSeconds        uint64               `json:"uptimeSeconds"`
	System               json.RawMessage      `json:"system"`
	CollectedAt          time.Time            `json:"collectedAt"`
}

// MetricPoint is one history sample returned by the Hub.
type MetricPoint struct {
	Timestamp  time.Time `json:"timestamp"`
	CPUPercent float64   `json:"cpuPercent"`
	MemPercent float64   `json:"memPercent"`
	RxBps      float64   `json:"rxBps"`
	TxBps      float64   `json:"txBps"`
}

// PostJSON sends one outbox payload. Any non-2xx response is an error, so the caller keeps the row.
func (c *Client) PostJSON(ctx context.Context, path string, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	_, err = c.do(req)
	return err
}

// GetJSON returns the raw response body for a read path.
func (c *Client) GetJSON(ctx context.Context, path string, query url.Values) ([]byte, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// LatestAgent returns nil when no agent has reported to the Hub yet.
func (c *Client) LatestAgent(ctx context.Context) (*AgentLatest, error) {
	raw, err := c.GetJSON(ctx, "/api/v1/agents/latest", nil)
	if err != nil {
		return nil, err
	}
	var body struct {
		Agent *AgentLatest `json:"agent"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, err
	}
	return body.Agent, nil
}

func (c *Client) AgentHistory(ctx context.Context, agentID string) ([]MetricPoint, error) {
	raw, err := c.GetJSON(ctx, "/api/v1/agents/"+url.PathEscape(agentID)+"/history", nil)
	if err != nil {
		return nil, err
	}
	var body struct {
		Points []MetricPoint `json:"points"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, err
	}
	return body.Points, nil
}

func (c *Client) do(req *http.Request) ([]byte, error) {
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("hub %s %s: status=%d body=%s", req.Method, req.URL.Path, res.StatusCode, string(raw))
	}
	return raw, nil
}
