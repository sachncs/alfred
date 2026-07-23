package contract

// ScientificObject is a reference to a scientific entity.
type ScientificObject struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Name     string         `json:"name"`
	Refs     []string       `json:"refs,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ScientificComparison compares two scientific objects.
type ScientificComparison struct {
	ObjectA ScientificObject `json:"objectA"`
	ObjectB ScientificObject `json:"objectB"`
	Diff    string           `json:"diff"`
	Summary string           `json:"summary"`
}
