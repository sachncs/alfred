package contract

// WorkspaceEntry represents a file in the workspace.
type WorkspaceEntry struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	IsDir    bool   `json:"isDir"`
	Size     int64  `json:"size"`
	MIMEType string `json:"mimeType,omitempty"`
}

// WorkspaceFile is a workspace file reference.
type WorkspaceFile struct {
	ID       string `json:"id"`
	Path     string `json:"path"`
	Name     string `json:"name"`
	MIMEType string `json:"mimeType"`
}

// WorkspaceClipboard holds clipboard payload.
type WorkspaceClipboard struct {
	Content  string `json:"content"`
	MIMEType string `json:"mimeType"`
}

// WorkspaceTreeNode is a node in the workspace tree.
type WorkspaceTreeNode struct {
	Name     string               `json:"name"`
	Path     string               `json:"path"`
	IsDir    bool                 `json:"isDir"`
	Children []*WorkspaceTreeNode `json:"children,omitempty"`
}
