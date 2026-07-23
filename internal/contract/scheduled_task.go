package contract

// ScheduledTaskV1 is a scheduled task definition.
type ScheduledTaskV1 struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Kind     string         `json:"kind"`
	Schedule string         `json:"schedule"`
	Config   map[string]any `json:"config,omitempty"`
	Enabled  bool           `json:"enabled"`
	LastRun  string         `json:"lastRun,omitempty"`
	NextRun  string         `json:"nextRun,omitempty"`
}
