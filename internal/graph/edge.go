package graph

import (
	"context"
	"database/sql"
	"time"

	"github.com/control-theory/lethe/internal/events"
	"github.com/oklog/ulid/v2"
)

// CreateEdge inserts a new edge into the graph.
func (s *Store) CreateEdge(ctx context.Context, e *Edge) error {
	now := time.Now().UnixMilli()
	e.ID = ulid.Make().String()
	e.CreatedAt = now
	e.LastAccessedAt = now
	if e.Weight == 0 {
		e.Weight = 1.0
	}
	if e.Attributes == nil {
		e.Attributes = map[string]any{}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO edges (id, src_id, dst_id, relation, custom_relation, weight, attributes,
		 created_at, last_accessed_at, access_count)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.SrcID, e.DstID, string(e.Relation), nullStr(e.CustomRelation),
		e.Weight, marshalAttrs(e.Attributes), e.CreatedAt, e.LastAccessedAt, e.AccessCount,
	)
	if err != nil {
		return err
	}

	if err := events.Log(ctx, tx, "create_edge", e.ID, "claude", nil); err != nil {
		return err
	}

	return tx.Commit()
}

// GetEdge retrieves an edge by ID.
func (s *Store) GetEdge(ctx context.Context, id string) (*Edge, error) {
	var e Edge
	var customRel sql.NullString
	var attrsJSON string

	err := s.db.QueryRowContext(ctx,
		`SELECT id, src_id, dst_id, relation, custom_relation, weight, attributes,
		 created_at, last_accessed_at, access_count
		 FROM edges WHERE id = ?`, id).Scan(
		&e.ID, &e.SrcID, &e.DstID, &e.Relation, &customRel, &e.Weight, &attrsJSON,
		&e.CreatedAt, &e.LastAccessedAt, &e.AccessCount,
	)
	if err != nil {
		return nil, err
	}
	e.CustomRelation = customRel.String
	e.Attributes = unmarshalAttrs(attrsJSON)
	return &e, nil
}

// DeleteEdge removes an edge by ID.
func (s *Store) DeleteEdge(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `DELETE FROM edges WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}

	if err := events.Log(ctx, tx, "delete_edge", id, "claude", nil); err != nil {
		return err
	}

	return tx.Commit()
}

// EdgesForNode returns all edges where the node is either source or destination.
func (s *Store) EdgesForNode(ctx context.Context, nodeID string) ([]Edge, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, src_id, dst_id, relation, custom_relation, weight, attributes,
		 created_at, last_accessed_at, access_count
		 FROM edges WHERE src_id = ? OR dst_id = ?`, nodeID, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []Edge
	for rows.Next() {
		var e Edge
		var customRel sql.NullString
		var attrsJSON string
		if err := rows.Scan(
			&e.ID, &e.SrcID, &e.DstID, &e.Relation, &customRel, &e.Weight, &attrsJSON,
			&e.CreatedAt, &e.LastAccessedAt, &e.AccessCount,
		); err != nil {
			return nil, err
		}
		e.CustomRelation = customRel.String
		e.Attributes = unmarshalAttrs(attrsJSON)
		edges = append(edges, e)
	}
	if edges == nil {
		edges = []Edge{}
	}
	return edges, rows.Err()
}

// ListEdges returns all edges up to limit.
func (s *Store) ListEdges(ctx context.Context, limit int) ([]Edge, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, src_id, dst_id, relation, custom_relation, weight, attributes,
		 created_at, last_accessed_at, access_count
		 FROM edges ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []Edge
	for rows.Next() {
		var e Edge
		var customRel sql.NullString
		var attrsJSON string
		if err := rows.Scan(
			&e.ID, &e.SrcID, &e.DstID, &e.Relation, &customRel, &e.Weight, &attrsJSON,
			&e.CreatedAt, &e.LastAccessedAt, &e.AccessCount,
		); err != nil {
			return nil, err
		}
		e.CustomRelation = customRel.String
		e.Attributes = unmarshalAttrs(attrsJSON)
		edges = append(edges, e)
	}
	if edges == nil {
		edges = []Edge{}
	}
	return edges, rows.Err()
}
