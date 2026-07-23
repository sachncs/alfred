package contract

// RemoteExecutorTarget is an SSH/Slurm target.
type RemoteExecutorTarget struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	User    string `json:"user"`
	Type    string `json:"type"`
	Trusted bool   `json:"trusted"`
}

// RemoteExecutorJob is a remote execution job.
type RemoteExecutorJob struct {
	ID       string `json:"id"`
	TargetID string `json:"targetId"`
	Command  string `json:"command"`
	Status   string `json:"status"`
	Output   string `json:"output,omitempty"`
}
