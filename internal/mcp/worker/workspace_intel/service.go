package workspaceintel

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alfred/alfred/internal/fs"
	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

func NewWorkspaceIntelServer() *worker.WorkerServer {
	return worker.NewWorkerServer("workspace-intel-worker", []tool.Tool{
		&workspaceListTool{},
		&workspaceReadTool{},
		&workspaceVisualInspectTool{},
	})
}

type workspaceListTool struct{}

func (t *workspaceListTool) Name() string        { return "workspace_list" }
func (t *workspaceListTool) Description() string { return "List files in the workspace" }
func (t *workspaceListTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}}}`)
}
func (t *workspaceListTool) Execute(ctx context.Context, raw json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Path == "" {
		in.Path = "."
	}

	// Delegate to fs.LsTool for real directory listing
	lsTool := fs.NewLsTool()
	input, _ := json.Marshal(map[string]string{"path": in.Path})
	res, err := lsTool.Execute(ctx, input, tc)
	if err != nil {
		return tool.Failure(err), nil
	}
	return res, nil
}

type workspaceReadTool struct{}

func (t *workspaceReadTool) Name() string        { return "workspace_read" }
func (t *workspaceReadTool) Description() string { return "Read a file from the workspace" }
func (t *workspaceReadTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`)
}
func (t *workspaceReadTool) Execute(ctx context.Context, raw json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Path == "" {
		return tool.FailureMsg("path is required"), nil
	}

	// Delegate to fs.ReadTool for real file reading
	readTool := fs.NewReadTool()
	input, _ := json.Marshal(map[string]string{"path": in.Path})
	res, err := readTool.Execute(ctx, input, tc)
	if err != nil {
		return tool.Failure(err), nil
	}
	return res, nil
}

type workspaceVisualInspectTool struct{}

func (t *workspaceVisualInspectTool) Name() string { return "workspace_visual_inspect" }
func (t *workspaceVisualInspectTool) Description() string {
	return "Visual inspection via model router"
}
func (t *workspaceVisualInspectTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`)
}
func (t *workspaceVisualInspectTool) Execute(ctx context.Context, raw json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in struct {
		Path string `json:"path"`
	}
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

	abs := in.Path
	if !filepath.IsAbs(in.Path) && root != "" {
		abs = filepath.Join(root, in.Path)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return tool.Failure(err), nil
	}

	// Build a description based on file metadata
	var b strings.Builder
	fmt.Fprintf(&b, "File: %s\n", filepath.Base(abs))
	fmt.Fprintf(&b, "Size: %d bytes\n", info.Size())
	fmt.Fprintf(&b, "Mode: %s\n", info.Mode())
	fmt.Fprintf(&b, "ModTime: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))

	if info.IsDir() {
		entries, err := os.ReadDir(abs)
		if err == nil {
			fmt.Fprintf(&b, "Type: directory (%d entries)\n", len(entries))
			for i, e := range entries {
				if i >= 20 {
					fmt.Fprintf(&b, "  ... and %d more\n", len(entries)-20)
					break
				}
				etype := "file"
				if e.IsDir() {
					etype = "dir"
				}
				fmt.Fprintf(&b, "  [%s] %s\n", etype, e.Name())
			}
		}
	} else {
		fmt.Fprintf(&b, "Type: file\n")
		ext := strings.ToLower(filepath.Ext(abs))
		if ext != "" {
			fmt.Fprintf(&b, "Extension: %s\n", ext)
		}
	}

	return tool.SuccessWith(b.String(), map[string]any{
		"path":    in.Path,
		"isDir":   info.IsDir(),
		"size":    info.Size(),
		"mode":    info.Mode().String(),
		"modTime": info.ModTime().Format("2006-01-02T15:04:05Z"),
	}), nil
}
