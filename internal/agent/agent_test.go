package agent_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/sachncs/alfred/internal/agent"
	"github.com/sachncs/alfred/internal/contract"
	"github.com/sachncs/alfred/internal/tool"
)

// stubAgent is a minimal Agent implementation for testing the contract.
type stubAgent struct {
	id    string
	tools map[string]tool.Tool
}

func (s *stubAgent) ID() string { return s.id }

func (s *stubAgent) StartTurn(ctx context.Context, thread *contract.Thread, input contract.UserInput) (*contract.Turn, error) {
	return &contract.Turn{
		ID:       contract.TurnID("turn-stub"),
		ThreadID: thread.ID,
		Status:   contract.TurnStatusCompleted,
	}, nil
}

func (s *stubAgent) Interrupt(threadID contract.ThreadID) error {
	if threadID == "" {
		return errors.New("empty thread id")
	}
	return nil
}

func (s *stubAgent) Resume(threadID contract.ThreadID) (*contract.Turn, error) {
	return &contract.Turn{ID: "turn-resumed", ThreadID: threadID, Status: contract.TurnStatusCompleted}, nil
}

func (s *stubAgent) RegisterTool(t tool.Tool) {
	if s.tools == nil {
		s.tools = map[string]tool.Tool{}
	}
	s.tools[t.Name()] = t
}

func (s *stubAgent) Tools() []tool.Tool {
	out := make([]tool.Tool, 0, len(s.tools))
	for _, t := range s.tools {
		out = append(out, t)
	}
	return out
}

// Compile-time interface conformance.
var _ agent.Agent = (*stubAgent)(nil)

func TestInterfaceConformance(t *testing.T) {
	var _ agent.Agent = &stubAgent{id: "stub"}
}

func TestStartTurnReturnsTurn(t *testing.T) {
	a := &stubAgent{id: "test"}
	thread := &contract.Thread{ID: "thr-1", Title: "t"}
	turn, err := a.StartTurn(context.Background(), thread, contract.UserInput{Text: "hi"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if turn.ThreadID != "thr-1" {
		t.Fatalf("thread id mismatch: %s", turn.ThreadID)
	}
	if turn.Status != contract.TurnStatusCompleted {
		t.Fatalf("status: %s", turn.Status)
	}
}

func TestRegisterAndListTools(t *testing.T) {
	a := &stubAgent{id: "t"}
	st := &echoTool{name: "echo"}
	a.RegisterTool(st)
	if got := a.Tools(); len(got) != 1 || got[0].Name() != "echo" {
		t.Fatalf("tools mismatch: %+v", got)
	}
}

func TestInterruptRejectsEmptyID(t *testing.T) {
	a := &stubAgent{id: "t"}
	if err := a.Interrupt(""); err == nil {
		t.Fatalf("expected error on empty thread id")
	}
	if err := a.Interrupt("thr-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgentStateTransitions(t *testing.T) {
	s := agent.NewAgentState()
	if s.IsBusy() {
		t.Fatalf("new state should not be busy")
	}
	s.BeginTurn("thr-1", "turn-1")
	if !s.IsBusy() {
		t.Fatalf("expected busy after BeginTurn")
	}
	s.EndTurn("turn-1", nil)
	if s.IsBusy() {
		t.Fatalf("expected idle after EndTurn")
	}
	snap := s.Snapshot()
	if snap.TurnCount != 1 {
		t.Fatalf("turn count: %d", snap.TurnCount)
	}
}

func TestAgentStateRecordsError(t *testing.T) {
	s := agent.NewAgentState()
	s.BeginTurn("thr", "turn")
	s.EndTurn("turn", errors.New("boom"))
	snap := s.Snapshot()
	if snap.LastError == nil {
		t.Fatalf("expected error to be recorded")
	}
}

func TestAgentStateSnapshotIsIsolated(t *testing.T) {
	s := agent.NewAgentState()
	s.BeginTurn("thr-1", "turn-1")
	snap := s.Snapshot()
	s.EndTurn("turn-1", nil)
	if snap.CurrentThreadID != "thr-1" {
		t.Fatalf("snapshot should not be mutated by later state changes")
	}
}

func TestAgentStateConcurrent(t *testing.T) {
	s := agent.NewAgentState()
	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			s.BeginTurn("thr", "turn")
			s.EndTurn("turn", nil)
		}
		close(done)
	}()
	for i := 0; i < 1000; i++ {
		_ = s.IsBusy()
		_ = s.Snapshot()
	}
	<-done
}

// echoTool is a minimal tool.Tool for RegisterAndListTools.
type echoTool struct{ name string }

func (e *echoTool) Name() string            { return e.name }
func (e *echoTool) Description() string     { return "echo" }
func (e *echoTool) Schema() json.RawMessage { return json.RawMessage(`{}`) }
func (e *echoTool) Execute(ctx context.Context, input json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	return tool.Success(string(input)), nil
}
