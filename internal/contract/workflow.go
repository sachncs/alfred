package contract

// WorkflowV1 is a workflow definition.
type WorkflowV1 struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Nodes       []WorkflowNodeV1       `json:"nodes"`
	Connections []WorkflowConnectionV1 `json:"connections"`
}

// WorkflowNodeV1 is a node in a workflow graph.
type WorkflowNodeV1 struct {
	ID     string         `json:"id"`
	Type   string         `json:"type"`
	Name   string         `json:"name"`
	Config map[string]any `json:"config,omitempty"`
}

// WorkflowConnectionV1 connects two nodes.
type WorkflowConnectionV1 struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// WorkflowRunV1 is a workflow execution record.
type WorkflowRunV1 struct {
	ID         string `json:"id"`
	WorkflowID string `json:"workflowId"`
	Status     string `json:"status"`
	StartedAt  string `json:"startedAt,omitempty"`
	EndedAt    string `json:"endedAt,omitempty"`
}
