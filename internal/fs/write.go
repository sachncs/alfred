package fs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alfred/alfred/internal/tool"
)

// WriteInput is the JSON schema for WriteTool input.
type WriteInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// WriteOutput is the structured payload returned by WriteTool.
type WriteOutput struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
}

// WriteTool writes content to a file in the workspace.
type WriteTool struct {
	FileSystemTool
}

// NewWriteTool returns a fresh WriteTool.
func NewWriteTool() *WriteTool {
	return &WriteTool{}
}

// Name implements tool.Tool.
func (*WriteTool) Name() string { return "fs_write" }

// Description implements tool.Tool.
func (*WriteTool) Description() string {
	return "Write content to a file in the workspace. Creates parent directories as needed."
}

// Schema implements tool.Tool.
func (*WriteTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path":    {"type": "string", "description": "Workspace-relative or absolute path"},
			"content": {"type": "string", "description": "File content to write"}
		},
		"required": ["path", "content"]
	}`)
}

// Execute implements tool.Tool.
func (t *WriteTool) Execute(ctx context.Context, raw json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in WriteInput
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

	// Normalize line endings: CRLF → LF.
	content := strings.ReplaceAll(in.Content, "\r\n", "\n")

	// Strip UTF-8 BOM if present.
	content = strings.TrimPrefix(content, "\xEF\xBB\xBF")

	if err := ctx.Err(); err != nil {
		return tool.Failure(err), nil
	}

	// Create parent directories.
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return tool.Failure(err), nil
	}

	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		return tool.Failure(err), nil
	}

	info, _ := os.Stat(abs)
	out := WriteOutput{Path: in.Path, Size: info.Size()}
	return tool.SuccessWith(fmt.Sprintf("wrote %d bytes to %s", out.Size, in.Path), out), nil
}

// hasBOM checks if a byte slice starts with UTF-8 BOM.
func hasBOM(data []byte) bool {
	return bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF})
}
