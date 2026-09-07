package contract

// LocalRuntimeSettingsV1 configures the local Alfred runtime.
type LocalRuntimeSettingsV1 struct {
	BinaryPath     string            `json:"binaryPath"`
	Port           int               `json:"port"`
	AutoStart      bool              `json:"autoStart"`
	RuntimeToken   string            `json:"runtimeToken"`
	DataDir        string            `json:"dataDir"`
	Model          string            `json:"model"`
	ApprovalPolicy ApprovalPolicy    `json:"approvalPolicy"`
	SandboxMode    SandboxMode       `json:"sandboxMode"`
	TokenEconomy   bool              `json:"tokenEconomyMode"`
	Insecure       bool              `json:"insecure"`
	MCPSearch      MCPSearchSettings `json:"mcpSearch"`
	Storage        StorageSettings   `json:"storage"`
	Tuning         RuntimeTuning     `json:"runtimeTuning"`
	Guards         RuntimeGuards     `json:"runtimeGuards"`
}

// MCPSearchSettings configures the MCP tool search.
type MCPSearchSettings struct {
	Enabled    bool `json:"enabled"`
	MaxResults int  `json:"maxResults"`
}

// StorageSettings configures data storage paths.
type StorageSettings struct {
	MaxFileSize int64 `json:"maxFileSize"`
	MaxThreads  int   `json:"maxThreads"`
}

// RuntimeTuning holds performance tuning knobs.
type RuntimeTuning struct {
	MaxConcurrentTurns int `json:"maxConcurrentTurns"`
	TimeoutSeconds     int `json:"timeoutSeconds"`
}

// RuntimeGuards holds safety limits.
type RuntimeGuards struct {
	MaxToolCalls     int `json:"maxToolCalls"`
	MaxTokensPerTurn int `json:"maxTokensPerTurn"`
}
