package contract

// AgentCapabilitySettings controls sub-agent fan-out behavior.
type AgentCapabilitySettings struct {
	MaxDelegationDepth int `json:"maxDelegationDepth"`
	MaxSubAgents       int `json:"maxSubAgents"`
}
