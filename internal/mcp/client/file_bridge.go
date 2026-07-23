package client

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/alfred/alfred/internal/mcp"
)

// FileBridgeClient communicates with a worker subprocess via HMAC-signed JSON files.
// ponytail: simple file-based IPC, good enough for local workers.
type FileBridgeClient struct {
	dir     string
	secret  []byte
	mu      sync.Mutex
	pending map[string]chan mcp.JSONRPCResponse
	nextID  int
}

// NewFileBridgeClient creates a client using the given directory for file exchange.
func NewFileBridgeClient(dir string, secret []byte) *FileBridgeClient {
	_ = os.MkdirAll(dir, 0700)
	return &FileBridgeClient{
		dir:     dir,
		secret:  secret,
		pending: make(map[string]chan mcp.JSONRPCResponse),
	}
}

// sign signs data with HMAC-SHA256.
func (c *FileBridgeClient) sign(data []byte) string {
	mac := hmac.New(sha256.New, c.secret)
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

// verify verifies an HMAC signature.
func (c *FileBridgeClient) verify(data []byte, sig string) bool {
	return hmac.Equal([]byte(c.sign(data)), []byte(sig))
}

type fileMessage struct {
	ID        string               `json:"id"`
	Direction string               `json:"direction"` // "request" or "response"
	Payload   mcp.JSONRPCRequest   `json:"payload,omitempty"`
	Response  *mcp.JSONRPCResponse `json:"response,omitempty"`
	Signature string               `json:"signature"`
	Timestamp int64                `json:"timestamp"`
}

// Send writes a signed request file and waits for a response file.
func (c *FileBridgeClient) Send(ctx context.Context, req mcp.JSONRPCRequest) (*mcp.JSONRPCResponse, error) {
	c.mu.Lock()
	c.nextID++
	id := fmt.Sprintf("msg-%d-%d", time.Now().UnixNano(), c.nextID)
	ch := make(chan mcp.JSONRPCResponse, 1)
	c.pending[id] = ch
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
	}()

	// Write request file
	msg := fileMessage{
		ID:        id,
		Direction: "request",
		Payload:   req,
		Timestamp: time.Now().UnixNano(),
	}
	b, _ := json.Marshal(msg)
	msg.Signature = c.sign(b)
	b, _ = json.Marshal(msg)

	reqPath := filepath.Join(c.dir, fmt.Sprintf("request-%s.json", id))
	if err := os.WriteFile(reqPath, b, 0600); err != nil {
		return nil, fmt.Errorf("file bridge: write request: %w", err)
	}

	// Wait for response file or timeout
	respPath := filepath.Join(c.dir, fmt.Sprintf("response-%s.json", id))
	deadline := time.After(30 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-deadline:
			return nil, fmt.Errorf("file bridge: timeout waiting for response")
		case <-ticker.C:
			data, err := os.ReadFile(respPath)
			if err != nil {
				continue
			}
			// Verify signature
			var respMsg fileMessage
			if err := json.Unmarshal(data, &respMsg); err != nil {
				continue
			}
			unsigned, _ := json.Marshal(fileMessage{
				ID:        respMsg.ID,
				Direction: respMsg.Direction,
				Response:  respMsg.Response,
				Timestamp: respMsg.Timestamp,
			})
			if !c.verify(unsigned, respMsg.Signature) {
				continue
			}
			_ = os.Remove(respPath)
			_ = os.Remove(reqPath)
			if respMsg.Response != nil {
				return respMsg.Response, nil
			}
			return nil, fmt.Errorf("file bridge: empty response")
		}
	}
}

// Watch starts watching for response files and routes them to pending channels.
// Runs until ctx is cancelled.
func (c *FileBridgeClient) Watch(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	pattern := filepath.Join(c.dir, "response-*.json")
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			matches, _ := filepath.Glob(pattern)
			for _, f := range matches {
				data, err := os.ReadFile(f)
				if err != nil {
					continue
				}
				var msg fileMessage
				if err := json.Unmarshal(data, &msg); err != nil {
					continue
				}
				c.mu.Lock()
				ch, ok := c.pending[msg.ID]
				c.mu.Unlock()
				if ok {
					if msg.Response != nil {
						select {
						case ch <- *msg.Response:
						default:
						}
					}
				}
				_ = os.Remove(f)
			}
		}
	}
}

// Close is a no-op for FileBridgeClient.
func (c *FileBridgeClient) Close() error { return nil }
