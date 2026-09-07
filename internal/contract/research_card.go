package contract

// ResearchCard is a card representing a research item.
type ResearchCard struct {
	ID        string         `json:"id"`
	Kind      string         `json:"kind"`
	Stage     string         `json:"stage"`
	Status    string         `json:"status"`
	Priority  string         `json:"priority"`
	Refs      []string       `json:"refs,omitempty"`
	Decisions []string       `json:"decisions,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}
