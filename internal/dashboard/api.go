package dashboard

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/control-theory/lethe/internal/graph"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (s *Server) handleListNodes(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}
	kind := r.URL.Query().Get("kind")
	includeFaded := r.URL.Query().Get("include_faded") == "true"
	query := r.URL.Query().Get("q")

	var nodes []graph.Node
	var err error

	if query != "" {
		var kindFilter *graph.NodeKind
		if kind != "" {
			k := graph.NodeKind(kind)
			kindFilter = &k
		}
		nodes, err = s.store.Search(r.Context(), query, graph.SearchOpts{
			Kind: kindFilter, Limit: limit, IncludeFaded: includeFaded,
		})
	} else {
		var kindFilter *graph.NodeKind
		if kind != "" {
			k := graph.NodeKind(kind)
			kindFilter = &k
		}
		if includeFaded {
			nodes, err = s.store.ListAll(r.Context(), limit, kindFilter)
		} else {
			nodes, err = s.store.ListRecent(r.Context(), limit, kindFilter)
		}
	}

	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, nodes)
}

func (s *Server) handleGetNode(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	node, err := s.store.GetNode(r.Context(), id)
	if err == sql.ErrNoRows {
		writeError(w, 404, "not found")
		return
	}
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}

	// Also fetch edges
	edges, _ := s.store.EdgesForNode(r.Context(), id)

	writeJSON(w, 200, map[string]any{
		"node":  node,
		"edges": edges,
	})
}

func (s *Server) handleUpdateNode(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var patch graph.NodePatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, 400, "invalid JSON")
		return
	}
	if err := s.store.UpdateNode(r.Context(), id, patch); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "updated"})
}

func (s *Server) handleDeleteNode(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	hard := r.URL.Query().Get("hard") == "true"
	var err error
	if hard {
		err = s.store.DeleteNode(r.Context(), id)
	} else {
		err = s.store.SoftForget(r.Context(), id)
	}
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

func (s *Server) handlePinNode(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.store.SetPinned(r.Context(), id, true); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "pinned"})
}

func (s *Server) handleUnpinNode(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.store.SetPinned(r.Context(), id, false); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "unpinned"})
}

func (s *Server) handleNodeEdges(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	edges, err := s.store.EdgesForNode(r.Context(), id)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, edges)
}

func (s *Server) handleListEdges(w http.ResponseWriter, r *http.Request) {
	// Simple: list all edges (paginated)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 100
	}
	edges, err := s.store.ListEdges(r.Context(), limit)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, edges)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.store.Stats(r.Context())
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, stats)
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}
	events, err := s.store.ListEvents(r.Context(), limit)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, events)
}

func (s *Server) handleDeleteEdge(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.store.DeleteEdge(r.Context(), id); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

func (s *Server) handleGraph(w http.ResponseWriter, r *http.Request) {
	// Return all nodes and edges for cytoscape rendering
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 200
	}
	nodes, err := s.store.ListRecent(r.Context(), limit, nil)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	edges, err := s.store.ListEdges(r.Context(), 500)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{
		"nodes": nodes,
		"edges": edges,
	})
}
