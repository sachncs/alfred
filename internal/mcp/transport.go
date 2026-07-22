package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// StdioTransport implements Transport over an in-process pair of
// io.Reader / io.Writer. Useful for tests and for embedding an MCP
// server inside another process.
type StdioTransport struct {
	reader io.Reader
	writer io.Writer

	mu      sync.Mutex
	closed  bool
	scanner *bufio.Scanner
}

// NewStdioTransport constructs a StdioTransport reading from r and
// writing to w. The scanner is configured for long lines (10 MiB).
func NewStdioTransport(r io.Reader, w io.Writer) *StdioTransport {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 10*1024*1024)
	return &StdioTransport{reader: r, writer: w, scanner: sc}
}

// ReadMessage implements Transport. Sequential — concurrent calls
// from multiple goroutines are not safe; the runtime uses one reader
// goroutine per transport.
func (t *StdioTransport) ReadMessage(ctx context.Context) (json.RawMessage, error) {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return nil, ErrTransportClosed
	}
	sc := t.scanner
	t.mu.Unlock()

	// Use a goroutine so the scanner can be cancelled via context.
	type result struct {
		line []byte
		err  error
	}
	done := make(chan result, 1)
	go func() {
		if !sc.Scan() {
			if err := sc.Err(); err != nil {
				done <- result{nil, err}
				return
			}
			done <- result{nil, io.EOF}
			return
		}
		done <- result{append([]byte(nil), sc.Bytes()...), nil}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-done:
		if r.err != nil {
			if r.err == io.EOF {
				return nil, ErrTransportClosed
			}
			return nil, r.err
		}
		return json.RawMessage(r.line), nil
	}
}

// WriteMessage implements Transport.
func (t *StdioTransport) WriteMessage(ctx context.Context, msg json.RawMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return ErrTransportClosed
	}
	if _, err := fmt.Fprintln(t.writer, string(msg)); err != nil {
		return fmt.Errorf("stdio write: %w", err)
	}
	return nil
}

// Close implements Transport.
func (t *StdioTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
	return nil
}
