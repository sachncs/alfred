// Package tool defines the base Tool interface that every concrete tool
// (filesystem, search, execution, MCP-backed) implements.
//
// The interface is deliberately small — Name/Description/Schema for
// discovery, Execute for invocation. Context and Result live in their
// own files to keep the core type readable.
package tool

import (
	"context"
	"encoding/json"
)

// Tool is the contract every concrete tool must satisfy.
//
// Implementations MUST be safe for concurrent use across goroutines;
// the runtime dispatches tools in parallel where the model permits.
type Tool interface {
	// Name returns the stable, machine-readable identifier.
	// It must be unique within a server/runner and should not change
	// across versions — rename = new tool.
	Name() string

	// Description returns a short human-readable summary of what the tool
	// does, suitable for inclusion in a model prompt.
	Description() string

	// Schema returns the JSON Schema (draft-07 or 2020-12) describing the
	// tool's input. Returned as raw JSON so the runtime can pass it
	// through to the model without re-marshalling.
	Schema() json.RawMessage

	// Execute runs the tool with the given input. The returned Result is
	// never nil on a nil error. Errors are wrapped with %w in Result.Error
	// when applicable.
	Execute(ctx context.Context, input json.RawMessage, tc *Context) (*Result, error)
}
