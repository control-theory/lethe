package dashboard

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/control-theory/lethe/internal/graph"
)

// Server is the localhost-only dashboard HTTP server.
type Server struct {
	store    *graph.Store
	mux      *http.ServeMux
	listener net.Listener
	port     int
}

// New creates a new dashboard server.
func New(store *graph.Store) *Server {
	s := &Server{store: store, mux: http.NewServeMux()}
	s.registerRoutes()
	return s
}

// Start begins listening on the given port. Pass 0 for auto-assign.
// Returns the actual port.
func (s *Server) Start(port int) (int, error) {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return 0, err
	}
	s.listener = ln
	s.port = ln.Addr().(*net.TCPAddr).Port

	go func() {
		srv := &http.Server{Handler: s.guardHandler(s.mux)}
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			slog.Error("dashboard server error", "error", err)
		}
	}()

	slog.Info("dashboard started", "url", fmt.Sprintf("http://127.0.0.1:%d", s.port))
	return s.port, nil
}

// Port returns the port the server is listening on.
func (s *Server) Port() int {
	return s.port
}

// Stop shuts down the dashboard server.
func (s *Server) Stop(ctx context.Context) error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// guardHandler rejects requests not from localhost (DNS rebinding protection).
func (s *Server) guardHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if idx := strings.LastIndex(host, ":"); idx > 0 {
			host = host[:idx]
		}
		if host != "127.0.0.1" && host != "localhost" && host != "::1" {
			http.Error(w, "forbidden: dashboard is localhost-only", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) registerRoutes() {
	// API routes
	s.mux.HandleFunc("GET /api/nodes", s.handleListNodes)
	s.mux.HandleFunc("GET /api/nodes/{id}", s.handleGetNode)
	s.mux.HandleFunc("PATCH /api/nodes/{id}", s.handleUpdateNode)
	s.mux.HandleFunc("DELETE /api/nodes/{id}", s.handleDeleteNode)
	s.mux.HandleFunc("POST /api/nodes/{id}/pin", s.handlePinNode)
	s.mux.HandleFunc("DELETE /api/nodes/{id}/pin", s.handleUnpinNode)
	s.mux.HandleFunc("GET /api/nodes/{id}/edges", s.handleNodeEdges)
	s.mux.HandleFunc("GET /api/edges", s.handleListEdges)
	s.mux.HandleFunc("DELETE /api/edges/{id}", s.handleDeleteEdge)
	s.mux.HandleFunc("GET /api/stats", s.handleStats)
	s.mux.HandleFunc("GET /api/events", s.handleEvents)
	s.mux.HandleFunc("GET /api/graph", s.handleGraph)

	// Frontend
	s.mux.HandleFunc("GET /", s.handleIndex)
}
