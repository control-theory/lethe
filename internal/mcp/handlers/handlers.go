package handlers

import "github.com/control-theory/lethe/internal/graph"

// Handlers holds shared dependencies for all MCP tool handlers.
type Handlers struct {
	Store *graph.Store
}

// New creates a new Handlers instance.
func New(store *graph.Store) *Handlers {
	return &Handlers{Store: store}
}
