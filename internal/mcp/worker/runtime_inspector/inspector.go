package runtimeinspector

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

// NewInspectorWorker creates a runtime-inspector MCP worker.
func NewInspectorWorker() *worker.WorkerServer {
	envTool := &envTool{}
	gitTool := &gitStatusTool{}
	return worker.NewWorkerServer("runtime-inspector", []tool.Tool{envTool, gitTool}).WithHandler(
		func(ctx context.Context, name string, input json.RawMessage, tc *tool.Context) (*tool.Result, error) {
			switch name {
			case "runtime_env":
				return envTool.Execute(ctx, input, tc)
			case "git_status":
				return gitTool.Execute(ctx, input, tc)
			default:
				return tool.FailureMsg(fmt.Sprintf("unknown tool: %s", name)), nil
			}
		},
	)
}

type envTool struct{}

func (e *envTool) Name() string { return "runtime_env" }

func (e *envTool) Description() string {
	return "Report runtime environment: Go version, OS, arch, and configured paths."
}

func (e *envTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type": "object", "properties": {}}`)
}

func (e *envTool) Execute(_ context.Context, _ json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	info := map[string]string{
		"go_version": runtime.Version(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"num_cpu":    fmt.Sprintf("%d", runtime.NumCPU()),
	}
	b, _ := json.Marshal(info)
	return tool.SuccessWith(fmt.Sprintf("Runtime: %s/%s %s", runtime.GOOS, runtime.GOARCH, runtime.Version()), json.RawMessage(b)), nil
}

type gitStatusTool struct{}

func (g *gitStatusTool) Name() string { return "git_status" }
func (g *gitStatusTool) Description() string {
	return "Report git status: branch, modified files, recent commits."
}

func (g *gitStatusTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Git repo path (default: cwd)"}
		}
	}`)
}

func (g *gitStatusTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Path string `json:"path"`
	}
	_ = json.Unmarshal(raw, &in)

	dir := in.Path
	if dir == "" {
		dir = "."
	}

	branch := runCmd(ctx, dir, "git", "rev-parse", "--abbrev-ref", "HEAD")
	status := runCmd(ctx, dir, "git", "status", "--short")
	log := runCmd(ctx, dir, "git", "log", "--oneline", "-5")

	info := map[string]string{
		"branch": strings.TrimSpace(branch),
		"status": strings.TrimSpace(status),
		"recent": strings.TrimSpace(log),
	}
	b, _ := json.Marshal(info)
	return tool.SuccessWith(fmt.Sprintf("Branch: %s", info["branch"]), json.RawMessage(b)), nil
}

func runCmd(ctx context.Context, dir, name string, args ...string) string {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return string(out)
}
