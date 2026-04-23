package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/control-theory/lethe/internal/graph"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Remember stores a new memory or reinforces an existing one.
func (h *Handlers) Remember(ctx context.Context, req *mcp.CallToolRequest, in RememberInput) (*mcp.CallToolResult, RememberOutput, error) {
	if strings.TrimSpace(in.Content) == "" {
		return errResult("content is required"), RememberOutput{}, nil
	}

	// Infer kind if not provided
	kind := graph.NodeKind(in.Kind)
	if in.Kind == "" {
		kind = inferKind(in.Content)
	} else if !graph.ValidNodeKind(in.Kind) {
		kind = graph.KindCustom
	}

	// Derive name if not provided
	name := in.Name
	if name == "" {
		name = deriveName(in.Content)
	}

	// Check for duplicates
	dedup, err := h.Store.CheckDuplicate(ctx, kind, name, in.Content)
	if err != nil {
		return nil, RememberOutput{}, fmt.Errorf("dedup check: %w", err)
	}

	if dedup.IsDuplicate {
		// Reinforce existing node
		if err := h.Store.Reinforce(ctx, dedup.ExistingID, h.Store.ReinforceBump, h.Store.StrengthCap); err != nil {
			return nil, RememberOutput{}, fmt.Errorf("reinforce: %w", err)
		}
		// If there's new content, append it to the body
		existing, err := h.Store.GetNode(ctx, dedup.ExistingID)
		if err != nil {
			return nil, RememberOutput{}, err
		}

		// Merge new content into body if it adds information
		if in.Content != existing.Summary && in.Content != existing.Body {
			newBody := existing.Body
			if newBody != "" {
				newBody += "\n"
			}
			newBody += in.Content
			if err := h.Store.UpdateNode(ctx, dedup.ExistingID, graph.NodePatch{Body: &newBody}); err != nil {
				return nil, RememberOutput{}, err
			}
		}

		out := RememberOutput{
			ID:      existing.ID,
			Name:    existing.Name,
			Kind:    string(existing.Kind),
			Action:  "reinforced",
			Summary: existing.Summary,
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Reinforced existing memory: %s (%s)", existing.Name, existing.Kind)},
			},
		}, out, nil
	}

	// Create new node
	node := &graph.Node{
		Kind:    kind,
		Name:    name,
		Summary: truncate(in.Content, 200),
		Tags:    in.Tags,
	}
	// If content is longer than the summary, store full content in body
	if len(in.Content) > 200 {
		node.Body = in.Content
	}

	if err := h.Store.CreateNode(ctx, node); err != nil {
		return nil, RememberOutput{}, fmt.Errorf("create node: %w", err)
	}

	// Auto-link related_to
	for _, ref := range in.RelatedTo {
		h.linkByReference(ctx, node.ID, ref)
	}

	out := RememberOutput{
		ID:      node.ID,
		Name:    node.Name,
		Kind:    string(node.Kind),
		Action:  "created",
		Summary: node.Summary,
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Remembered: %s (%s)", node.Name, node.Kind)},
		},
	}, out, nil
}

// linkByReference tries to find a node by ID or name and create an edge.
func (h *Handlers) linkByReference(ctx context.Context, fromID, ref string) {
	// Try as ID first
	target, err := h.Store.GetNode(ctx, ref)
	if err != nil {
		// Try as name across all kinds
		for _, kind := range []graph.NodeKind{
			graph.KindPerson, graph.KindProject, graph.KindOrganization,
			graph.KindConcept, graph.KindPreference, graph.KindFact,
			graph.KindNote, graph.KindPlace, graph.KindEvent, graph.KindArtifact,
		} {
			target, err = h.Store.GetNodeByName(ctx, kind, ref)
			if err == nil {
				break
			}
		}
	}
	if target == nil {
		return
	}

	_ = h.Store.CreateEdge(ctx, &graph.Edge{
		SrcID:    fromID,
		DstID:    target.ID,
		Relation: graph.RelRelatedTo,
	})
}

// inferKind guesses the node kind from content using simple heuristics.
func inferKind(content string) graph.NodeKind {
	lower := strings.ToLower(content)
	switch {
	case strings.Contains(lower, "prefer") || strings.Contains(lower, "like ") ||
		strings.Contains(lower, "rather ") || strings.Contains(lower, "favorite"):
		return graph.KindPreference
	case strings.Contains(lower, "name is ") || strings.Contains(lower, "person"):
		return graph.KindPerson
	case strings.Contains(lower, "http://") || strings.Contains(lower, "https://") ||
		strings.Contains(lower, "github.com"):
		return graph.KindArtifact
	case strings.Contains(lower, "meeting") || strings.Contains(lower, "deadline") ||
		strings.Contains(lower, "scheduled"):
		return graph.KindEvent
	case strings.Contains(lower, "project") || strings.Contains(lower, "building") ||
		strings.Contains(lower, "working on"):
		return graph.KindProject
	case strings.Contains(lower, "company") || strings.Contains(lower, "team") ||
		strings.Contains(lower, "organization"):
		return graph.KindOrganization
	default:
		return graph.KindFact
	}
}

// deriveName extracts a short name from content.
func deriveName(content string) string {
	// Take the first sentence or first 50 chars
	if idx := strings.IndexAny(content, ".!?\n"); idx > 0 && idx < 80 {
		return strings.TrimSpace(content[:idx])
	}
	return truncate(content, 50)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	// Try to cut at a word boundary
	cut := strings.LastIndex(s[:maxLen], " ")
	if cut < maxLen/2 {
		cut = maxLen
	}
	return s[:cut]
}

func errResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: true,
	}
}
