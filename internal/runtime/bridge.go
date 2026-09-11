package runtime

import (
	"context"

	"github.com/sachncs/alfred/internal/agent"
	"github.com/sachncs/alfred/internal/contract"
)

// AgentBridge connects a ChatAgent to the runtime's turn loop.
type AgentBridge struct {
	agent *agent.ChatAgent
}

// NewAgentBridge creates a bridge between the agent and the runtime.
func NewAgentBridge(a *agent.ChatAgent) *AgentBridge {
	return &AgentBridge{agent: a}
}

// RunTurn delegates to the ChatAgent's StartTurn method.
func (b *AgentBridge) RunTurn(ctx context.Context, thread *contract.Thread, input contract.UserInput) (*contract.Turn, error) {
	return b.agent.StartTurn(ctx, thread, input)
}

// Agent returns the underlying ChatAgent.
func (b *AgentBridge) Agent() *agent.ChatAgent {
	return b.agent
}
