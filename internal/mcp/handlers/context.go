package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetContextInput is the input for the get_context tool.
type GetContextInput struct {
	ID    string `json:"id"              jsonschema:"The ID of the memory to get context for."`
	Depth int    `json:"depth,omitempty" jsonschema:"How many hops to traverse. Default 1, max 3."`
}

// ContextNode is a node in a context subgraph.
type ContextNode struct {
	ID       string  `json:"id"`
	Kind     string  `json:"kind"`
	Name     string  `json:"name"`
	Summary  string  `json:"summary,omitempty"`
	Strength float64 `json:"strength"`
}

// ContextEdge is an edge in a context subgraph.
type ContextEdge struct {
	ID       string  `json:"id"`
	FromID   string  `json:"from_id"`
	ToID     string  `json:"to_id"`
	Relation string  `json:"relation"`
	Weight   float64 `json:"weight"`
}

// GetContextOutput is the subgraph returned by get_context.
type GetContextOutput struct {
	Nodes []ContextNode `json:"nodes"`
	Edges []ContextEdge `json:"edges"`
}

// GetContext returns the subgraph around a memory.
func (h *Handlers) GetContext(ctx context.Context, req *mcp.CallToolRequest, in GetContextInput) (*mcp.CallToolResult, GetContextOutput, error) {
	if in.ID == "" {
		return errResult("id is required"), GetContextOutput{}, nil
	}
	if in.Depth <= 0 {
		in.Depth = 1
	}

	sg, err := h.Store.Subgraph(ctx, in.ID, in.Depth)
	if err != nil {
		return errResult(fmt.Sprintf("memory not found: %s", in.ID)), GetContextOutput{}, nil
	}

	out := GetContextOutput{
		Nodes: make([]ContextNode, 0, len(sg.Nodes)),
		Edges: make([]ContextEdge, 0, len(sg.Edges)),
	}
	for _, n := range sg.Nodes {
		out.Nodes = append(out.Nodes, ContextNode{
			ID: n.ID, Kind: string(n.Kind), Name: n.Name,
			Summary: n.Summary, Strength: n.Strength,
		})
	}
	for _, e := range sg.Edges {
		out.Edges = append(out.Edges, ContextEdge{
			ID: e.ID, FromID: e.SrcID, ToID: e.DstID,
			Relation: string(e.Relation), Weight: e.Weight,
		})
	}

	// Build text summary
	var sb strings.Builder
	fmt.Fprintf(&sb, "Context for %s (%d nodes, %d edges):\n", in.ID, len(out.Nodes), len(out.Edges))
	for _, n := range out.Nodes {
		fmt.Fprintf(&sb, "- %s (%s)", n.Name, n.Kind)
		if n.Summary != "" {
			fmt.Fprintf(&sb, ": %s", n.Summary)
		}
		sb.WriteString("\n")
	}
	if len(out.Edges) > 0 {
		sb.WriteString("Connections:\n")
		// Build a name lookup
		nameMap := make(map[string]string)
		for _, n := range out.Nodes {
			nameMap[n.ID] = n.Name
		}
		for _, e := range out.Edges {
			fmt.Fprintf(&sb, "- %s -[%s]-> %s\n", nameMap[e.FromID], e.Relation, nameMap[e.ToID])
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
	}, out, nil
}
