package graph

import (
	"context"
	"strings"
)

// Search performs an FTS5 search over nodes, ranked by relevance.
// If query is empty, falls back to ListRecent.
func (s *Store) Search(ctx context.Context, query string, opts SearchOpts) ([]Node, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return s.ListRecent(ctx, opts.Limit, opts.Kind)
	}

	if opts.Limit <= 0 {
		opts.Limit = 10
	}

	// Escape double quotes in user query for FTS5 safety, then use implicit AND.
	safeQuery := strings.ReplaceAll(query, `"`, `""`)

	var args []any
	var conditions []string

	// FTS match
	conditions = append(conditions, `nodes_fts MATCH ?`)
	args = append(args, safeQuery)

	// Kind filter
	if opts.Kind != nil {
		conditions = append(conditions, `n.kind = ?`)
		args = append(args, string(*opts.Kind))
	}

	// Exclude faded unless requested
	if !opts.IncludeFaded {
		conditions = append(conditions, `n.strength >= 0.1`)
	}

	where := strings.Join(conditions, " AND ")
	args = append(args, opts.Limit)

	q := `SELECT n.id, n.kind, n.custom_type, n.name, n.summary, n.body, n.tags, n.attributes,
		n.created_at, n.updated_at, n.last_accessed_at, n.access_count, n.strength, n.pinned
		FROM nodes_fts f
		JOIN nodes n ON n.rowid = f.rowid
		WHERE ` + where + `
		ORDER BY bm25(nodes_fts)
		LIMIT ?`

	rows, err := s.db.QueryContext(ctx, q, args...)
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
