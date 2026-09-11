package visualdocument

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sachncs/alfred/internal/mcp/worker"
	"github.com/sachncs/alfred/internal/tool"
)

type VisualOpenInput struct {
	DocumentURL string `json:"document_url"`
}

type VisualEditInput struct {
	DocumentURL  string `json:"document_url"`
	Instructions string `json:"instructions"`
}

type VisualAcceptInput struct {
	DocumentURL string `json:"document_url"`
	RevisionID  string `json:"revision_id"`
}

type VisualRejectInput struct {
	DocumentURL string `json:"document_url"`
	RevisionID  string `json:"revision_id"`
	Reason      string `json:"reason,omitempty"`
}

type VisualDocumentServer struct {
	*worker.WorkerServer
}

func NewVisualDocumentServer() *worker.WorkerServer {
	return worker.NewWorkerServer("visual-document-worker", []tool.Tool{
		&visualOpenTool{},
		&visualEditTool{},
		&visualAcceptTool{},
		&visualRejectTool{},
	})
}

type visualOpenTool struct{}

func (t *visualOpenTool) Name() string { return "visual_open" }
func (t *visualOpenTool) Description() string {
	return "Opens a visual document for viewing and editing."
}
func (t *visualOpenTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type":"object",
		"properties":{
			"document_url":{"type":"string","description":"URL or path of the visual document to open"}
		},
		"required":["document_url"]
	}`)
}
func (t *visualOpenTool) Execute(ctx context.Context, input json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in VisualOpenInput
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
		}
	}
	if in.DocumentURL == "" {
		return tool.FailureMsg("document_url is required"), nil
	}
	return tool.SuccessWith(
		fmt.Sprintf("Opened document %s", in.DocumentURL),
		map[string]any{
			"document_url": in.DocumentURL,
			"status":       "opened",
			"revision":     "latest",
		},
	), nil
}

type visualEditTool struct{}

func (t *visualEditTool) Name() string { return "visual_edit" }
func (t *visualEditTool) Description() string {
	return "Edits a visual document with provided instructions."
}
func (t *visualEditTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type":"object",
		"properties":{
			"document_url":{"type":"string","description":"URL or path of the visual document to edit"},
			"instructions":{"type":"string","description":"Edit instructions"}
		},
		"required":["document_url","instructions"]
	}`)
}
func (t *visualEditTool) Execute(ctx context.Context, input json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in VisualEditInput
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
		}
	}
	if in.DocumentURL == "" || in.Instructions == "" {
		return tool.FailureMsg("document_url and instructions are required"), nil
	}
	return tool.SuccessWith(
		fmt.Sprintf("Applied edit to %s", in.DocumentURL),
		map[string]any{
			"document_url": in.DocumentURL,
			"instructions": in.Instructions,
			"revision_id":  "rev_001",
			"status":       "edited",
		},
	), nil
}

type visualAcceptTool struct{}

func (t *visualAcceptTool) Name() string        { return "visual_accept" }
func (t *visualAcceptTool) Description() string { return "Accepts a revision of a visual document." }
func (t *visualAcceptTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type":"object",
		"properties":{
			"document_url":{"type":"string","description":"URL or path of the visual document"},
			"revision_id":{"type":"string","description":"ID of the revision to accept"}
		},
		"required":["document_url","revision_id"]
	}`)
}
func (t *visualAcceptTool) Execute(ctx context.Context, input json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in VisualAcceptInput
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
		}
	}
	if in.DocumentURL == "" || in.RevisionID == "" {
		return tool.FailureMsg("document_url and revision_id are required"), nil
	}
	return tool.SuccessWith(
		fmt.Sprintf("Accepted revision %s", in.RevisionID),
		map[string]any{
			"document_url": in.DocumentURL,
			"revision_id":  in.RevisionID,
			"status":       "accepted",
		},
	), nil
}

type visualRejectTool struct{}

func (t *visualRejectTool) Name() string        { return "visual_reject" }
func (t *visualRejectTool) Description() string { return "Rejects a revision of a visual document." }
func (t *visualRejectTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type":"object",
		"properties":{
			"document_url":{"type":"string","description":"URL or path of the visual document"},
			"revision_id":{"type":"string","description":"ID of the revision to reject"},
			"reason":{"type":"string","description":"Reason for rejection"}
		},
		"required":["document_url","revision_id"]
	}`)
}
func (t *visualRejectTool) Execute(ctx context.Context, input json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in VisualRejectInput
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
		}
	}
	if in.DocumentURL == "" || in.RevisionID == "" {
		return tool.FailureMsg("document_url and revision_id are required"), nil
	}
	return tool.SuccessWith(
		fmt.Sprintf("Rejected revision %s", in.RevisionID),
		map[string]any{
			"document_url": in.DocumentURL,
			"revision_id":  in.RevisionID,
			"reason":       in.Reason,
			"status":       "rejected",
		},
	), nil
}
