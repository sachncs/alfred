// Package exec provides the BashTool for running shell commands.
package exec

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/alfred/alfred/internal/tool"
)

// BashInput is the JSON schema for BashTool input.
type BashInput struct {
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"` // seconds, default 30
}

// BashOutput is the structured payload returned by BashTool.
type BashOutput struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
}

// BashTool runs shell commands with timeout and captures output.
type BashTool struct{}

// NewBashTool returns a fresh BashTool.
func NewBashTool() *BashTool { return &BashTool{} }

// Name implements tool.Tool.
func (*BashTool) Name() string { return "exec_bash" }

// Description implements tool.Tool.
func (*BashTool) Description() string {
	return "Execute a shell command with timeout. Returns stdout, stderr, and exit code."
}

// Schema implements tool.Tool.
func (*BashTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"command": {"type": "string", "description": "Shell command to execute"},
			"timeout": {"type": "integer", "minimum": 1, "description": "Timeout in seconds (default 30)"}
		},
		"required": ["command"]
	}`)
}

// Execute implements tool.Tool.
func (t *BashTool) Execute(ctx context.Context, raw json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in BashInput
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Command == "" {
		return tool.FailureMsg("command is required"), nil
	}

	timeout := 30
	if in.Timeout > 0 {
		timeout = in.Timeout
	}

	// Apply timeout to context.
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", in.Command)

	// Set working directory to workspace root if available.
	if tc != nil && tc.WorkspaceRoot != "" {
		cmd.Dir = tc.WorkspaceRoot
	}

	stdout, err := cmd.Output()
	stderr := ""
	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr = string(exitErr.Stderr)
		return &tool.Result{
			OK:      false,
			Content: fmt.Sprintf("command failed with exit code %d", exitErr.ExitCode()),
			Structured: mustMarshal(BashOutput{
				Stdout:   string(stdout),
				Stderr:   stderr,
				ExitCode: exitErr.ExitCode(),
			}),
		}, nil
	}
	if err != nil {
		return tool.Failure(err), nil
	}

	out := BashOutput{
		Stdout:   string(stdout),
		ExitCode: 0,
	}
	return tool.SuccessWith(string(stdout), out), nil
}

func mustMarshal(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
}

func init() {
	_ = os.Stdout // ensure os is imported
}
