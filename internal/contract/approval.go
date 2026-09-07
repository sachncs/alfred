package contract

// ApprovalPolicy controls how tool invocations are approved.
type ApprovalPolicy string

const (
	ApprovalAuto      ApprovalPolicy = "auto"
	ApprovalOnRequest ApprovalPolicy = "on-request"
	ApprovalUntrusted ApprovalPolicy = "untrusted"
	ApprovalSuggest   ApprovalPolicy = "suggest"
	ApprovalNever     ApprovalPolicy = "never"
)

// SandboxMode controls the filesystem access level.
type SandboxMode string

const (
	SandboxReadOnly       SandboxMode = "read-only"
	SandboxWorkspaceWrite SandboxMode = "workspace-write"
	SandboxFullAccess     SandboxMode = "danger-full-access"
)
