package contract

// AgentRuntimeId identifies the active agent runtime.
type AgentRuntimeId string

const (
	AgentRuntimeAlfred   AgentRuntimeId = "alfred"
	AgentRuntimeOpencode AgentRuntimeId = "opencode"
)

// NormalizeAgentRuntimeId returns a valid AgentRuntimeId, defaulting to Alfred.
func NormalizeAgentRuntimeId(v string) AgentRuntimeId {
	switch AgentRuntimeId(v) {
	case AgentRuntimeAlfred, AgentRuntimeOpencode:
		return AgentRuntimeId(v)
	default:
		return AgentRuntimeAlfred
	}
}
