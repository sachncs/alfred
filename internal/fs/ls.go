package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sachncs/alfred/internal/tool"
)

// LsInput is the JSON schema for LsTool input.
type LsInput struct {
	Path string `json:"path"`
}

// LsEntry is a single directory entry.
type LsEntry struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"isDir"`
	Size    int64  `json:"size"`
	ModTime int64  `json:"modTime"` // unix timestamp
}

// LsOutput is the structured payload returned by LsTool.
type LsOutput struct {
	Entries []LsEntry `json:"entries"`
	Count   int       `json:"count"`
}

// LsTool lists directory contents with metadata.
type LsTool struct {
	FileSystemTool
}

// NewLsTool returns a fresh LsTool.
func NewLsTool() *LsTool { return &LsTool{} }

// Name implements tool.Tool.
func (*LsTool) Name() string { return "fs_ls" }

// Description implements tool.Tool.
func (*LsTool) Description() string {
	return "List directory contents with metadata (size, modtime, isDir)."
}

// Schema implements tool.Tool.
func (*LsTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Workspace-relative or absolute directory path"}
		},
		"required": ["path"]
	}`)
}

// Execute implements tool.Tool.
func (t *LsTool) Execute(ctx context.Context, raw json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in LsInput
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Path == "" {
		return tool.FailureMsg("path is required"), nil
	}

	root := ""
	if tc != nil {
		root = tc.WorkspaceRoot
	}
	abs, err := t.ResolvePath(in.Path, root)
	if err != nil {
		return tool.Failure(err), nil
	}

	entries, err := os.ReadDir(abs)
	if err != nil {
		return tool.Failure(err), nil
	}

	var results []LsEntry
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		results = append(results, LsEntry{
			Name:    e.Name(),
			Path:    filepath.Join(in.Path, e.Name()),
			IsDir:   e.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Unix(),
		})
	}

	out := LsOutput{Entries: results, Count: len(results)}
	return tool.SuccessWith(fmt.Sprintf("%d entries", len(results)), out), nil
}
