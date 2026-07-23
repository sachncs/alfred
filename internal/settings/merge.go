package settings

import "github.com/alfred/alfred/internal/contract"

// Merge deep-merges a partial AppSettingsV1 into an existing one.
// Non-zero fields in partial overwrite base; zero fields are kept from base.
func Merge(base, partial contract.AppSettingsV1) contract.AppSettingsV1 {
	if partial.Version != "" {
		base.Version = partial.Version
	}

	// Agents
	if partial.Agents.ActiveRuntime != "" {
		base.Agents.ActiveRuntime = partial.Agents.ActiveRuntime
	}
	if partial.Agents.OpenCode.Command != "" {
		base.Agents.OpenCode = partial.Agents.OpenCode
	}

	// LocalRuntime
	if partial.LocalRuntime.BinaryPath != "" {
		base.LocalRuntime.BinaryPath = partial.LocalRuntime.BinaryPath
	}
	if partial.LocalRuntime.Port != 0 {
		base.LocalRuntime.Port = partial.LocalRuntime.Port
	}
	if partial.LocalRuntime.Model != "" {
		base.LocalRuntime.Model = partial.LocalRuntime.Model
	}
	if partial.LocalRuntime.ApprovalPolicy != "" {
		base.LocalRuntime.ApprovalPolicy = partial.LocalRuntime.ApprovalPolicy
	}
	if partial.LocalRuntime.SandboxMode != "" {
		base.LocalRuntime.SandboxMode = partial.LocalRuntime.SandboxMode
	}
	if partial.LocalRuntime.RuntimeToken != "" {
		base.LocalRuntime.RuntimeToken = partial.LocalRuntime.RuntimeToken
	}
	if partial.LocalRuntime.DataDir != "" {
		base.LocalRuntime.DataDir = partial.LocalRuntime.DataDir
	}

	// ModelRouter
	if partial.ModelRouter.DefaultModel != "" {
		base.ModelRouter.DefaultModel = partial.ModelRouter.DefaultModel
	}
	if len(partial.ModelRouter.RoutingRules) > 0 {
		if base.ModelRouter.RoutingRules == nil {
			base.ModelRouter.RoutingRules = make(map[string]string)
		}
		for k, v := range partial.ModelRouter.RoutingRules {
			base.ModelRouter.RoutingRules[k] = v
		}
	}

	// ModelAccess
	if len(partial.ModelAccess.AllowedModels) > 0 {
		base.ModelAccess.AllowedModels = partial.ModelAccess.AllowedModels
	}
	if partial.ModelAccess.MaxTokens != 0 {
		base.ModelAccess.MaxTokens = partial.ModelAccess.MaxTokens
	}

	// AgentCapabilities
	if partial.AgentCapabilities.MaxDelegationDepth != 0 {
		base.AgentCapabilities.MaxDelegationDepth = partial.AgentCapabilities.MaxDelegationDepth
	}
	if partial.AgentCapabilities.MaxSubAgents != 0 {
		base.AgentCapabilities.MaxSubAgents = partial.AgentCapabilities.MaxSubAgents
	}

	// KeyboardShortcuts
	if len(partial.KeyboardShortcuts.Shortcuts) > 0 {
		base.KeyboardShortcuts = partial.KeyboardShortcuts
	}

	// GUIUpdate
	if partial.GUIUpdate.Channel != "" {
		base.GUIUpdate.Channel = partial.GUIUpdate.Channel
	}

	// AppBehavior
	if partial.AppBehavior.Theme != "" {
		base.AppBehavior.Theme = partial.AppBehavior.Theme
	}
	if partial.AppBehavior.Language != "" {
		base.AppBehavior.Language = partial.AppBehavior.Language
	}
	if partial.AppBehavior.UIScale != 0 {
		base.AppBehavior.UIScale = partial.AppBehavior.UIScale
	}
	if partial.AppBehavior.FontScale != 0 {
		base.AppBehavior.FontScale = partial.AppBehavior.FontScale
	}

	// Boolean fields: always take partial's value (zero = explicitly set to false)
	base.LocalRuntime.AutoStart = partial.LocalRuntime.AutoStart
	base.LocalRuntime.TokenEconomy = partial.LocalRuntime.TokenEconomy
	base.LocalRuntime.Insecure = partial.LocalRuntime.Insecure
	base.Agents.OpenCode.AutoStart = partial.Agents.OpenCode.AutoStart
	base.AppBehavior.AutoStart = partial.AppBehavior.AutoStart
	base.ComputerUse.Enabled = partial.ComputerUse.Enabled
	base.RemoteChannel.Enabled = partial.RemoteChannel.Enabled
	base.RemoteExecutor.Enabled = partial.RemoteExecutor.Enabled
	base.Schedule.Enabled = partial.Schedule.Enabled
	base.Workflow.Enabled = partial.Workflow.Enabled
	base.Write.Enabled = partial.Write.Enabled
	base.SpeechToText.Enabled = partial.SpeechToText.Enabled
	base.EvidenceDag.Enabled = partial.EvidenceDag.Enabled
	base.ImageGeneration.Enabled = partial.ImageGeneration.Enabled

	return base
}
