package settings

import "github.com/sachncs/alfred/internal/contract"

// Normalize takes a partial AppSettingsV1 and fills in missing fields with defaults.
func Normalize(partial contract.AppSettingsV1) contract.AppSettingsV1 {
	def := DefaultAppSettings()

	if partial.Version == "" {
		partial.Version = def.Version
	}
	if partial.Agents.ActiveRuntime == "" {
		partial.Agents.ActiveRuntime = def.Agents.ActiveRuntime
	}
	if partial.Agents.OpenCode.Command == "" {
		partial.Agents.OpenCode = def.Agents.OpenCode
	}
	if partial.LocalRuntime.Port == 0 {
		partial.LocalRuntime.Port = def.LocalRuntime.Port
	}
	if partial.LocalRuntime.Model == "" {
		partial.LocalRuntime.Model = def.LocalRuntime.Model
	}
	if partial.LocalRuntime.ApprovalPolicy == "" {
		partial.LocalRuntime.ApprovalPolicy = def.LocalRuntime.ApprovalPolicy
	}
	if partial.LocalRuntime.SandboxMode == "" {
		partial.LocalRuntime.SandboxMode = def.LocalRuntime.SandboxMode
	}
	if partial.ModelRouter.DefaultModel == "" {
		partial.ModelRouter.DefaultModel = def.ModelRouter.DefaultModel
	}
	if len(partial.ModelAccess.AllowedModels) == 0 {
		partial.ModelAccess.AllowedModels = def.ModelAccess.AllowedModels
	}
	if partial.ModelAccess.MaxTokens == 0 {
		partial.ModelAccess.MaxTokens = def.ModelAccess.MaxTokens
	}
	if partial.AgentCapabilities.MaxDelegationDepth == 0 {
		partial.AgentCapabilities = def.AgentCapabilities
	}
	if len(partial.KeyboardShortcuts.Shortcuts) == 0 {
		partial.KeyboardShortcuts = def.KeyboardShortcuts
	}
	if partial.GUIUpdate.Channel == "" {
		partial.GUIUpdate.Channel = def.GUIUpdate.Channel
	}
	if partial.AppBehavior.Theme == "" {
		partial.AppBehavior.Theme = def.AppBehavior.Theme
	}
	if partial.AppBehavior.Language == "" {
		partial.AppBehavior.Language = def.AppBehavior.Language
	}
	if partial.AppBehavior.UIScale == 0 {
		partial.AppBehavior.UIScale = def.AppBehavior.UIScale
	}
	if partial.AppBehavior.FontScale == 0 {
		partial.AppBehavior.FontScale = def.AppBehavior.FontScale
	}
	return partial
}
