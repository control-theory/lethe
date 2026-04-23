package handlers

import (
	"context"
	"fmt"

	"github.com/control-theory/lethe/internal/graph"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// UpdateMemoryInput is the input for the update_memory tool.
type UpdateMemoryInput struct {
	ID      string   `json:"id"                jsonschema:"The ID of the memory to update."`
	Name    string   `json:"name,omitempty"    jsonschema:"New name for the memory."`
	Summary string   `json:"summary,omitempty" jsonschema:"New summary."`
	Body    string   `json:"body,omitempty"    jsonschema:"New body content."`
	Tags    []string `json:"tags,omitempty"    jsonschema:"New tags (replaces existing)."`
}

// UpdateMemoryOutput is returned by the update_memory tool.
type UpdateMemoryOutput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// UpdateMemory edits an existing memory.
func (h *Handlers) UpdateMemory(ctx context.Context, req *mcp.CallToolRequest, in UpdateMemoryInput) (*mcp.CallToolResult, UpdateMemoryOutput, error) {
	if in.ID == "" {
		return errResult("id is required"), UpdateMemoryOutput{}, nil
	}

	// Verify existence
	existing, err := h.Store.GetNode(ctx, in.ID)
	if err != nil {
		return errResult(fmt.Sprintf("memory not found: %s", in.ID)), UpdateMemoryOutput{}, nil
	}

	patch := graph.NodePatch{}
	if in.Name != "" {
		patch.Name = &in.Name
	}
	if in.Summary != "" {
		patch.Summary = &in.Summary
	}
	if in.Body != "" {
		patch.Body = &in.Body
	}
	if in.Tags != nil {
		patch.Tags = &in.Tags
	}

	if err := h.Store.UpdateNode(ctx, in.ID, patch); err != nil {
		return nil, UpdateMemoryOutput{}, fmt.Errorf("update: %w", err)
	}

	name := existing.Name
	if in.Name != "" {
		name = in.Name
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Updated: %s", name)},
		},
	}, UpdateMemoryOutput{ID: in.ID, Name: name}, nil
}
