package search

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func Open(ctx context.Context, baseURL string) (*Client, error) {
	client := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http: &http.Client{Timeout: 3 * time.Second},
	}
	if err := client.Ping(ctx); err != nil {
		return nil, err
	}
	return client, nil
}

func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/_cluster/health", nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("opensearch health returned %s", resp.Status)
	}
	return nil
}


func (c *Client) IndexPhone(ctx context.Context, id string, document []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+"/phone_numbers/_doc/"+id, strings.NewReader(string(document)))
	if err != nil { return err }
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("opensearch index returned %s", resp.Status)
	}
	return nil
}


func (c *Client) BulkIndexPhones(ctx context.Context, payload []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/_bulk", strings.NewReader(string(payload)))
	if err != nil { return err }
	req.Header.Set("Content-Type", "application/x-ndjson")
	resp, err := c.http.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("opensearch bulk returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var result struct { Errors bool `json:"errors"` }
	if err := json.Unmarshal(body, &result); err != nil { return err }
	if result.Errors { return fmt.Errorf("opensearch bulk response contains item errors") }
	return nil
}
