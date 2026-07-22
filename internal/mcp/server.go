// Package mcp defines the Model Context Protocol server abstraction.
// Phase 1 ships a stdio transport; HTTP/StreamableHTTP/file-bridge
// transports land in Phase 5 alongside the real workers.
package mcp

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/alfred/alfred/internal/tool"
)

// Server is the base interface every concrete MCP server must satisfy.
type Server interface {
	// ID returns the stable identifier used in MCP `initialize` responses.
	ID() string

	// Tools returns the catalog of tools this server exposes. The list
	// is captured at server-start time; tools added after Start are
	// ignored until the next restart.
	Tools() []tool.Tool

	// Start begins serving on the given transport. Start MUST block
	// until either the transport closes, ctx is cancelled, or an
	// unrecoverable error occurs. Returning nil means graceful shutdown.
	Start(ctx context.Context, transport Transport) error
}

// Transport abstracts the wire protocol over which MCP messages flow.
type Transport interface {
	// ReadMessage returns the next JSON-RPC message from the transport.
	// It MUST return io.EOF when the peer closes cleanly.
	ReadMessage(ctx context.Context) (json.RawMessage, error)

	// WriteMessage sends a JSON-RPC message to the peer.
	WriteMessage(ctx context.Context, msg json.RawMessage) error

	// Close releases transport resources.
	Close() error
}

// JSONRPCRequest is the wire schema for an MCP `initialize` / tool call.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse is the wire schema for an MCP response.
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError is the wire schema for an MCP error response.
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Standard JSON-RPC error codes.
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInternalError  = -32603
)

// ErrTransportClosed is returned by Transport.ReadMessage when the peer
// has closed the connection cleanly.
var ErrTransportClosed = errors.New("mcp: transport closed")
