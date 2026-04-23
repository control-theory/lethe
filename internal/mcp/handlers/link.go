package handlers

import (
	"context"
	"fmt"

	"github.com/control-theory/lethe/internal/graph"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LinkInput is the input for the link tool.
type LinkInput struct {
	FromID   string  `json:"from_id"           jsonschema:"ID of the source memory."`
	ToID     string  `json:"to_id"             jsonschema:"ID of the destination memory."`
	Relation string  `json:"relation"          jsonschema:"Relation type: knows, works_on, works_at, located_in, related_to, part_of, caused, mentioned, prefers, opposes, participated_in, created, depends_on."`
	Weight   float64 `json:"weight,omitempty"  jsonschema:"Edge strength from 0 to 1. Default 1.0."`
}

// LinkOutput is returned by the link tool.
type LinkOutput struct {
	EdgeID   string `json:"edge_id"`
	FromID   string `json:"from_id"`
	ToID     string `json:"to_id"`
	Relation string `json:"relation"`
}

// UnlinkInput is the input for the unlink tool.
type UnlinkInput struct {
	EdgeID string `json:"edge_id" jsonschema:"The ID of the edge to remove."`
}

// Link creates a relation between two memories.
func (h *Handlers) Link(ctx context.Context, req *mcp.CallToolRequest, in LinkInput) (*mcp.CallToolResult, LinkOutput, error) {
	if in.FromID == "" || in.ToID == "" {
		return errResult("from_id and to_id are required"), LinkOutput{}, nil
	}
	if in.Relation == "" {
		in.Relation = "related_to"
	}

	// Verify both nodes exist
	from, err := h.Store.GetNode(ctx, in.FromID)
	if err != nil {
		return errResult(fmt.Sprintf("source memory not found: %s", in.FromID)), LinkOutput{}, nil
	}
	to, err := h.Store.GetNode(ctx, in.ToID)
	if err != nil {
		return errResult(fmt.Sprintf("destination memory not found: %s", in.ToID)), LinkOutput{}, nil
	}

	e := &graph.Edge{
		SrcID:    in.FromID,
		DstID:    in.ToID,
		Relation: graph.RelationType(in.Relation),
		Weight:   in.Weight,
	}
	if err := h.Store.CreateEdge(ctx, e); err != nil {
		return nil, LinkOutput{}, fmt.Errorf("create edge: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Linked: %s -[%s]-> %s", from.Name, in.Relation, to.Name)},
		},
	}, LinkOutput{EdgeID: e.ID, FromID: in.FromID, ToID: in.ToID, Relation: in.Relation}, nil
}

// Unlink removes a relation between memories.
func (h *Handlers) Unlink(ctx context.Context, req *mcp.CallToolRequest, in UnlinkInput) (*mcp.CallToolResult, ForgetOutput, error) {
	if in.EdgeID == "" {
		return errResult("edge_id is required"), ForgetOutput{}, nil
	}

	if err := h.Store.DeleteEdge(ctx, in.EdgeID); err != nil {
		return errResult(fmt.Sprintf("edge not found: %s", in.EdgeID)), ForgetOutput{}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Removed connection: %s", in.EdgeID)},
		},
	}, ForgetOutput{ID: in.EdgeID, Action: "unlinked"}, nil
}
