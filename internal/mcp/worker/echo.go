package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alfred/alfred/internal/tool"
)

// EchoInput is the input schema for the EchoServer's echo tool.
type EchoInput struct {
	Message string `json:"message"`
}

// EchoServer is the hello-world MCP worker. It exposes one tool, `echo`,
// that returns whatever message the client sends.
type EchoServer struct {
	*WorkerServer
}

// NewEchoServer constructs an EchoServer.
func NewEchoServer() *EchoServer {
	echo := newEchoTool()
	return &EchoServer{
		WorkerServer: NewWorkerServer("echo-worker", []tool.Tool{echo}),
	}
}

// echoTool implements the echo behavior. Defined as a separate type so
// it's easy to swap out in tests.
type echoTool struct{}

func newEchoTool() *echoTool { return &echoTool{} }

func (e *echoTool) Name() string { return "echo" }
func (e *echoTool) Description() string {
	return "Echoes back the provided message. Useful for smoke testing the MCP transport."
}
func (e *echoTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"message":{"type":"string"}},"required":["message"]}`)
}
func (e *echoTool) Execute(ctx context.Context, input json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in EchoInput
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
		}
	}
	if in.Message == "" {
		return tool.FailureMsg("message is required"), nil
	}
	return tool.SuccessWith("echo: "+in.Message, map[string]any{"echo": in.Message}), nil
}
