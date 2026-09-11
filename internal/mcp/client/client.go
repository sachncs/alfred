// Package client provides MCP client implementations for connecting to
// remote or subprocess-based MCP servers.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/sachncs/alfred/internal/mcp"
)

// Client is a connection to a remote MCP server.
type Client struct {
	serverID  string
	transport mcp.Transport
	tools     []ToolInfo
	mu        sync.Mutex
}

// ToolInfo describes a tool from a remote server.
type ToolInfo struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// NewStdioClient spawns a subprocess and connects via stdio.
// The caller must call Close when done.
func NewStdioClient(ctx context.Context, name string, args ...string) (*Client, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("mcp stdio: stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("mcp stdio: stdout pipe: %w", err)
	}
	cmd.Stderr = io.Discard // ponytail: discard stderr, workers log to stdout

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("mcp stdio: start: %w", err)
	}

	t := mcp.NewStdioTransport(stdout, stdin)
	c := &Client{transport: t}

	// MCP handshake
	if err := c.initialize(ctx); err != nil {
		_ = t.Close()
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("mcp stdio: initialize: %w", err)
	}

	// Fetch tool list
	if err := c.listTools(ctx); err != nil {
		_ = t.Close()
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("mcp stdio: list tools: %w", err)
	}

	c.serverID = name
	return c, nil
}

func (c *Client) initialize(ctx context.Context) error {
	req := mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion":"2024-11-05","clientInfo":{"name":"alfred","version":"0.1.0"}}`),
	}
	return c.call(ctx, req, nil)
}

func (c *Client) listTools(ctx context.Context) error {
	req := mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`2`),
		Method:  "tools/list",
	}
	var result struct {
		Tools []ToolInfo `json:"tools"`
	}
	if err := c.call(ctx, req, &result); err != nil {
		return err
	}
	c.tools = result.Tools
	return nil
}

func (c *Client) call(ctx context.Context, req mcp.JSONRPCRequest, result any) error {
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}
	if err := c.transport.WriteMessage(ctx, b); err != nil {
		return err
	}
	raw, err := c.transport.ReadMessage(ctx)
	if err != nil {
		return err
	}
	var resp mcp.JSONRPCResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return err
	}
	if resp.Error != nil {
		return fmt.Errorf("mcp error %d: %s", resp.Error.Code, resp.Error.Message)
	}
	if result != nil {
		b, _ := json.Marshal(resp.Result)
		return json.Unmarshal(b, result)
	}
	return nil
}

// Tools returns the tool catalog fetched during initialization.
func (c *Client) Tools() []ToolInfo {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]ToolInfo, len(c.tools))
	copy(out, c.tools)
	return out
}

// CallTool invokes a tool on the remote server.
func (c *Client) CallTool(ctx context.Context, name string, input json.RawMessage) (json.RawMessage, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	params, _ := json.Marshal(map[string]any{
		"name":      name,
		"arguments": input,
	})
	req := mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`"call"`),
		Method:  "tools/call",
		Params:  params,
	}
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	if err := c.transport.WriteMessage(ctx, b); err != nil {
		return nil, err
	}
	raw, err := c.transport.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}
	var resp mcp.JSONRPCResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("mcp error %d: %s", resp.Error.Code, resp.Error.Message)
	}
	return json.Marshal(resp.Result)
}

// Close shuts down the client transport.
func (c *Client) Close() error {
	return c.transport.Close()
}

// ServerID returns the name of the connected server.
func (c *Client) ServerID() string { return c.serverID }
