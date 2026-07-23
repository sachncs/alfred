package contract

// GUIUpdateChannel is the release channel for GUI updates.
type GUIUpdateChannel string

const (
	GUIChannelStable GUIUpdateChannel = "stable"
	GUIChannelBeta   GUIUpdateChannel = "beta"
	GUIChannelDev    GUIUpdateChannel = "dev"
)

// GUIUpdateState tracks the current update status.
type GUIUpdateState string

const (
	GUIStateIdle       GUIUpdateState = "idle"
	GUIStateChecking   GUIUpdateState = "checking"
	GUIStateReady      GUIUpdateState = "ready"
	GUIStateInstalling GUIUpdateState = "installing"
	GUIStateFailed     GUIUpdateState = "failed"
)
