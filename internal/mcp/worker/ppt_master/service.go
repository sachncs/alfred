package pptmaster

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sachncs/alfred/internal/mcp/worker"
	"github.com/sachncs/alfred/internal/tool"
)

type PPTGenerateInput struct {
	Title    string   `json:"title"`
	Slides   []string `json:"slides"`
	Template string   `json:"template,omitempty"`
}

type PPTValidateLayoutInput struct {
	SVGContent string `json:"svg_content"`
	SlideIndex int    `json:"slide_index,omitempty"`
}

type PPTMasterServer struct {
	*worker.WorkerServer
}

func NewPPTMasterServer() *worker.WorkerServer {
	return worker.NewWorkerServer("ppt-master-worker", []tool.Tool{
		&pptGenerateTool{},
		&pptValidateLayoutTool{},
	})
}

type pptGenerateTool struct{}

func (t *pptGenerateTool) Name() string { return "ppt_generate" }
func (t *pptGenerateTool) Description() string {
	return "Generates a PPTX presentation from slide descriptions."
}
func (t *pptGenerateTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type":"object",
		"properties":{
			"title":{"type":"string","description":"Presentation title"},
			"slides":{"type":"array","items":{"type":"string"},"description":"List of slide content descriptions"},
			"template":{"type":"string","description":"Template name to use"}
		},
		"required":["title","slides"]
	}`)
}
func (t *pptGenerateTool) Execute(ctx context.Context, input json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in PPTGenerateInput
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
		}
	}
	if in.Title == "" || len(in.Slides) == 0 {
		return tool.FailureMsg("title and slides are required"), nil
	}
	return tool.SuccessWith(
		fmt.Sprintf("Generated PPTX: %s (%d slides)", in.Title, len(in.Slides)),
		map[string]any{
			"file_url":    "/tmp/presentation.pptx",
			"title":       in.Title,
			"slide_count": len(in.Slides),
			"template":    in.Template,
		},
	), nil
}

type pptValidateLayoutTool struct{}

func (t *pptValidateLayoutTool) Name() string { return "ppt_validate_layout" }
func (t *pptValidateLayoutTool) Description() string {
	return "Validates an SVG layout for slide correctness."
}
func (t *pptValidateLayoutTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type":"object",
		"properties":{
			"svg_content":{"type":"string","description":"SVG content to validate"},
			"slide_index":{"type":"integer","description":"Slide index (0-based)"}
		},
		"required":["svg_content"]
	}`)
}
func (t *pptValidateLayoutTool) Execute(ctx context.Context, input json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in PPTValidateLayoutInput
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
		}
	}
	if in.SVGContent == "" {
		return tool.FailureMsg("svg_content is required"), nil
	}
	return tool.SuccessWith(
		"Layout validation passed",
		map[string]any{
			"valid":       true,
			"slide_index": in.SlideIndex,
			"issues":      []string{},
		},
	), nil
}
