package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// PinInput is the input for the pin tool.
type PinInput struct {
	ID string `json:"id" jsonschema:"The ID of the memory to pin."`
}

// UnpinInput is the input for the unpin tool.
type UnpinInput struct {
	ID string `json:"id" jsonschema:"The ID of the memory to unpin."`
}

// PinOutput is returned by pin/unpin tools.
type PinOutput struct {
	ID     string `json:"id"`
	Pinned bool   `json:"pinned"`
}

// Pin prevents a memory from decaying.
func (h *Handlers) Pin(ctx context.Context, req *mcp.CallToolRequest, in PinInput) (*mcp.CallToolResult, PinOutput, error) {
	if in.ID == "" {
		return errResult("id is required"), PinOutput{}, nil
	}

	node, err := h.Store.GetNode(ctx, in.ID)
	if err != nil {
		return errResult(fmt.Sprintf("memory not found: %s", in.ID)), PinOutput{}, nil
	}

	if err := h.Store.SetPinned(ctx, in.ID, true); err != nil {
		return nil, PinOutput{}, fmt.Errorf("pin: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Pinned: %s (will never decay)", node.Name)},
		},
	}, PinOutput{ID: in.ID, Pinned: true}, nil
}

// Unpin allows a memory to decay normally again.
func (h *Handlers) Unpin(ctx context.Context, req *mcp.CallToolRequest, in UnpinInput) (*mcp.CallToolResult, PinOutput, error) {
	if in.ID == "" {
		return errResult("id is required"), PinOutput{}, nil
	}

	node, err := h.Store.GetNode(ctx, in.ID)
	if err != nil {
		return errResult(fmt.Sprintf("memory not found: %s", in.ID)), PinOutput{}, nil
	}

	if err := h.Store.SetPinned(ctx, in.ID, false); err != nil {
		return nil, PinOutput{}, fmt.Errorf("unpin: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Unpinned: %s (will decay normally)", node.Name)},
		},
	}, PinOutput{ID: in.ID, Pinned: false}, nil
}
