package contract

// VisualStyle describes rendering style for visual outputs.
type VisualStyle struct {
	Theme    string            `json:"theme,omitempty"`
	Colors   map[string]string `json:"colors,omitempty"`
	Font     string            `json:"font,omitempty"`
	FontSize int               `json:"fontSize,omitempty"`
}

// VisualDocument represents a visual document for review.
type VisualDocument struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Path   string `json:"path"`
}
