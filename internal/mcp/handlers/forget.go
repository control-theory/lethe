package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Forget soft-forgets or hard-deletes a memory.
func (h *Handlers) Forget(ctx context.Context, req *mcp.CallToolRequest, in ForgetInput) (*mcp.CallToolResult, ForgetOutput, error) {
	if in.ID == "" {
		return errResult("id is required"), ForgetOutput{}, nil
	}

	// Verify the node exists
	node, err := h.Store.GetNode(ctx, in.ID)
	if err != nil {
		return errResult(fmt.Sprintf("memory not found: %s", in.ID)), ForgetOutput{}, nil
	}

	if in.Hard {
		if err := h.Store.DeleteNode(ctx, in.ID); err != nil {
			return nil, ForgetOutput{}, fmt.Errorf("hard delete: %w", err)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Permanently deleted: %s", node.Name)},
			},
		}, ForgetOutput{ID: in.ID, Action: "deleted"}, nil
	}

	if err := h.Store.SoftForget(ctx, in.ID); err != nil {
		return nil, ForgetOutput{}, fmt.Errorf("soft forget: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Soft-forgotten: %s (can be recovered from dashboard)", node.Name)},
		},
	}, ForgetOutput{ID: in.ID, Action: "soft_forgotten"}, nil
}
