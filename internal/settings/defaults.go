package settings

import "github.com/alfred/alfred/internal/contract"

// DefaultAppSettings returns a populated AppSettingsV1 with safe defaults.
func DefaultAppSettings() contract.AppSettingsV1 {
	return contract.AppSettingsV1{
		Version: "1",
		Agents: contract.AgentSettings{
			ActiveRuntime: contract.AgentRuntimeAlfred,
			OpenCode: contract.OpenCodeRuntimeSettingsV1{
				Command:   "opencode",
				AutoStart: false,
				Port:      8080,
				Hostname:  "127.0.0.1",
			},
		},
		LocalRuntime: contract.LocalRuntimeSettingsV1{
			BinaryPath:     "",
			Port:           8899,
			AutoStart:      true,
			Model:          "gpt-4o",
			ApprovalPolicy: contract.ApprovalOnRequest,
			SandboxMode:    contract.SandboxWorkspaceWrite,
			TokenEconomy:   true,
			MCPSearch: contract.MCPSearchSettings{
				Enabled:    true,
				MaxResults: 10,
			},
			Storage: contract.StorageSettings{
				MaxFileSize: 10 * 1024 * 1024,
				MaxThreads:  1000,
			},
			Tuning: contract.RuntimeTuning{
				MaxConcurrentTurns: 3,
				TimeoutSeconds:     300,
			},
			Guards: contract.RuntimeGuards{
				MaxToolCalls:     50,
				MaxTokensPerTurn: 128000,
			},
		},
		ModelRouter: contract.ModelRouterSettings{
			DefaultModel: "gpt-4o",
		},
		ModelAccess: contract.ModelAccessSettings{
			AllowedModels: []string{"gpt-4o", "gpt-4o-mini", "claude-sonnet-4-20250514"},
			MaxTokens:     128000,
		},
		AgentCapabilities: contract.AgentCapabilitySettings{
			MaxDelegationDepth: 2,
			MaxSubAgents:       5,
		},
		KeyboardShortcuts: contract.KeyboardShortcutCatalog{
			Shortcuts: []contract.KeyboardShortcut{
				{Command: "submit", Binding: "Enter"},
				{Command: "newThread", Binding: "Cmd+N"},
				{Command: "search", Binding: "Cmd+K"},
			},
		},
		GUIUpdate: contract.GUIUpdateSettings{
			Channel: contract.GUIChannelStable,
		},
		AppBehavior: contract.AppBehaviorSettings{
			Theme:     "dark",
			Language:  "en",
			UIScale:   100,
			FontScale: 100,
			AutoStart: true,
		},
	}
}
