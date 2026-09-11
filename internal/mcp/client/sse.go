package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/sachncs/alfred/internal/mcp"
)

// SSEClient connects to an MCP server over SSE + HTTP POST.
type SSEClient struct {
	endpoint  string
	sessionID string
	client    *http.Client
	notify    chan struct{}
	ready     bool
	mu        sync.Mutex
}

// NewSSEClient creates an SSEClient. The endpoint is the SSE URL (e.g. /sse).
func NewSSEClient(endpoint string) *SSEClient {
	return &SSEClient{
		endpoint: endpoint,
		client:   &http.Client{},
		notify:   make(chan struct{}, 1),
	}
}

// Run connects to the SSE endpoint and listens for the session endpoint.
// Blocks until ctx is cancelled. Call Ready() to wait for the session.
func (c *SSEClient) Run(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "event: endpoint") {
			// Next data line is the POST endpoint
			if scanner.Scan() {
				data := strings.TrimPrefix(scanner.Text(), "data: ")
				c.mu.Lock()
				c.sessionID = data
				c.ready = true
				c.mu.Unlock()
				select {
				case c.notify <- struct{}{}:
				default:
				}
			}
		}
		// ponytail: ignore other SSE events (responses come via POST response)
	}
	return scanner.Err()
}

// Ready blocks until the SSE session endpoint is available.
func (c *SSEClient) Ready(ctx context.Context) error {
	if c.IsReady() {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.notify:
		return nil
	}
}

// IsReady reports whether the SSE session is established.
func (c *SSEClient) IsReady() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ready
}

// Send sends a JSON-RPC request via HTTP POST and returns the response.
func (c *SSEClient) Send(ctx context.Context, req mcp.JSONRPCRequest) (*mcp.JSONRPCResponse, error) {
	c.mu.Lock()
	postURL := c.sessionID
	c.mu.Unlock()

	if postURL == "" {
		return nil, fmt.Errorf("sse: no session endpoint")
	}

	// Resolve relative URLs against the SSE endpoint
	if !strings.HasPrefix(postURL, "http") {
		u, err := url.Parse(c.endpoint)
		if err != nil {
			return nil, err
		}
		u.Path = postURL
		postURL = u.String()
	}

	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", postURL, strings.NewReader(string(b)))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.sessionID != "" {
		httpReq.Header.Set("Mcp-Session-Id", c.sessionID)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result mcp.JSONRPCResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("sse: unmarshal response: %w", err)
	}
	return &result, nil
}
