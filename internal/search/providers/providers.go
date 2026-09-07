// Package providers implements search providers for the research service.
package providers

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Result is a unified search result across all providers.
type Result struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
	Source  string `json:"source"` // "duckduckgo", "arxiv", etc.
	Date    string `json:"date,omitempty"`
	Authors string `json:"authors,omitempty"`
}

// Provider is a search provider interface.
type Provider interface {
	Name() string
	Search(ctx context.Context, query string, limit int) ([]Result, error)
}

var httpClient = &http.Client{Timeout: 15 * time.Second}

func doGet(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Alfred/0.4 (research assistant)")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var buf [64 * 1024]byte
	n, _ := resp.Body.Read(buf[:])
	return buf[:n], nil
}

func doPost(ctx context.Context, url, contentType string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", "Alfred/0.4 (research assistant)")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var buf [64 * 1024]byte
	n, _ := resp.Body.Read(buf[:])
	return buf[:n], nil
}

func encodeQuery(q string) string {
	return url.QueryEscape(q)
}
