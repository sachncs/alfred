package contract

import "time"

// CreateThreadRequest is the wire schema for POST /v1/threads.
type CreateThreadRequest struct {
	Title    string         `json:"title,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// CreateThreadResponse is the wire schema for the response of POST /v1/threads.
type CreateThreadResponse struct {
	Thread Thread `json:"thread"`
}

// ListThreadsResponse is the wire schema for GET /v1/threads.
type ListThreadsResponse struct {
	Threads   []Thread `json:"threads"`
	NextToken string   `json:"nextToken,omitempty"`
}

// StartTurnRequest is the wire schema for POST /v1/threads/:id/turns.
type StartTurnRequest struct {
	Input UserInput `json:"input"`
}

// StartTurnResponse is the wire schema for the response of POST /v1/threads/:id/turns.
type StartTurnResponse struct {
	Turn Turn `json:"turn"`
}

// HealthResponse is the wire schema for GET /healthz.
type HealthResponse struct {
	Status    string    `json:"status"`
	Version   string    `json:"version,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
