package capability

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// FileCapability is a ResourceCapability that manages a file handle.
type FileCapability struct {
	BaseResource
	id   ID
	path string
}

// NewFileCapability creates a FileCapability for the given path.
func NewFileCapability(id ID, path string) *FileCapability {
	return &FileCapability{id: id, path: path}
}

func (f *FileCapability) ID() ID                             { return f.id }
func (f *FileCapability) IsAvailable(_ context.Context) bool { return true }

func (f *FileCapability) Invoke(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var req struct {
		Op   string `json:"op"`   // "read", "write", "stat"
		Path string `json:"path"` // optional override
		Data string `json:"data"` // for write
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("file: %w", err)
	}
	p := f.path
	if req.Path != "" {
		p = filepath.Join(filepath.Dir(f.path), req.Path)
	}
	switch req.Op {
	case "read":
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		return json.Marshal(map[string]string{"content": string(data), "path": p})
	case "write":
		if err := os.WriteFile(p, []byte(req.Data), 0644); err != nil {
			return nil, err
		}
		return json.Marshal(map[string]string{"status": "written", "path": p})
	case "stat":
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		return json.Marshal(map[string]any{"path": p, "size": info.Size(), "mode": info.Mode().String()})
	default:
		return nil, fmt.Errorf("file: unknown op %q", req.Op)
	}
}

// ReadOnlyFileCapability wraps a FileCapability and rejects writes.
type ReadOnlyFileCapability struct {
	FileCapability
}

// NewReadOnlyFileCapability creates a read-only file capability.
func NewReadOnlyFileCapability(id ID, path string) *ReadOnlyFileCapability {
	return &ReadOnlyFileCapability{FileCapability: *NewFileCapability(id, path)}
}

func (r *ReadOnlyFileCapability) Invoke(_ context.Context, input json.RawMessage) (json.RawMessage, error) {
	var req struct {
		Op string `json:"op"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("readonly_file: %w", err)
	}
	if req.Op == "write" {
		return nil, fmt.Errorf("readonly_file: write denied for %s", r.id)
	}
	return r.FileCapability.Invoke(context.Background(), input)
}

var (
	_ Capability         = (*FileCapability)(nil)
	_ ResourceCapability = (*FileCapability)(nil)
	_ Capability         = (*ReadOnlyFileCapability)(nil)
)
