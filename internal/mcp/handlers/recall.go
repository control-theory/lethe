package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/control-theory/lethe/internal/graph"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Recall searches stored memories by query.
func (h *Handlers) Recall(ctx context.Context, req *mcp.CallToolRequest, in RecallInput) (*mcp.CallToolResult, RecallOutput, error) {
	if strings.TrimSpace(in.Query) == "" {
		return errResult("query is required"), RecallOutput{}, nil
	}

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

	nodes, err := h.Store.Search(ctx, in.Query, graph.SearchOpts{
		Kind:         kindFilter,
		Limit:        limit,
		IncludeFaded: in.IncludeFaded,
	})
	if err != nil {
		return nil, RecallOutput{}, fmt.Errorf("search: %w", err)
	}

	// Reinforce each returned node and build summaries
	summaries := make([]NodeSummary, 0, len(nodes))
	for _, n := range nodes {
		_ = h.Store.Reinforce(ctx, n.ID, h.Store.ReinforceBump, h.Store.StrengthCap)
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

	// Format text response
	var text string
	if len(summaries) == 0 {
		text = "No memories found matching your query."
	} else {
		var sb strings.Builder
		fmt.Fprintf(&sb, "Found %d memories:\n", len(summaries))
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
