package agent_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/alfred/alfred/internal/agent"
	"github.com/alfred/alfred/internal/contract"
	"github.com/alfred/alfred/internal/model"
	"github.com/alfred/alfred/internal/tool"
)

// countingTool is a minimal tool that records calls and returns OK.
type countingTool struct {
	name  string
	count int
}

func (c *countingTool) Name() string             { return c.name }
func (c *countingTool) Description() string      { return "counting tool" }
func (c *countingTool) Schema() json.RawMessage  { return json.RawMessage(`{}`) }
func (c *countingTool) Execute(ctx context.Context, input json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	c.count++
	return tool.SuccessWith("ok", map[string]any{"count": c.count}), nil
}

func newThread() *contract.Thread {
	return &contract.Thread{ID: "thr-1", Title: "t", Status: contract.ThreadStatusIdle}
}

func TestChatAgentBasicStreamingTurn(t *testing.T) {
	t.Parallel()
	client := model.ScriptedChatClient("hello world")
	a := agent.NewChatAgent("test", "you are helpful", client)
	turn, err := a.StartTurn(context.Background(), newThread(), contract.UserInput{Text: "hi"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if turn.Status != contract.TurnStatusCompleted {
		t.Fatalf("status: %s", turn.Status)
	}
	if len(turn.Items) != 2 {
		t.Fatalf("expected 2 items (user + assistant), got %d: %+v", len(turn.Items), turn.Items)
	}
	if turn.Items[0].Kind != contract.ItemKindUserMessage {
		t.Fatalf("first item should be user message, got %s", turn.Items[0].Kind)
	}
	if turn.Items[1].Kind != contract.ItemKindAssistantText {
		t.Fatalf("second item should be assistant text, got %s", turn.Items[1].Kind)
	}
	if turn.Items[1].Text == nil || *turn.Items[1].Text != "hello world" {
		t.Fatalf("assistant text wrong: %+v", turn.Items[1].Text)
	}
}

func TestChatAgentDisplayTextOverridesText(t *testing.T) {
	t.Parallel()
	client := model.ScriptedChatClient("ack")
	a := agent.NewChatAgent("t", "", client)
	turn, err := a.StartTurn(context.Background(), newThread(), contract.UserInput{
		Text:        "raw with secrets",
		DisplayText: "clean",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if *turn.Items[0].Text != "clean" {
		t.Fatalf("expected display text override, got %q", *turn.Items[0].Text)
	}
}

func TestChatAgentToolCallDispatch(t *testing.T) {
	t.Parallel()
	client := model.ScriptedToolCallClient("call-1", "counter", "{}")
	a := agent.NewChatAgent("t", "", client)
	ct := &countingTool{name: "counter"}
	a.RegisterTool(ct)
	turn, err := a.StartTurn(context.Background(), newThread(), contract.UserInput{Text: "call the counter"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if ct.count != 1 {
		t.Fatalf("tool not invoked: count=%d", ct.count)
	}
	if len(turn.Items) < 2 {
		t.Fatalf("expected user + tool result items, got %d", len(turn.Items))
	}
	last := turn.Items[len(turn.Items)-1]
	if last.Kind != contract.ItemKindToolResult {
		t.Fatalf("last item should be tool result, got %s", last.Kind)
	}
	if last.ToolCall == nil || last.ToolCall.Name != "counter" {
		t.Fatalf("tool call wrong: %+v", last.ToolCall)
	}
	if last.ToolCall.Error != "" {
		t.Fatalf("unexpected error: %s", last.ToolCall.Error)
	}
}

func TestChatAgentToolNotFound(t *testing.T) {
	t.Parallel()
	client := model.ScriptedToolCallClient("call-1", "ghost", "{}")
	a := agent.NewChatAgent("t", "", client)
	turn, err := a.StartTurn(context.Background(), newThread(), contract.UserInput{Text: "go"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	last := turn.Items[len(turn.Items)-1]
	if !strings.Contains(last.ToolCall.Error, "tool not found") {
		t.Fatalf("expected tool-not-found error, got: %s", last.ToolCall.Error)
	}
}

func TestChatAgentStreamErrorReturnsError(t *testing.T) {
	t.Parallel()
	client := model.NewStubClient(
		[]model.StreamChunk{{Error: "model rate limited"}},
	)
	a := agent.NewChatAgent("t", "", client)
	_, err := a.StartTurn(context.Background(), newThread(), contract.UserInput{Text: "x"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "rate limited") {
		t.Fatalf("err: %v", err)
	}
}

func TestChatAgentNilThreadRejected(t *testing.T) {
	t.Parallel()
	client := model.ScriptedChatClient("x")
	a := agent.NewChatAgent("t", "", client)
	if _, err := a.StartTurn(context.Background(), nil, contract.UserInput{}); err == nil {
		t.Fatalf("expected error on nil thread")
	}
}

func TestChatAgentInterrupt(t *testing.T) {
	t.Parallel()
	client := model.ScriptedChatClient("x")
	a := agent.NewChatAgent("t", "", client)
	a.State().BeginTurn("thr-1", "turn-1")
	if err := a.Interrupt("thr-1"); err != nil {
		t.Fatalf("err: %v", err)
	}
	if a.State().IsBusy() {
		t.Fatalf("expected idle after interrupt")
	}
}

func TestChatAgentInterruptRejectsForeignThread(t *testing.T) {
	t.Parallel()
	client := model.ScriptedChatClient("x")
	a := agent.NewChatAgent("t", "", client)
	a.State().BeginTurn("thr-1", "turn-1")
	if err := a.Interrupt("thr-other"); err == nil {
		t.Fatalf("expected error interrupting different thread")
	}
}

func TestChatAgentResume(t *testing.T) {
	t.Parallel()
	client := model.ScriptedChatClient("x")
	a := agent.NewChatAgent("t", "", client)
	turn, err := a.Resume("thr-1")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if turn.ThreadID != "thr-1" {
		t.Fatalf("thread id: %s", turn.ThreadID)
	}
}

func TestChatAgentResumeEmptyThreadRejected(t *testing.T) {
	t.Parallel()
	client := model.ScriptedChatClient("x")
	a := agent.NewChatAgent("t", "", client)
	if _, err := a.Resume(""); err == nil {
		t.Fatalf("expected error")
	}
}

func TestChatAgentToolsList(t *testing.T) {
	t.Parallel()
	client := model.ScriptedChatClient("x")
	a := agent.NewChatAgent("t", "", client)
	a.RegisterTool(&countingTool{name: "alpha"})
	a.RegisterTool(&countingTool{name: "beta"})
	tools := a.Tools()
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}
}

func TestChatAgentRecordsSystemPrompt(t *testing.T) {
	t.Parallel()
	client := model.ScriptedChatClient("x")
	a := agent.NewChatAgent("t", "be concise", client)
	_, err := a.StartTurn(context.Background(), newThread(), contract.UserInput{Text: "hi"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	reqs := client.Requests()
	if len(reqs) != 1 {
		t.Fatalf("expected 1 request, got %d", len(reqs))
	}
	if reqs[0].SystemPrompt != "be concise" {
		t.Fatalf("system prompt not threaded: %q", reqs[0].SystemPrompt)
	}
}

func TestChatAgentConcurrentRegistrations(t *testing.T) {
	t.Parallel()
	client := model.ScriptedChatClient("x")
	a := agent.NewChatAgent("t", "", client)
	done := make(chan struct{})
	go func() {
		for i := 0; i < 200; i++ {
			a.RegisterTool(&countingTool{name: "t"})
		}
		close(done)
	}()
	for i := 0; i < 200; i++ {
		_ = a.Tools()
	}
	<-done
}

func TestStubClientClosedChannel(t *testing.T) {
	t.Parallel()
	client := model.ScriptedChatClient("hi")
	ch, err := client.Stream(context.Background(), model.Request{Messages: []model.Message{{Role: "user", Content: "x"}}})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	count := 0
	for range ch {
		count++
	}
	if count < 1 {
		t.Fatalf("expected at least 1 chunk, got %d", count)
	}
}

// ensure errors import is used.
var _ = errors.New
