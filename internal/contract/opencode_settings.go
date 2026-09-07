package contract

// OpenCodeRuntimeSettingsV1 configures the OpenCode runtime adapter.
type OpenCodeRuntimeSettingsV1 struct {
	Command        string   `json:"command"`
	AutoStart      bool     `json:"autoStart"`
	Port           int      `json:"port"`
	Hostname       string   `json:"hostname"`
	Model          string   `json:"model"`
	Provider       string   `json:"provider"`
	Agent          string   `json:"agent"`
	ServerPassword string   `json:"serverPassword"`
	ExtraArgs      []string `json:"extraArgs"`
}
