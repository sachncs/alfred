// Package fs provides filesystem-backed tools. Every concrete file tool
// extends FileSystemTool to share path resolution and sandbox policy.
package fs

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/alfred/alfred/internal/tool"
)

// FileSystemTool is the abstract base for any tool that operates on
// filesystem paths. It is the only fs tool that knows how to resolve
// paths against the active workspace; concrete tools delegate to it.
type FileSystemTool struct {
	tool.Tool
}

// ResolvePath returns an absolute, cleaned path. If rel is absolute, it
// is returned after cleaning. Otherwise, rel is joined to workspaceRoot
// and cleaned.
func (t *FileSystemTool) ResolvePath(rel, workspaceRoot string) (string, error) {
	if workspaceRoot == "" {
		return "", fmt.Errorf("workspace root is empty")
	}
	if rel == "" {
		return "", fmt.Errorf("path is empty")
	}
	var p string
	if filepath.IsAbs(rel) {
		p = filepath.Clean(rel)
	} else {
		p = filepath.Clean(filepath.Join(workspaceRoot, rel))
	}
	// Containment check: resolved path must remain inside workspace root.
	absRoot, err := filepath.Abs(workspaceRoot)
	if err != nil {
		return "", fmt.Errorf("resolve workspace root: %w", err)
	}
	absPath, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	if absPath != absRoot && !isInside(absPath, absRoot) {
		return "", fmt.Errorf("path %q escapes workspace root %q", rel, workspaceRoot)
	}
	return absPath, nil
}

func isInside(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	return true
}
