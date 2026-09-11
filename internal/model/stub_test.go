package model_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/sachncs/alfred/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStubClientScriptedText(t *testing.T) {
	t.Parallel()
	client := model.ScriptedChatClient("hello world")
	ch, err := client.Stream(context.Background(), model.Request{Messages: []model.Message{{Role: "user", Content: "hi"}}})
	require.NoError(t, err)
	defer func() {
		// Drain channel so the goroutine can exit cleanly.
		for range ch {
		}
	}()
	var got string
	var done bool
	deadline := time.After(2 * time.Second)
	for !done {
		select {
		case chunk, ok := <-ch:
			require.True(t, ok, "channel closed prematurely")
			if chunk.DeltaText != "" {
				got += chunk.DeltaText
			}
			if chunk.Done {
				done = true
			}
		case <-deadline:
			t.Fatal("timeout")
		}
	}
	assert.Equal(t, "hello world", got)
}

func TestStubClientScriptedToolCall(t *testing.T) {
	t.Parallel()
	client := model.ScriptedToolCallClient("call-1", "echo", `{"msg":"hi"}`)
	ch, err := client.Stream(context.Background(), model.Request{})
	require.NoError(t, err)
	var tc *model.ToolCall
	for chunk := range ch {
		if chunk.ToolCall != nil {
			tc = chunk.ToolCall
		}
		if chunk.Done {
			break
		}
	}
	require.NotNil(t, tc)
	assert.Equal(t, "call-1", tc.ID)
	assert.Equal(t, "echo", tc.Name)
	assert.JSONEq(t, `{"msg":"hi"}`, string(tc.Input))
}

func TestStubClientEmptyScriptsEmitsDone(t *testing.T) {
	t.Parallel()
	client := model.NewStubClient()
	ch, err := client.Stream(context.Background(), model.Request{})
	require.NoError(t, err)
	var done bool
	for chunk := range ch {
		if chunk.Done {
			done = true
		}
	}
	assert.True(t, done, "empty scripts should still emit a Done chunk")
}

func TestStubClientRecordsRequest(t *testing.T) {
	t.Parallel()
	client := model.ScriptedChatClient("ok")
	req := model.Request{
		SystemPrompt: "be terse",
		Messages:     []model.Message{{Role: "user", Content: "ping"}},
		Model:        "test-model",
	}
	ch, err := client.Stream(context.Background(), req)
	require.NoError(t, err)
	for range ch {
	}
	reqs := client.Requests()
	require.Len(t, reqs, 1)
	assert.Equal(t, "be terse", reqs[0].SystemPrompt)
	assert.Equal(t, "test-model", reqs[0].Model)
}

// ponytail: ScriptedChatClient defaults to 2 scripts so a smoke test and one
// runtime caller both get a response. If this breaks, the "boot smoke + HTTP
// turn gets an empty stream" bug returns.
func TestScriptedChatClientDefaultsToTwoScripts(t *testing.T) {
	t.Parallel()
	client := model.ScriptedChatClient("ok")

	for i := 0; i < 2; i++ {
		ch, err := client.Stream(context.Background(), model.Request{})
		require.NoError(t, err)
		var got string
		var done bool
		for chunk := range ch {
			if chunk.DeltaText != "" {
				got += chunk.DeltaText
			}
			if chunk.Done {
				done = true
			}
		}
		require.True(t, done, "call %d: stream never finished", i)
		assert.Equal(t, "ok", got, "call %d: text mismatch", i)
	}

	// Third call should fall back to the empty-script Done chunk.
	ch, err := client.Stream(context.Background(), model.Request{})
	require.NoError(t, err)
	for chunk := range ch {
		require.True(t, chunk.Done, "third call should emit Done")
	}
}

func TestScriptedChatClientN(t *testing.T) {
	t.Parallel()
	client := model.ScriptedChatClientN("ok", 5)
	for i := 0; i < 5; i++ {
		ch, err := client.Stream(context.Background(), model.Request{})
		require.NoError(t, err)
		var got string
		for chunk := range ch {
			if chunk.DeltaText != "" {
				got += chunk.DeltaText
			}
			if chunk.Done {
				break
			}
		}
		assert.Equal(t, "ok", got, "call %d", i)
	}
}

func TestRequestJSONRoundTrip(t *testing.T) {
	t.Parallel()
	req := model.Request{
		SystemPrompt: "you are helpful",
		Messages: []model.Message{
			{Role: "system", Content: "sys"},
			{Role: "user", Content: "hi"},
			{Role: "assistant", Content: "hello"},
			{Role: "tool", Content: "out", Name: "echo"},
		},
		Model: "x",
	}
	b, err := json.Marshal(req)
	require.NoError(t, err)
	var out model.Request
	require.NoError(t, json.Unmarshal(b, &out))
	assert.Equal(t, req.SystemPrompt, out.SystemPrompt)
	assert.Equal(t, req.Model, out.Model)
	assert.Len(t, out.Messages, 4)
	assert.Equal(t, "tool", out.Messages[3].Role)
	assert.Equal(t, "echo", out.Messages[3].Name)
}

func TestStreamChunkJSONRoundTrip(t *testing.T) {
	t.Parallel()
	c := model.StreamChunk{
		DeltaText: "hi",
		ToolCall:  &model.ToolCall{ID: "t1", Name: "x", Input: []byte(`{}`)},
		Done:      true,
	}
	b, err := json.Marshal(c)
	require.NoError(t, err)
	var out model.StreamChunk
	require.NoError(t, json.Unmarshal(b, &out))
	assert.Equal(t, "hi", out.DeltaText)
	assert.True(t, out.Done)
	require.NotNil(t, out.ToolCall)
	assert.Equal(t, "t1", out.ToolCall.ID)
}
