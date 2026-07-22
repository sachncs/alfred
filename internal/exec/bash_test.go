package exec

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alfred/alfred/internal/tool"
)

func TestBashToolEcho(t *testing.T) {
	bt := NewBashTool()
	input, _ := json.Marshal(BashInput{Command: "echo hello"})
	res, err := bt.Execute(context.Background(), input, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Fatalf("bash failed: %s", res.Error)
	}
	if res.Content != "hello\n" {
		t.Errorf("content = %q, want 'hello\\n'", res.Content)
	}
}

func TestBashToolExitCode(t *testing.T) {
	bt := NewBashTool()
	input, _ := json.Marshal(BashInput{Command: "exit 42"})
	res, err := bt.Execute(context.Background(), input, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.OK {
		t.Error("expected failure for exit 42")
	}
}

func TestBashToolStderr(t *testing.T) {
	bt := NewBashTool()
	input, _ := json.Marshal(BashInput{Command: "echo err >&2; exit 1"})
	res, err := bt.Execute(context.Background(), input, nil)
	if err != nil {
		t.Fatal(err)
	}
	var out BashOutput
	json.Unmarshal(res.Structured, &out)
	if out.Stderr == "" {
		t.Error("expected stderr output")
	}
	if out.ExitCode != 1 {
		t.Errorf("exit code = %d, want 1", out.ExitCode)
	}
}

func TestBashToolTimeout(t *testing.T) {
	bt := NewBashTool()
	input, _ := json.Marshal(BashInput{Command: "sleep 10", Timeout: 1})
	res, err := bt.Execute(context.Background(), input, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.OK {
		t.Error("expected failure for timeout")
	}
}

func TestBashToolContextCancellation(t *testing.T) {
	bt := NewBashTool()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	input, _ := json.Marshal(BashInput{Command: "sleep 10"})
	res, err := bt.Execute(ctx, input, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.OK {
		t.Error("expected failure for context cancellation")
	}
}

func TestBashToolEmptyCommand(t *testing.T) {
	bt := NewBashTool()
	input, _ := json.Marshal(BashInput{Command: ""})
	res, err := bt.Execute(context.Background(), input, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.OK {
		t.Error("expected failure for empty command")
	}
}

func TestBashToolWorkspaceDir(t *testing.T) {
	bt := NewBashTool()
	tc := tool.NewContext(context.Background(), "", "", "", "/tmp")
	input, _ := json.Marshal(BashInput{Command: "pwd"})
	res, err := bt.Execute(context.Background(), input, tc)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Fatalf("bash failed: %s", res.Error)
	}
	// Output should be /tmp or similar.
	if res.Content == "" {
		t.Error("expected pwd output")
	}
}
