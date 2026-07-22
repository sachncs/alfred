package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/alfred/alfred/internal/tool"
)

// EditInput is the JSON schema for EditTool input.
type EditInput struct {
	Path    string `json:"path"`
	Old     string `json:"old"`
	New     string `json:"new"`
	Fuzzy   bool   `json:"fuzzy,omitempty"`
}

// EditOutput is the structured payload returned by EditTool.
type EditOutput struct {
	Path string `json:"path"`
	Diff string `json:"diff"`
}

// EditTool performs search-and-replace on a file with unified diff output.
type EditTool struct {
	FileSystemTool
}

// NewEditTool returns a fresh EditTool.
func NewEditTool() *EditTool {
	return &EditTool{}
}

// Name implements tool.Tool.
func (*EditTool) Name() string { return "fs_edit" }

// Description implements tool.Tool.
func (*EditTool) Description() string {
	return "Edit a file by replacing text. Supports exact and fuzzy matching."
}

// Schema implements tool.Tool.
func (*EditTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path":  {"type": "string", "description": "Workspace-relative or absolute path"},
			"old":   {"type": "string", "description": "Exact text to find and replace"},
			"new":   {"type": "string", "description": "Replacement text"},
			"fuzzy": {"type": "boolean", "description": "Use whitespace-normalized matching"}
		},
		"required": ["path", "old", "new"]
	}`)
}

// Execute implements tool.Tool.
func (t *EditTool) Execute(ctx context.Context, raw json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in EditInput
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Path == "" || in.Old == "" {
		return tool.FailureMsg("path and old are required"), nil
	}

	root := ""
	if tc != nil {
		root = tc.WorkspaceRoot
	}
	abs, err := t.ResolvePath(in.Path, root)
	if err != nil {
		return tool.Failure(err), nil
	}

	data, err := os.ReadFile(abs)
	if err != nil {
		return tool.Failure(err), nil
	}

	content := string(data)

	// Find the old text.
	oldText := in.Old
	if in.Fuzzy {
		oldText = normalizeWhitespace(in.Old)
		content = normalizeWhitespace(content)
	}

	idx := strings.Index(content, oldText)
	if idx == -1 {
		return tool.FailureMsg(fmt.Sprintf("text not found in %s", in.Path)), nil
	}

	// Check for multiple matches.
	rest := content[idx+len(oldText):]
	if strings.Index(rest, oldText) != -1 {
		return tool.FailureMsg(fmt.Sprintf("multiple matches found in %s — provide more context", in.Path)), nil
	}

	// Perform the replacement.
	newContent := content[:idx] + in.New + content[idx+len(oldText):]

	// Generate unified diff.
	diff := unifiedDiff(abs, content, newContent)

	// Write the result.
	if err := os.WriteFile(abs, []byte(newContent), 0o644); err != nil {
		return tool.Failure(err), nil
	}

	out := EditOutput{Path: in.Path, Diff: diff}
	return tool.SuccessWith(fmt.Sprintf("edited %s", in.Path), out), nil
}

// normalizeWhitespace collapses runs of whitespace into single spaces
// and trims leading/trailing whitespace per line.
func normalizeWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		fields := strings.Fields(line)
		lines[i] = strings.Join(fields, " ")
	}
	return strings.Join(lines, "\n")
}

// unifiedDiff produces a minimal unified diff between old and new content.
func unifiedDiff(path, old, new string) string {
	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	var diff strings.Builder
	fmt.Fprintf(&diff, "--- a/%s\n+++ b/%s\n", path, path)

	// Simple line-by-line diff.
	maxLines := len(oldLines)
	if len(newLines) > maxLines {
		maxLines = len(newLines)
	}

	for i := 0; i < maxLines; i++ {
		oldLine := ""
		newLine := ""
		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}
		if oldLine != newLine {
			if i < len(oldLines) {
				fmt.Fprintf(&diff, "-%s\n", oldLine)
			}
			if i < len(newLines) {
				fmt.Fprintf(&diff, "+%s\n", newLine)
			}
		}
	}

	return diff.String()
}
