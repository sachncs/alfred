package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/sachncs/alfred/internal/mcp"
)

// StreamableHTTPClient connects to an MCP server over bidirectional HTTP streaming.
type StreamableHTTPClient struct {
	endpoint  string
	sessionID string
	client    *http.Client
	mu        sync.Mutex
}

// NewStreamableHTTPClient creates a client for the given endpoint.
func NewStreamableHTTPClient(endpoint string) *StreamableHTTPClient {
	return &StreamableHTTPClient{
		endpoint: endpoint,
		client:   &http.Client{},
	}
}

// Send sends a JSON-RPC request and reads the streaming response.
func (c *StreamableHTTPClient) Send(ctx context.Context, req mcp.JSONRPCRequest) (*mcp.JSONRPCResponse, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream, application/json")
	c.mu.Lock()
	if c.sessionID != "" {
		httpReq.Header.Set("Mcp-Session-Id", c.sessionID)
	}
	c.mu.Unlock()

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	// Capture session ID from response header
	if sid := resp.Header.Get("Mcp-Session-Id"); sid != "" {
		c.mu.Lock()
		c.sessionID = sid
		c.mu.Unlock()
	}

	ct := resp.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "application/json") {
		body, _ := io.ReadAll(resp.Body)
		var result mcp.JSONRPCResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}

	// SSE stream — read until we get a message event
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			var result mcp.JSONRPCResponse
			if err := json.Unmarshal([]byte(data), &result); err != nil {
				continue
			}
			return &result, nil
		}
	}
	return nil, fmt.Errorf("streamable http: no response received")
}

// Close is a no-op for StreamableHTTPClient.
func (c *StreamableHTTPClient) Close() error { return nil }
