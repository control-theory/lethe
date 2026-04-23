package graph

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/control-theory/lethe/internal/events"
	"github.com/oklog/ulid/v2"
)

// CreateNode inserts a new node into the graph. It generates an ID and sets timestamps.
func (s *Store) CreateNode(ctx context.Context, n *Node) error {
	now := time.Now().UnixMilli()
	n.ID = ulid.Make().String()
	n.CreatedAt = now
	n.UpdatedAt = now
	n.LastAccessedAt = now
	if n.AccessCount == 0 {
		n.AccessCount = 1
	}
	if n.Strength == 0 {
		n.Strength = 1.0
	}
	if n.Tags == nil {
		n.Tags = []string{}
	}
	if n.Attributes == nil {
		n.Attributes = map[string]any{}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO nodes (id, kind, custom_type, name, summary, body, tags, attributes,
		 created_at, updated_at, last_accessed_at, access_count, strength, pinned)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		n.ID, string(n.Kind), nullStr(n.CustomType), n.Name, nullStr(n.Summary), nullStr(n.Body),
		marshalTags(n.Tags), marshalAttrs(n.Attributes),
		n.CreatedAt, n.UpdatedAt, n.LastAccessedAt, n.AccessCount, n.Strength, boolToInt(n.Pinned),
	)
	if err != nil {
		return err
	}

	if err := events.Log(ctx, tx, "create_node", n.ID, "claude", nil); err != nil {
		return err
	}

	return tx.Commit()
}

// GetNode retrieves a node by ID with decay-adjusted strength.
func (s *Store) GetNode(ctx context.Context, id string) (*Node, error) {
	n, err := scanNode(s.db.QueryRowContext(ctx,
		`SELECT id, kind, custom_type, name, summary, body, tags, attributes,
		 created_at, updated_at, last_accessed_at, access_count, strength, pinned
		 FROM nodes WHERE id = ?`, id))
	if err != nil {
		return nil, err
	}
	ApplyDecay(n, s.HalfLifeDays)
	return n, nil
}

// GetNodeByName finds a node by kind and case-insensitive name.
func (s *Store) GetNodeByName(ctx context.Context, kind NodeKind, name string) (*Node, error) {
	return scanNode(s.db.QueryRowContext(ctx,
		`SELECT id, kind, custom_type, name, summary, body, tags, attributes,
		 created_at, updated_at, last_accessed_at, access_count, strength, pinned
		 FROM nodes WHERE kind = ? AND LOWER(name) = LOWER(?)`, string(kind), name))
}

// UpdateNode applies a partial patch to an existing node.
func (s *Store) UpdateNode(ctx context.Context, id string, patch NodePatch) error {
	var sets []string
	var args []any

	if patch.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *patch.Name)
	}
	if patch.Summary != nil {
		sets = append(sets, "summary = ?")
		args = append(args, *patch.Summary)
	}
	if patch.Body != nil {
		sets = append(sets, "body = ?")
		args = append(args, *patch.Body)
	}
	if patch.Tags != nil {
		sets = append(sets, "tags = ?")
		args = append(args, marshalTags(*patch.Tags))
	}
	if patch.Attributes != nil {
		sets = append(sets, "attributes = ?")
		args = append(args, marshalAttrs(*patch.Attributes))
	}
	if len(sets) == 0 {
		return nil
	}

	now := time.Now().UnixMilli()
	sets = append(sets, "updated_at = ?")
	args = append(args, now)
	args = append(args, id)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		fmt.Sprintf("UPDATE nodes SET %s WHERE id = ?", strings.Join(sets, ", ")),
		args...,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}

	if err := events.Log(ctx, tx, "update_node", id, "claude", nil); err != nil {
		return err
	}

	return tx.Commit()
}

// SoftForget sets a node's strength to 0 (soft delete).
func (s *Store) SoftForget(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().UnixMilli()
	res, err := tx.ExecContext(ctx,
		`UPDATE nodes SET strength = 0, updated_at = ? WHERE id = ?`, now, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}

	if err := events.Log(ctx, tx, "delete_node", id, "claude", map[string]string{"type": "soft"}); err != nil {
		return err
	}

	return tx.Commit()
}

