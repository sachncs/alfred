package search

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alfred/alfred/internal/tool"
)

// FindInput is the JSON schema for FindTool input.
type FindInput struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path,omitempty"`
}

// FindResult is a single file found.
type FindResult struct {
	Path string `json:"path"`
}

// FindOutput is the structured payload returned by FindTool.
type FindOutput struct {
	Files []FindResult `json:"files"`
	Count int          `json:"count"`
}

// FindTool discovers files by name pattern, similar to fd.
type FindTool struct{}

// NewFindTool returns a fresh FindTool.
func NewFindTool() *FindTool { return &FindTool{} }

// Name implements tool.Tool.
func (*FindTool) Name() string { return "search_find" }

// Description implements tool.Tool.
func (*FindTool) Description() string {
	return "Find files by name pattern. Skips .git directories."
}

// Schema implements tool.Tool.
func (*FindTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"pattern": {"type": "string", "description": "Glob pattern to match file names (e.g. *.go)"},
			"path":    {"type": "string", "description": "Directory to search in (default: workspace root)"}
		},
		"required": ["pattern"]
	}`)
}

// Execute implements tool.Tool.
func (t *FindTool) Execute(ctx context.Context, raw json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in FindInput
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Pattern == "" {
		return tool.FailureMsg("pattern is required"), nil
	}

	root := "."
	if tc != nil && tc.WorkspaceRoot != "" {
		root = tc.WorkspaceRoot
	}
	if in.Path != "" {
		root = in.Path
	}

	var files []FindResult
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if strings.Contains(path, "/.git/") || strings.HasSuffix(path, "/.git") {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		matched, _ := filepath.Match(in.Pattern, filepath.Base(path))
		if matched {
			files = append(files, FindResult{Path: path})
		}
		return nil
	})
	if err != nil {
		return tool.Failure(err), nil
	}

	out := FindOutput{Files: files, Count: len(files)}
	if len(files) == 0 {
		return tool.SuccessWith("no files found", out), nil
	}
	return tool.SuccessWith(fmt.Sprintf("%d files found", len(files)), out), nil
}

// isBinaryFile checks if data looks like a binary file.
func isBinaryFile(data []byte) bool {
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		for _, b := range scanner.Bytes() {
			if b == 0 {
				return true
			}
		}
	}
	return false
}
