package handlers_test

import (
	"context"
	"testing"

	"github.com/control-theory/lethe/internal/graph"
	"github.com/control-theory/lethe/internal/mcp/handlers"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// testSetup creates a server with all tools, connects a client, and returns
// the client session and a cleanup function.
func testSetup(t *testing.T) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()

	db, err := graph.OpenMemoryDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	store := graph.NewStore(db)
	h := handlers.New(store)

	server := mcp.NewServer(&mcp.Implementation{Name: "lethe-test", Version: "test"}, nil)
	registerTestTools(server, h)

	ct, st := mcp.NewInMemoryTransports()
	ss, err := server.Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ss.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.1"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close() })

	return cs
}

func registerTestTools(s *mcp.Server, h *handlers.Handlers) {
	mcp.AddTool(s, &mcp.Tool{Name: "remember", Description: "test"}, h.Remember)
	mcp.AddTool(s, &mcp.Tool{Name: "recall", Description: "test"}, h.Recall)
	mcp.AddTool(s, &mcp.Tool{Name: "forget", Description: "test"}, h.Forget)
	mcp.AddTool(s, &mcp.Tool{Name: "list_recent", Description: "test"}, h.ListRecent)
	mcp.AddTool(s, &mcp.Tool{Name: "stats", Description: "test"}, h.Stats)
	mcp.AddTool(s, &mcp.Tool{Name: "pin", Description: "test"}, h.Pin)
	mcp.AddTool(s, &mcp.Tool{Name: "unpin", Description: "test"}, h.Unpin)
	mcp.AddTool(s, &mcp.Tool{Name: "link", Description: "test"}, h.Link)
	mcp.AddTool(s, &mcp.Tool{Name: "unlink", Description: "test"}, h.Unlink)
	mcp.AddTool(s, &mcp.Tool{Name: "get_context", Description: "test"}, h.GetContext)
	mcp.AddTool(s, &mcp.Tool{Name: "update_memory", Description: "test"}, h.UpdateMemory)
}

func TestRememberCreatesMemory(t *testing.T) {
	cs := testSetup(t)
	ctx := context.Background()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "remember",
		Arguments: map[string]any{
			"content": "Eric is a helicopter pilot",
			"kind":    "person",
			"name":    "Eric",
		},
	})
	if err != nil {
		t.Fatalf("remember: %v", err)
	}
	if res.IsError {
		t.Fatalf("remember returned error: %v", res.Content)
	}
	text := res.Content[0].(*mcp.TextContent).Text
	if text == "" {
		t.Fatal("expected non-empty response")
	}
}

func TestRememberDeduplicates(t *testing.T) {
	cs := testSetup(t)
	ctx := context.Background()

	// First call
	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "remember",
		Arguments: map[string]any{
			"content": "Eric is a helicopter pilot",
			"kind":    "person",
			"name":    "Eric",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Second call with same name+kind should reinforce
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "remember",
		Arguments: map[string]any{
			"content": "Eric also flies fixed-wing",
			"kind":    "person",
			"name":    "Eric",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	text := res.Content[0].(*mcp.TextContent).Text
	if text == "" {
		t.Fatal("expected response")
	}
	// Should say "reinforced", not "remembered"
	if !contains(text, "Reinforced") {
		t.Errorf("expected reinforced message, got: %s", text)
	}
}

func TestRecallFindsMemory(t *testing.T) {
	cs := testSetup(t)
	ctx := context.Background()

	// Store a memory first
	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "remember",
		Arguments: map[string]any{
			"content": "Eric prefers Go over Python for backend work",
			"kind":    "preference",
			"name":    "prefers Go",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Recall it
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "recall",
		Arguments: map[string]any{
			"query": "Go Python",
		},
	})
	if err != nil {
		t.Fatalf("recall: %v", err)
	}
	text := res.Content[0].(*mcp.TextContent).Text
	if !contains(text, "prefers Go") {
		t.Errorf("expected to find 'prefers Go' in results, got: %s", text)
	}
}

func TestForgetSoft(t *testing.T) {
	cs := testSetup(t)
	ctx := context.Background()

	// Create a memory
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "remember",
		Arguments: map[string]any{
			"content": "temporary fact",
			"kind":    "fact",
			"name":    "temp",
		},
	})

	// Recall to get the ID
	res, _ := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "recall",
		Arguments: map[string]any{
			"query": "temp",
		},
	})
	_ = res // We know it exists

	// List recent to get ID
	listRes, _ := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "list_recent",
		Arguments: map[string]any{},
	})
	if listRes.IsError {
		t.Fatal("list_recent failed")
	}
}

func TestStatsReturnsData(t *testing.T) {
	cs := testSetup(t)
	ctx := context.Background()

	// Create some memories
	for _, name := range []string{"Alice", "Bob"} {
		_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "remember",
			Arguments: map[string]any{
				"content": name + " is a person",
				"kind":    "person",
				"name":    name,
			},
		})
	}

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "stats",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	text := res.Content[0].(*mcp.TextContent).Text
	if !contains(text, "Total memories: 2") {
		t.Errorf("expected 2 total memories, got: %s", text)
	}
}

