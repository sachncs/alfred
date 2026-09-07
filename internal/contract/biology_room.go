package contract

// BiologyRoomAsset represents a file in the biology room.
type BiologyRoomAsset struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Size     int64  `json:"size"`
	MIMEType string `json:"mimeType"`
}

// BiologyRoomContig is a DNA/protein sequence contig.
type BiologyRoomContig struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Length   int    `json:"length"`
	Sequence string `json:"sequence,omitempty"`
}

// BiologyRoomSelection is a region selection on an asset.
type BiologyRoomSelection struct {
	AssetID string `json:"assetId"`
	Start   int    `json:"start"`
	End     int    `json:"end"`
	Label   string `json:"label,omitempty"`
}

// BiologyRoomMutation records a modification to a sequence.
type BiologyRoomMutation struct {
	ContigID   string `json:"contigId"`
	Position   int    `json:"position"`
	OldResidue string `json:"oldResidue"`
	NewResidue string `json:"newResidue"`
}

// BiologyRoomHistoryEntry is one history entry.
type BiologyRoomHistoryEntry struct {
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
	Detail    string `json:"detail,omitempty"`
}
