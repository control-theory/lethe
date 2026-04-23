package graph

import (
	"context"
	"encoding/json"
)

// ExportData is the full export format for backup/restore.
type ExportData struct {
	Version string  `json:"version"`
	Nodes   []Node  `json:"nodes"`
	Edges   []Edge  `json:"edges"`
}

// Export returns all nodes and edges as a serializable struct.
func (s *Store) Export(ctx context.Context) (*ExportData, error) {
	nodes, err := s.ListAll(ctx, 100000, nil)
	if err != nil {
		return nil, err
	}
	edges, err := s.ListEdges(ctx, 100000)
	if err != nil {
		return nil, err
	}
	return &ExportData{
		Version: "1",
		Nodes:   nodes,
		Edges:   edges,
	}, nil
}

// ExportJSON returns the full export as JSON bytes.
func (s *Store) ExportJSON(ctx context.Context) ([]byte, error) {
	data, err := s.Export(ctx)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(data, "", "  ")
}

// Import loads nodes and edges from an ExportData, skipping duplicates by ID.
func (s *Store) Import(ctx context.Context, data *ExportData) (imported int, err error) {
	for i := range data.Nodes {
		n := &data.Nodes[i]
		// Check if node already exists
		_, err := s.GetNode(ctx, n.ID)
		if err == nil {
			continue // already exists
		}
		// Insert directly with the existing ID
		if err := s.importNode(ctx, n); err != nil {
			return imported, err
		}
		imported++
	}
	for i := range data.Edges {
		e := &data.Edges[i]
		_, err := s.GetEdge(ctx, e.ID)
		if err == nil {
			continue
		}
		if err := s.importEdge(ctx, e); err != nil {
			return imported, err
		}
		imported++
	}
	return imported, nil
}

// importNode inserts a node preserving its original ID and timestamps.
func (s *Store) importNode(ctx context.Context, n *Node) error {
	if n.Tags == nil {
		n.Tags = []string{}
	}
	if n.Attributes == nil {
		n.Attributes = map[string]any{}
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO nodes (id, kind, custom_type, name, summary, body, tags, attributes,
		 created_at, updated_at, last_accessed_at, access_count, strength, pinned)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		n.ID, string(n.Kind), nullStr(n.CustomType), n.Name, nullStr(n.Summary), nullStr(n.Body),
		marshalTags(n.Tags), marshalAttrs(n.Attributes),
		n.CreatedAt, n.UpdatedAt, n.LastAccessedAt, n.AccessCount, n.Strength, boolToInt(n.Pinned),
	)
	return err
}

// importEdge inserts an edge preserving its original ID.
func (s *Store) importEdge(ctx context.Context, e *Edge) error {
	if e.Attributes == nil {
		e.Attributes = map[string]any{}
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO edges (id, src_id, dst_id, relation, custom_relation, weight, attributes,
		 created_at, last_accessed_at, access_count)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.SrcID, e.DstID, string(e.Relation), nullStr(e.CustomRelation),
		e.Weight, marshalAttrs(e.Attributes), e.CreatedAt, e.LastAccessedAt, e.AccessCount,
	)
	return err
}
