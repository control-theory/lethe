package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Stats returns aggregate statistics about the memory graph.
func (h *Handlers) Stats(ctx context.Context, req *mcp.CallToolRequest, in StatsInput) (*mcp.CallToolResult, StatsOutput, error) {
	gs, err := h.Store.Stats(ctx)
	if err != nil {
		return nil, StatsOutput{}, fmt.Errorf("stats: %w", err)
	}

	out := StatsOutput{
		TotalNodes:  gs.TotalNodes,
		TotalEdges:  gs.TotalEdges,
		ByKind:      gs.ByKind,
		PinnedCount: gs.PinnedCount,
		FadingCount: gs.FadingCount,
		RecentCount: gs.RecentCount,
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Memory stats:\n")
	fmt.Fprintf(&sb, "- Total memories: %d\n", gs.TotalNodes)
	fmt.Fprintf(&sb, "- Total connections: %d\n", gs.TotalEdges)
	fmt.Fprintf(&sb, "- Pinned: %d\n", gs.PinnedCount)
	fmt.Fprintf(&sb, "- Fading: %d\n", gs.FadingCount)
	fmt.Fprintf(&sb, "- Recent (24h): %d\n", gs.RecentCount)
	if len(gs.ByKind) > 0 {
		sb.WriteString("- By type: ")
		parts := make([]string, 0, len(gs.ByKind))
		for k, v := range gs.ByKind {
			parts = append(parts, fmt.Sprintf("%s=%d", k, v))
		}
		sb.WriteString(strings.Join(parts, ", "))
		sb.WriteString("\n")
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
	}, out, nil
}
