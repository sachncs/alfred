package model

import (
	"context"
	"sync"
)

// StubClient is a deterministic, test-friendly Client implementation.
// It returns a scripted sequence of StreamChunks per request.
//
// StubClient is safe for concurrent use.
type StubClient struct {
	mu       sync.Mutex
	scripts  [][]StreamChunk // round-robin: each Stream call consumes one script
	requests []Request       // observed requests, for assertions
}

// NewStubClient constructs a StubClient with the given scripts.
// Each script is consumed in order; if scripts run out, an empty
// (Done) chunk is returned.
func NewStubClient(scripts ...[]StreamChunk) *StubClient {
	c := &StubClient{scripts: scripts}
	return c
}

// Stream implements Client.
func (c *StubClient) Stream(ctx context.Context, req Request) (<-chan StreamChunk, error) {
	c.mu.Lock()
	if len(c.scripts) == 0 {
		c.mu.Unlock()
		ch := make(chan StreamChunk, 1)
		ch <- StreamChunk{Done: true}
		close(ch)
		return ch, nil
	}
	script := c.scripts[0]
	c.scripts = c.scripts[1:]
	c.requests = append(c.requests, req)
	c.mu.Unlock()

	ch := make(chan StreamChunk, len(script))
	go func() {
		defer close(ch)
		for _, sc := range script {
			select {
			case <-ctx.Done():
				ch <- StreamChunk{Done: true, Error: ctx.Err().Error()}
				return
			case ch <- sc:
			}
		}
	}()
	return ch, nil
}

// Requests returns the captured requests (defensive copy).
func (c *StubClient) Requests() []Request {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Request, len(c.requests))
	copy(out, c.requests)
	return out
}

// ScriptedChatClient is a higher-level StubClient that emits a single
// assistant text turn. Useful for "happy path" tests.
//
// Defaults to 2 scripts so a boot-time smoke test and one runtime caller
// both get a response without manual wiring.
// ponytail: 2 = smoke test + first HTTP turn; bump with ScriptedChatClientN if more consumers.
func ScriptedChatClient(text string) *StubClient {
	return ScriptedChatClientN(text, 2)
}

// ScriptedChatClientN creates a StubClient with n copies of the same script.
// Use when more than the default number of consumers need a response.
func ScriptedChatClientN(text string, n int) *StubClient {
	scripts := make([][]StreamChunk, n)
	for i := range scripts {
		scripts[i] = []StreamChunk{
			{DeltaText: text},
			{Done: true},
		}
	}
	return NewStubClient(scripts...)
}

// ScriptedToolCallClient emits a single tool call.
func ScriptedToolCallClient(id, name, input string) *StubClient {
	return NewStubClient([]StreamChunk{
		{ToolCall: &ToolCall{ID: id, Name: name, Input: []byte(input)}},
		{Done: true},
	})
}