// DeleteNode permanently removes a node and its edges (via CASCADE).
func (s *Store) DeleteNode(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `DELETE FROM nodes WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}

	if err := events.Log(ctx, tx, "delete_node", id, "claude", map[string]string{"type": "hard"}); err != nil {
		return err
	}

	return tx.Commit()
}

// ListAll returns nodes including faded ones, up to limit.
func (s *Store) ListAll(ctx context.Context, limit int, kind *NodeKind) ([]Node, error) {
	if limit <= 0 {
		limit = 50
	}
	var query string
	var args []any
	if kind != nil {
		query = `SELECT id, kind, custom_type, name, summary, body, tags, attributes,
			created_at, updated_at, last_accessed_at, access_count, strength, pinned
			FROM nodes WHERE kind = ? ORDER BY updated_at DESC LIMIT ?`
		args = []any{string(*kind), limit}
	} else {
		query = `SELECT id, kind, custom_type, name, summary, body, tags, attributes,
			created_at, updated_at, last_accessed_at, access_count, strength, pinned
			FROM nodes ORDER BY updated_at DESC LIMIT ?`
		args = []any{limit}
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	nodes, err := scanNodes(rows)
	if err != nil {
		return nil, err
	}
	s.applyDecayToNodes(nodes)
	return nodes, nil
}

// ListRecent returns the most recently updated active nodes.
func (s *Store) ListRecent(ctx context.Context, limit int, kind *NodeKind) ([]Node, error) {
	if limit <= 0 {
		limit = 10
	}

	var query string
	var args []any

	if kind != nil {
		query = `SELECT id, kind, custom_type, name, summary, body, tags, attributes,
			created_at, updated_at, last_accessed_at, access_count, strength, pinned
			FROM nodes WHERE strength >= 0.1 AND kind = ? ORDER BY updated_at DESC LIMIT ?`
		args = []any{string(*kind), limit}
	} else {
		query = `SELECT id, kind, custom_type, name, summary, body, tags, attributes,
			created_at, updated_at, last_accessed_at, access_count, strength, pinned
			FROM nodes WHERE strength >= 0.1 ORDER BY updated_at DESC LIMIT ?`
		args = []any{limit}
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	nodes, err := scanNodes(rows)
	if err != nil {
		return nil, err
	}
	s.applyDecayToNodes(nodes)
	return nodes, nil
}

// Reinforce computes the current decayed strength, adds bump, persists the result.
func (s *Store) Reinforce(ctx context.Context, id string, bump float64, cap float64) error {
	now := time.Now().UnixMilli()

	// Read stored values
	var storedStrength float64
	var lastAccessedAt int64
	var pinned int
	err := s.db.QueryRowContext(ctx,
		`SELECT strength, last_accessed_at, pinned FROM nodes WHERE id = ?`, id).
		Scan(&storedStrength, &lastAccessedAt, &pinned)
	if err != nil {
		return err
	}

	// Compute decayed strength (skip decay for pinned nodes)
	current := storedStrength
	if pinned == 0 {
		current = CurrentStrength(storedStrength, lastAccessedAt, s.HalfLifeDays)
	}
	newStrength := ReinforceBump(current, bump, cap)

	_, err = s.db.ExecContext(ctx,
		`UPDATE nodes SET
			strength = ?,
			access_count = access_count + 1,
			last_accessed_at = ?,
			updated_at = ?
		 WHERE id = ?`,
		newStrength, now, now, id,
	)
	return err
}

// ListPinned returns all pinned nodes, up to limit.
func (s *Store) ListPinned(ctx context.Context, limit int) ([]Node, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, kind, custom_type, name, summary, body, tags, attributes,
		 created_at, updated_at, last_accessed_at, access_count, strength, pinned
		 FROM nodes WHERE pinned = 1 ORDER BY updated_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNodes(rows)
}

// SetPinned sets or clears the pinned flag on a node.
func (s *Store) SetPinned(ctx context.Context, id string, pinned bool) error {
	now := time.Now().UnixMilli()
	res, err := s.db.ExecContext(ctx,
		`UPDATE nodes SET pinned = ?, updated_at = ? WHERE id = ?`,
		boolToInt(pinned), now, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// MaterializeStrength updates the stored strength column for all non-pinned nodes
// with their current decayed value. Used by housekeeping.
func (s *Store) MaterializeStrength(ctx context.Context) (int64, error) {
	now := time.Now().UnixMilli()
	// Batch update: compute decay in SQL using the exponential formula.
	// SQLite doesn't have exp(), so we do this in Go.
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, strength, last_accessed_at FROM nodes WHERE pinned = 0 AND strength > 0`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type update struct {
		id       string
		strength float64
	}
	var updates []update
	for rows.Next() {
		var id string
		var strength float64
		var lastAccessed int64
		if err := rows.Scan(&id, &strength, &lastAccessed); err != nil {
			return 0, err
		}
		decayed := CurrentStrengthAt(strength, lastAccessed, s.HalfLifeDays, now)
		if decayed != strength {
			updates = append(updates, update{id, decayed})
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	if len(updates) == 0 {
		return 0, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `UPDATE nodes SET strength = ? WHERE id = ?`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	for _, u := range updates {
		if _, err := stmt.ExecContext(ctx, u.strength, u.id); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return int64(len(updates)), nil
}

// scanNode scans a single node from a *sql.Row.
func scanNode(row *sql.Row) (*Node, error) {
	var n Node
	var customType, summary, body sql.NullString
	var tagsJSON, attrsJSON string
	var pinned int

	err := row.Scan(
		&n.ID, &n.Kind, &customType, &n.Name, &summary, &body, &tagsJSON, &attrsJSON,
		&n.CreatedAt, &n.UpdatedAt, &n.LastAccessedAt, &n.AccessCount, &n.Strength, &pinned,
	)
	if err != nil {
		return nil, err
	}
	n.CustomType = customType.String
	n.Summary = summary.String
	n.Body = body.String
	n.Tags = unmarshalTags(tagsJSON)
	n.Attributes = unmarshalAttrs(attrsJSON)
	n.Pinned = pinned != 0
	return &n, nil
}

// scanNodes scans multiple nodes from *sql.Rows.
func scanNodes(rows *sql.Rows) ([]Node, error) {
	var nodes []Node
	for rows.Next() {
		var n Node
		var customType, summary, body sql.NullString
		var tagsJSON, attrsJSON string
		var pinned int

		err := rows.Scan(
			&n.ID, &n.Kind, &customType, &n.Name, &summary, &body, &tagsJSON, &attrsJSON,
			&n.CreatedAt, &n.UpdatedAt, &n.LastAccessedAt, &n.AccessCount, &n.Strength, &pinned,
		)
		if err != nil {
			return nil, err
		}
		n.CustomType = customType.String
		n.Summary = summary.String
		n.Body = body.String
		n.Tags = unmarshalTags(tagsJSON)
		n.Attributes = unmarshalAttrs(attrsJSON)
		n.Pinned = pinned != 0
		nodes = append(nodes, n)
	}
	if nodes == nil {
		nodes = []Node{}
	}
	return nodes, rows.Err()
}

func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
