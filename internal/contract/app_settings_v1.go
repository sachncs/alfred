package contract

// AppSettingsV1 is the top-level settings schema for Alfred.
type AppSettingsV1 struct {
	Version           string                  `json:"version"`
	Agents            AgentSettings           `json:"agents"`
	LocalRuntime      LocalRuntimeSettingsV1  `json:"localRuntime"`
	ModelRouter       ModelRouterSettings     `json:"modelRouter"`
	ModelAccess       ModelAccessSettings     `json:"modelAccess"`
	ComputerUse       ComputerUseSettings     `json:"computerUse"`
	RemoteChannel     RemoteChannelSettings   `json:"remoteChannel"`
	RemoteExecutor    RemoteExecutorSettings  `json:"remoteExecutor"`
	Schedule          ScheduleSettings        `json:"schedule"`
	Workflow          WorkflowSettings        `json:"workflow"`
	Write             WriteSettings           `json:"write"`
	SpeechToText      SpeechToTextSettings    `json:"speechToText"`
	EvidenceDag       EvidenceDagSettings     `json:"evidenceDag"`
	ImageGeneration   ImageGenerationSettings `json:"imageGeneration"`
	AgentCapabilities AgentCapabilitySettings `json:"agentCapabilities"`
	KeyboardShortcuts KeyboardShortcutCatalog `json:"keyboardShortcuts"`
	GUIUpdate         GUIUpdateSettings       `json:"guiUpdate"`
	AppBehavior       AppBehaviorSettings     `json:"appBehavior"`
}

// AgentSettings holds agent runtime configuration.
type AgentSettings struct {
	ActiveRuntime AgentRuntimeId            `json:"activeRuntime"`
	OpenCode      OpenCodeRuntimeSettingsV1 `json:"openCode"`
}

// ModelRouterSettings holds model routing configuration.
type ModelRouterSettings struct {
	DefaultModel string            `json:"defaultModel"`
	RoutingRules map[string]string `json:"routingRules,omitempty"`
}

// ModelAccessSettings controls model access policies.
type ModelAccessSettings struct {
	AllowedModels []string `json:"allowedModels"`
	MaxTokens     int      `json:"maxTokens"`
}

// ComputerUseSettings configures computer-use capabilities.
type ComputerUseSettings struct {
	Enabled bool `json:"enabled"`
}

// RemoteChannelSettings configures the remote channel runtime.
type RemoteChannelSettings struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url,omitempty"`
}

// RemoteExecutorSettings configures the remote executor worker.
type RemoteExecutorSettings struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url,omitempty"`
}

// ScheduleSettings configures the schedule worker.
type ScheduleSettings struct {
	Enabled bool `json:"enabled"`
}

// WorkflowSettings configures the workflow DSL runtime.
type WorkflowSettings struct {
	Enabled bool `json:"enabled"`
}

// WriteSettings configures write-mode features.
type WriteSettings struct {
	Enabled bool `json:"enabled"`
}

// SpeechToTextSettings configures speech-to-text.
type SpeechToTextSettings struct {
	Enabled  bool   `json:"enabled"`
	Protocol string `json:"protocol"`
}

// EvidenceDagSettings configures the evidence DAG worker.
type EvidenceDagSettings struct {
	Enabled bool `json:"enabled"`
}

// ImageGenerationSettings configures the image generation worker.
type ImageGenerationSettings struct {
	Enabled  bool   `json:"enabled"`
	Provider string `json:"provider,omitempty"`
}

// GUIUpdateSettings configures GUI update behavior.
type GUIUpdateSettings struct {
	Channel GUIUpdateChannel `json:"channel"`
}

// AppBehaviorSettings holds general app behavior toggles.
type AppBehaviorSettings struct {
	Theme     string `json:"theme"`
	Language  string `json:"language"`
	UIScale   int    `json:"uiScale"`
	FontScale int    `json:"fontScale"`
	AutoStart bool   `json:"autoStart"`
}
