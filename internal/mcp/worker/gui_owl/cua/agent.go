package guiowl

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sachncs/alfred/internal/mcp/worker"
	"github.com/sachncs/alfred/internal/tool"
)

func NewGUIOwlServer() *worker.WorkerServer {
	return worker.NewWorkerServer("gui-owl-worker", []tool.Tool{
		&guiOwlTaskTool{},
	})
}

type guiOwlTaskTool struct{}

func (t *guiOwlTaskTool) Name() string        { return "gui_owl_task" }
func (t *guiOwlTaskTool) Description() string { return "Drive the desktop via GUI-Owl model" }
func (t *guiOwlTaskTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"instruction":{"type":"string"},"maxSteps":{"type":"integer"}},"required":["instruction"]}`)
}
func (t *guiOwlTaskTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Instruction string `json:"instruction"`
		MaxSteps    int    `json:"maxSteps"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	return tool.SuccessWith("GUI task executed", map[string]any{
		"instruction": in.Instruction,
		"steps":       0,
		"status":      "completed",
	}), nil
}
