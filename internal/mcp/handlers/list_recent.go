package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/control-theory/lethe/internal/graph"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListRecent returns the most recently created or reinforced memories.
func (h *Handlers) ListRecent(ctx context.Context, req *mcp.CallToolRequest, in ListRecentInput) (*mcp.CallToolResult, RecallOutput, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	var kindFilter *graph.NodeKind
	if in.Kind != "" {
		k := graph.NodeKind(in.Kind)
		kindFilter = &k
	}

	nodes, err := h.Store.ListRecent(ctx, limit, kindFilter)
	if err != nil {
		return nil, RecallOutput{}, fmt.Errorf("list recent: %w", err)
	}

	summaries := make([]NodeSummary, 0, len(nodes))
	for _, n := range nodes {
		summaries = append(summaries, NodeSummary{
			ID:       n.ID,
			Kind:     string(n.Kind),
			Name:     n.Name,
			Summary:  n.Summary,
			Strength: n.Strength,
			Tags:     n.Tags,
			Pinned:   n.Pinned,
		})
	}

	out := RecallOutput{
		Memories: summaries,
		Total:    len(summaries),
	}

	var text string
	if len(summaries) == 0 {
		text = "No recent memories."
	} else {
		var sb strings.Builder
		fmt.Fprintf(&sb, "Recent memories (%d):\n", len(summaries))
		for _, m := range summaries {
			fmt.Fprintf(&sb, "- **%s** (%s)", m.Name, m.Kind)
			if m.Summary != "" {
				fmt.Fprintf(&sb, ": %s", m.Summary)
			}
			sb.WriteString("\n")
		}
		text = sb.String()
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}, out, nil
}
