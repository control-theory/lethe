package mcpserver_test

import (
	"context"
	"testing"

	mcpserver "github.com/control-theory/lethe/internal/mcp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestPingTool(t *testing.T) {
	ctx := context.Background()
	ct, st := mcp.NewInMemoryTransports()

	s := mcpserver.New("test")

	ss, err := s.Inner().Connect(ctx, st, nil)
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

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "ping",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool(ping) error: %v", err)
	}
	if len(res.Content) == 0 {
		t.Fatal("expected content in response")
	}
	text, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", res.Content[0])
	}
	if text.Text != "pong" {
		t.Errorf("got %q, want %q", text.Text, "pong")
	}
}

func TestListToolsIncludesPing(t *testing.T) {
	ctx := context.Background()
	ct, st := mcp.NewInMemoryTransports()

	s := mcpserver.New("test")

	ss, err := s.Inner().Connect(ctx, st, nil)
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

	res, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools error: %v", err)
	}

	found := false
	for _, tool := range res.Tools {
		if tool.Name == "ping" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ping tool not found in tool list")
	}
}
