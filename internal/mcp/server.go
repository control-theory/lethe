package mcpserver

import (
	"context"

	"github.com/control-theory/lethe/internal/graph"
	"github.com/control-theory/lethe/internal/mcp/handlers"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps the MCP server and manages tool registration.
type Server struct {
	inner *mcp.Server
}

// Option configures the Server.
type Option func(*serverConfig)

type serverConfig struct {
	store *graph.Store
}

// WithStore injects a graph store for the memory tools.
func WithStore(store *graph.Store) Option {
	return func(cfg *serverConfig) {
		cfg.store = store
	}
}

// New creates a new MCP server with all registered tools.
func New(version string, opts ...Option) *Server {
	var cfg serverConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	s := &Server{
		inner: mcp.NewServer(&mcp.Implementation{
			Name:    "lethe",
			Version: version,
		}, nil),
	}

	registerPing(s.inner)

	if cfg.store != nil {
		h := handlers.New(cfg.store)
		registerTools(s.inner, h)
		registerResources(s.inner, cfg.store)
	}

	return s
}

// Run starts the MCP server on stdio transport. Blocks until ctx is cancelled.
func (s *Server) Run(ctx context.Context) error {
	return s.inner.Run(ctx, &mcp.StdioTransport{})
}

// Inner returns the underlying mcp.Server, for testing.
func (s *Server) Inner() *mcp.Server {
	return s.inner
}

type pingInput struct{}

func registerPing(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "ping",
		Description: "Health check. Returns pong.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in pingInput) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "pong"},
			},
		}, nil, nil
	})
}