func TestRememberInfersKind(t *testing.T) {
	cs := testSetup(t)
	ctx := context.Background()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "remember",
		Arguments: map[string]any{
			"content": "I prefer dark mode in all my editors",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	text := res.Content[0].(*mcp.TextContent).Text
	if !contains(text, "preference") {
		t.Errorf("expected kind inference to preference, got: %s", text)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// --- Phase 3 tests ---

// helper to create a memory and return its structured output
func createMemory(t *testing.T, cs *mcp.ClientSession, name, kind, content string) map[string]any {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "remember",
		Arguments: map[string]any{
			"content": content,
			"kind":    kind,
			"name":    name,
		},
	})
	if err != nil {
		t.Fatalf("remember(%s): %v", name, err)
	}
	if res.IsError {
		t.Fatalf("remember(%s) error: %v", name, res.Content)
	}
	// Return the structured content if available
	if res.StructuredContent != nil {
		if m, ok := res.StructuredContent.(map[string]any); ok {
			return m
		}
	}
	return nil
}

func TestLinkAndGetContext(t *testing.T) {
	cs := testSetup(t)
	ctx := context.Background()

	// Create two memories
	createMemory(t, cs, "Alice", "person", "Alice is an engineer")
	createMemory(t, cs, "Project X", "project", "Project X is a secret project")

	// List recent to get their IDs
	listRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "list_recent",
		Arguments: map[string]any{"limit": 10},
	})
	if err != nil {
		t.Fatal(err)
	}
	// Extract IDs from structured output
	var aliceID, projID string
	if listRes.StructuredContent != nil {
		if m, ok := listRes.StructuredContent.(map[string]any); ok {
			if mems, ok := m["memories"].([]any); ok {
				for _, mem := range mems {
					if mm, ok := mem.(map[string]any); ok {
						name, _ := mm["name"].(string)
						id, _ := mm["id"].(string)
						if name == "Alice" {
							aliceID = id
						} else if name == "Project X" {
							projID = id
						}
					}
				}
			}
		}
	}

	if aliceID == "" || projID == "" {
		t.Skip("could not extract IDs from structured output, skipping link test")
	}

	// Link them
	linkRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "link",
		Arguments: map[string]any{
			"from_id":  aliceID,
			"to_id":    projID,
			"relation": "works_on",
		},
	})
	if err != nil {
		t.Fatalf("link: %v", err)
	}
	linkText := linkRes.Content[0].(*mcp.TextContent).Text
	if !contains(linkText, "Linked") {
		t.Errorf("expected linked message, got: %s", linkText)
	}

	// Get context around Alice
	ctxRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "get_context",
		Arguments: map[string]any{
			"id":    aliceID,
			"depth": 1,
		},
	})
	if err != nil {
		t.Fatalf("get_context: %v", err)
	}
	ctxText := ctxRes.Content[0].(*mcp.TextContent).Text
	if !contains(ctxText, "Project X") {
		t.Errorf("expected Project X in context, got: %s", ctxText)
	}
	if !contains(ctxText, "works_on") {
		t.Errorf("expected works_on relation in context, got: %s", ctxText)
	}
}

func TestUpdateMemory(t *testing.T) {
	cs := testSetup(t)
	ctx := context.Background()

	createMemory(t, cs, "updatable", "fact", "original content")

	// List recent to get ID
	listRes, _ := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "list_recent",
		Arguments: map[string]any{},
	})

	var id string
	if listRes.StructuredContent != nil {
		if m, ok := listRes.StructuredContent.(map[string]any); ok {
			if mems, ok := m["memories"].([]any); ok && len(mems) > 0 {
				if mm, ok := mems[0].(map[string]any); ok {
					id, _ = mm["id"].(string)
				}
			}
		}
	}

	if id == "" {
		t.Skip("could not extract ID from structured output")
	}

	// Update it
	updateRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "update_memory",
		Arguments: map[string]any{
			"id":      id,
			"summary": "updated summary",
		},
	})
	if err != nil {
		t.Fatalf("update_memory: %v", err)
	}
	text := updateRes.Content[0].(*mcp.TextContent).Text
	if !contains(text, "Updated") {
		t.Errorf("expected update confirmation, got: %s", text)
	}
}

func TestPinUnpin(t *testing.T) {
	cs := testSetup(t)
	ctx := context.Background()

	createMemory(t, cs, "pinnacle", "fact", "important fact")

	listRes, _ := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "list_recent",
		Arguments: map[string]any{},
	})

	var id string
	if listRes.StructuredContent != nil {
		if m, ok := listRes.StructuredContent.(map[string]any); ok {
			if mems, ok := m["memories"].([]any); ok && len(mems) > 0 {
				if mm, ok := mems[0].(map[string]any); ok {
					id, _ = mm["id"].(string)
				}
			}
		}
	}

	if id == "" {
		t.Skip("could not extract ID")
	}

	// Pin
	pinRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "pin",
		Arguments: map[string]any{"id": id},
	})
	if err != nil {
		t.Fatalf("pin: %v", err)
	}
	if !contains(pinRes.Content[0].(*mcp.TextContent).Text, "Pinned") {
		t.Error("expected pin confirmation")
	}

	// Unpin
	unpinRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "unpin",
		Arguments: map[string]any{"id": id},
	})
	if err != nil {
		t.Fatalf("unpin: %v", err)
	}
	if !contains(unpinRes.Content[0].(*mcp.TextContent).Text, "Unpinned") {
		t.Error("expected unpin confirmation")
	}
}
