package settings

import "github.com/alfred/alfred/internal/contract"

// ModelAccessRuntimePolicy is the resolved model access policy.
type ModelAccessRuntimePolicy struct {
	AllowedModels []string
	MaxTokens     int
	SandboxMode   contract.SandboxMode
}

// ResolveModelAccessRuntimePolicy computes the effective model access policy.
func ResolveModelAccessRuntimePolicy(s contract.AppSettingsV1) ModelAccessRuntimePolicy {
	return ModelAccessRuntimePolicy{
		AllowedModels: s.ModelAccess.AllowedModels,
		MaxTokens:     s.ModelAccess.MaxTokens,
		SandboxMode:   s.LocalRuntime.SandboxMode,
	}
}

// IsModelAllowed checks if a model is in the allowed list.
func (p ModelAccessRuntimePolicy) IsModelAllowed(model string) bool {
	if len(p.AllowedModels) == 0 {
		return true
	}
	for _, m := range p.AllowedModels {
		if m == model {
			return true
		}
	}
	return false
}
