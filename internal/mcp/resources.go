package mcpserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/control-theory/lethe/internal/graph"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerResources(s *mcp.Server, store *graph.Store) {
	// lethe://context/active — top-N active memories for session start
	s.AddResource(&mcp.Resource{
		URI:         "lethe://context/active",
		Name:        "Active memories",
		Description: "The top 20 most active memories, suitable for session-start context.",
		MIMEType:    "text/markdown",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return activeContextResource(ctx, store)
	})

	// lethe://stats/summary — compact stats
	s.AddResource(&mcp.Resource{
		URI:         "lethe://stats/summary",
		Name:        "Memory stats",
		Description: "Compact summary of the memory system state.",
		MIMEType:    "text/markdown",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return statsSummaryResource(ctx, store)
	})

	// lethe://schema/types — node kinds and edge relations
	s.AddResource(&mcp.Resource{
		URI:         "lethe://schema/types",
		Name:        "Memory schema types",
		Description: "Available node kinds and edge relation types.",
		MIMEType:    "text/markdown",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return schemaTypesResource(ctx, store)
	})
}

func activeContextResource(ctx context.Context, store *graph.Store) (*mcp.ReadResourceResult, error) {
	// Get pinned memories first
	pinned, err := store.ListPinned(ctx, 20)
	if err != nil {
		return nil, err
	}

	// Then get top by strength (most recent/reinforced)
	remaining := 20 - len(pinned)
	var recent []graph.Node
	if remaining > 0 {
		recent, err = store.ListRecent(ctx, remaining, nil)
		if err != nil {
			return nil, err
		}
	}

	var sb strings.Builder
	sb.WriteString("# Active memories\n\n")

	if len(pinned) > 0 {
		sb.WriteString("## Pinned\n")
		for _, n := range pinned {
			fmt.Fprintf(&sb, "- **%s** (%s)", n.Name, n.Kind)
			if n.Summary != "" {
				fmt.Fprintf(&sb, ": %s", n.Summary)
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	if len(recent) > 0 {
		sb.WriteString("## Recent & reinforced\n")
		for _, n := range recent {
			// Skip if already in pinned
			skip := false
			for _, p := range pinned {
				if p.ID == n.ID {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
			fmt.Fprintf(&sb, "- **%s** (%s)", n.Name, n.Kind)
			if n.Summary != "" {
				fmt.Fprintf(&sb, ": %s", n.Summary)
			}
			sb.WriteString("\n")
		}
	}

	if len(pinned) == 0 && len(recent) == 0 {
		sb.WriteString("No memories stored yet.\n")
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      "lethe://context/active",
			MIMEType: "text/markdown",
			Text:     sb.String(),
		}},
	}, nil
}

func statsSummaryResource(ctx context.Context, store *graph.Store) (*mcp.ReadResourceResult, error) {
	stats, err := store.Stats(ctx)
	if err != nil {
		return nil, err
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "# Memory stats\n\n")
	fmt.Fprintf(&sb, "- Memories: %d\n", stats.TotalNodes)
	fmt.Fprintf(&sb, "- Connections: %d\n", stats.TotalEdges)
	fmt.Fprintf(&sb, "- Pinned: %d\n", stats.PinnedCount)
	fmt.Fprintf(&sb, "- Fading: %d\n", stats.FadingCount)
	fmt.Fprintf(&sb, "- Recent (24h): %d\n", stats.RecentCount)

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      "lethe://stats/summary",
			MIMEType: "text/markdown",
			Text:     sb.String(),
		}},
	}, nil
}

func schemaTypesResource(ctx context.Context, store *graph.Store) (*mcp.ReadResourceResult, error) {
	var sb strings.Builder
	sb.WriteString("# Memory schema\n\n")
	sb.WriteString("## Node kinds\n")
	for _, k := range []string{
		"person", "place", "project", "organization", "concept",
		"event", "artifact", "preference", "fact", "note", "custom",
	} {
		fmt.Fprintf(&sb, "- `%s`\n", k)
	}

	sb.WriteString("\n## Edge relations\n")
	for _, r := range []string{
		"knows", "works_on", "works_at", "located_in", "related_to",
		"part_of", "caused", "mentioned", "prefers", "opposes",
		"participated_in", "created", "depends_on", "custom",
	} {
		fmt.Fprintf(&sb, "- `%s`\n", r)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      "lethe://schema/types",
			MIMEType: "text/markdown",
			Text:     sb.String(),
		}},
	}, nil
}
