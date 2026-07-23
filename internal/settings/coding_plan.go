package settings

// CodingPlanAdapter describes a coding-plan integration.
type CodingPlanAdapter struct {
	ID          string `json:"id"`
	RuntimeID   string `json:"runtimeId"`
	UpstreamURL string `json:"upstreamURL"`
	Model       string `json:"model"`
}

// CodingPlanAdapters is the catalog of known coding plan integrations.
var CodingPlanAdapters = []CodingPlanAdapter{
	{
		ID:          "opencode",
		RuntimeID:   "opencode",
		UpstreamURL: "http://127.0.0.1:8080",
		Model:       "gpt-4o",
	},
}

// FindAdapter returns the adapter with the given ID, or nil.
func FindAdapter(id string) *CodingPlanAdapter {
	for _, a := range CodingPlanAdapters {
		if a.ID == id {
			return &a
		}
	}
	return nil
}
