package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/alfred/alfred/internal/tool"
)

// ApplyPatchInput is the JSON schema for ApplyPatchTool input.
type ApplyPatchInput struct {
	Path   string `json:"path"`
	Patch  string `json:"patch"`
	DryRun bool   `json:"dryRun,omitempty"`
}

// ApplyPatchOutput is the structured payload returned by ApplyPatchTool.
type ApplyPatchOutput struct {
	Path   string `json:"path"`
	Status string `json:"status"`
}

// ApplyPatchTool applies unified diff patches to files.
type ApplyPatchTool struct {
	FileSystemTool
}

// NewApplyPatchTool returns a fresh ApplyPatchTool.
func NewApplyPatchTool() *ApplyPatchTool {
	return &ApplyPatchTool{}
}

// Name implements tool.Tool.
func (*ApplyPatchTool) Name() string { return "fs_apply_patch" }

// Description implements tool.Tool.
func (*ApplyPatchTool) Description() string {
	return "Apply a unified diff patch to a file. Supports dry-run mode."
}

// Schema implements tool.Tool.
func (*ApplyPatchTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path":    {"type": "string", "description": "Workspace-relative or absolute path"},
			"patch":   {"type": "string", "description": "Unified diff patch content"},
			"dryRun":  {"type": "boolean", "description": "Validate without writing"}
		},
		"required": ["path", "patch"]
	}`)
}

// Execute implements tool.Tool.
func (t *ApplyPatchTool) Execute(ctx context.Context, raw json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in ApplyPatchInput
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Path == "" || in.Patch == "" {
		return tool.FailureMsg("path and patch are required"), nil
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
	patches := parsePatch(in.Patch)

	for _, p := range patches {
		idx := strings.Index(content, p.Old)
		if idx == -1 {
			return tool.FailureMsg(fmt.Sprintf("hunk not found in %s: %q", in.Path, truncate(p.Old, 50))), nil
		}
		content = content[:idx] + p.New + content[idx+len(p.Old):]
	}

	if in.DryRun {
		return tool.SuccessWith("dry run: patch is valid", ApplyPatchOutput{
			Path:   in.Path,
			Status: "dry_run",
		}), nil
	}

	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		return tool.Failure(err), nil
	}

	return tool.SuccessWith(fmt.Sprintf("applied %d hunk(s) to %s", len(patches), in.Path), ApplyPatchOutput{
		Path:   in.Path,
		Status: "applied",
	}), nil
}

type patchHunk struct {
	Old string
	New string
}

// parsePatch extracts old/new text pairs from a unified diff.
// It handles the standard format:
//
//	--- a/file.txt
//	+++ b/file.txt
//	@@ -1 +1 @@
//	-old line
//	+new line
func parsePatch(patch string) []patchHunk {
	var hunks []patchHunk
	lines := strings.Split(patch, "\n")

	var oldLines, newLines []string
	inHunk := false

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "--- "):
			// Start of a new hunk — flush previous.
			if inHunk && (len(oldLines) > 0 || len(newLines) > 0) {
				hunks = append(hunks, patchHunk{
					Old: strings.Join(oldLines, "\n"),
					New: strings.Join(newLines, "\n"),
				})
				oldLines, newLines = nil, nil
			}
		case strings.HasPrefix(line, "+++ "):
			// New file header — mark that we're in a hunk.
			inHunk = true
		case strings.HasPrefix(line, "@@ "):
			// Hunk header — flush previous and reset.
			if len(oldLines) > 0 || len(newLines) > 0 {
				hunks = append(hunks, patchHunk{
					Old: strings.Join(oldLines, "\n"),
					New: strings.Join(newLines, "\n"),
				})
				oldLines, newLines = nil, nil
			}
			inHunk = true
		case strings.HasPrefix(line, "-") && inHunk:
			oldLines = append(oldLines, line[1:])
		case strings.HasPrefix(line, "+") && inHunk:
			newLines = append(newLines, line[1:])
		}
	}

	// Flush last hunk.
	if len(oldLines) > 0 || len(newLines) > 0 {
		hunks = append(hunks, patchHunk{
			Old: strings.Join(oldLines, "\n"),
			New: strings.Join(newLines, "\n"),
		})
	}

	return hunks
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
