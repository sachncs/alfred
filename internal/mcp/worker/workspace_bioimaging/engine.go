package workspacebioimaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

func NewBioimagingServer() *worker.WorkerServer {
	return worker.NewWorkerServer("workspace-bioimaging-worker", []tool.Tool{
		&bioimagingPreviewTool{},
		&bioimagingSelectRegionTool{},
		&bioimagingExportROITool{},
	})
}

type bioimagingPreviewTool struct{}

func (t *bioimagingPreviewTool) Name() string        { return "bioimaging_preview" }
func (t *bioimagingPreviewTool) Description() string { return "Preview a bioimaging file (OME/TIFF)" }
func (t *bioimagingPreviewTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`)
}
func (t *bioimagingPreviewTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	return tool.SuccessWith("Bioimaging preview", map[string]any{"path": in.Path, "type": "OME-TIFF"}), nil
}

type bioimagingSelectRegionTool struct{}

func (t *bioimagingSelectRegionTool) Name() string        { return "bioimaging_select_region" }
func (t *bioimagingSelectRegionTool) Description() string { return "Select a region of interest" }
func (t *bioimagingSelectRegionTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"x":{"type":"integer"},"y":{"type":"integer"},"width":{"type":"integer"},"height":{"type":"integer"}},"required":["path"]}`)
}
func (t *bioimagingSelectRegionTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Path   string `json:"path"`
		X      int    `json:"x"`
		Y      int    `json:"y"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	}
	_ = json.Unmarshal(raw, &in)
	return tool.SuccessWith("Region selected", map[string]any{"selectionId": "sel-001"}), nil
}

type bioimagingExportROITool struct{}

func (t *bioimagingExportROITool) Name() string        { return "bioimaging_export_roi" }
func (t *bioimagingExportROITool) Description() string { return "Export a region of interest" }
func (t *bioimagingExportROITool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"selectionId":{"type":"string"}},"required":["selectionId"]}`)
}
func (t *bioimagingExportROITool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		SelectionID string `json:"selectionId"`
	}
	_ = json.Unmarshal(raw, &in)
	return tool.SuccessWith("ROI exported", map[string]any{"path": "/tmp/roi-export.tiff"}), nil
}
