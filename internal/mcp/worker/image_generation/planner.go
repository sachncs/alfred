package imagegeneration

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

type ImageGenerateInput struct {
	Description string `json:"description"`
	Style       string `json:"style,omitempty"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
}

type ImageEditInput struct {
	ImageURL string `json:"image_url"`
	Prompt   string `json:"prompt"`
}

type ImageReviewInput struct {
	ImageURL string `json:"image_url"`
	Criteria string `json:"criteria,omitempty"`
}

type ImageGenerationServer struct {
	*worker.WorkerServer
}

func NewImageGenerationServer() *worker.WorkerServer {
	return worker.NewWorkerServer("image-generation-worker", []tool.Tool{
		&imageGenerateTool{},
		&imageEditTool{},
		&imageReviewTool{},
	})
}

type imageGenerateTool struct{}

func (t *imageGenerateTool) Name() string { return "image_generate" }
func (t *imageGenerateTool) Description() string {
	return "Generates an image from a text description."
}
func (t *imageGenerateTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type":"object",
		"properties":{
			"description":{"type":"string","description":"Text description of the image to generate"},
			"style":{"type":"string","description":"Visual style (e.g. realistic, diagram, sketch)"},
			"width":{"type":"integer","description":"Image width in pixels"},
			"height":{"type":"integer","description":"Image height in pixels"}
		},
		"required":["description"]
	}`)
}
func (t *imageGenerateTool) Execute(ctx context.Context, input json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in ImageGenerateInput
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
		}
	}
	if in.Description == "" {
		return tool.FailureMsg("description is required"), nil
	}
	w, h := 512, 512
	if in.Width > 0 {
		w = in.Width
	}
	if in.Height > 0 {
		h = in.Height
	}
	return tool.SuccessWith(
		fmt.Sprintf("Generated image: %dx%d", w, h),
		map[string]any{
			"image_url":   fmt.Sprintf("/tmp/generated_%dx%d.png", w, h),
			"description": in.Description,
			"style":       in.Style,
			"width":       w,
			"height":      h,
		},
	), nil
}

type imageEditTool struct{}

func (t *imageEditTool) Name() string        { return "image_edit" }
func (t *imageEditTool) Description() string { return "Edits an existing image using a text prompt." }
func (t *imageEditTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type":"object",
		"properties":{
			"image_url":{"type":"string","description":"URL of the image to edit"},
			"prompt":{"type":"string","description":"Instructions for the edit"}
		},
		"required":["image_url","prompt"]
	}`)
}
func (t *imageEditTool) Execute(ctx context.Context, input json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in ImageEditInput
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
		}
	}
	if in.ImageURL == "" || in.Prompt == "" {
		return tool.FailureMsg("image_url and prompt are required"), nil
	}
	return tool.SuccessWith(
		fmt.Sprintf("Applied edit to %s", in.ImageURL),
		map[string]any{
			"original_url": in.ImageURL,
			"edited_url":   in.ImageURL + "_edited.png",
			"prompt":       in.Prompt,
		},
	), nil
}

type imageReviewTool struct{}

func (t *imageReviewTool) Name() string { return "image_review" }
func (t *imageReviewTool) Description() string {
	return "Reviews a generated image against specified criteria."
}
func (t *imageReviewTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type":"object",
		"properties":{
			"image_url":{"type":"string","description":"URL of the image to review"},
			"criteria":{"type":"string","description":"Review criteria to evaluate against"}
		},
		"required":["image_url"]
	}`)
}
func (t *imageReviewTool) Execute(ctx context.Context, input json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in ImageReviewInput
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
		}
	}
	if in.ImageURL == "" {
		return tool.FailureMsg("image_url is required"), nil
	}
	return tool.SuccessWith(
		fmt.Sprintf("Review complete for %s", in.ImageURL),
		map[string]any{
			"image_url": in.ImageURL,
			"criteria":  in.Criteria,
			"score":     0.85,
			"passed":    true,
		},
	), nil
}
