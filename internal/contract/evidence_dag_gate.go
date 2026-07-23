package contract

// EvidenceDAGGate is a high-impact action gate.
type EvidenceDAGGate struct {
	ID          string `json:"id"`
	Level       string `json:"level"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Approved    bool   `json:"approved"`
}
