package graph

import (
	"context"
	"database/sql"
	"time"
)

// Store is the central data access layer for the memory graph.
type Store struct {
	db             *sql.DB
	HalfLifeDays   float64
	ReinforceBump  float64
	StrengthCap    float64
}

// NewStore wraps an existing database connection with default decay params.
func NewStore(db *sql.DB) *Store {
	return &Store{
		db:            db,
		HalfLifeDays:  30,
		ReinforceBump: 1.0,
		StrengthCap:   10.0,
	}
}

// DB returns the underlying database, for use by event logging.
func (s *Store) DB() *sql.DB {
	return s.db
}

// applyDecayToNodes applies decay-adjusted strength to a slice of nodes in place.
func (s *Store) applyDecayToNodes(nodes []Node) {
	for i := range nodes {
		ApplyDecay(&nodes[i], s.HalfLifeDays)
	}
}

// Stats returns aggregate statistics about the graph.
func (s *Store) Stats(ctx context.Context) (*GraphStats, error) {
	stats := &GraphStats{
		ByKind: make(map[string]int),
	}

	// Total nodes
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM nodes`).Scan(&stats.TotalNodes); err != nil {
		return nil, err
	}
	// Total edges
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM edges`).Scan(&stats.TotalEdges); err != nil {
		return nil, err
	}
	// By kind
	rows, err := s.db.QueryContext(ctx, `SELECT kind, COUNT(*) FROM nodes GROUP BY kind`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var kind string
		var count int
		if err := rows.Scan(&kind, &count); err != nil {
			return nil, err
		}
		stats.ByKind[kind] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Pinned count
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM nodes WHERE pinned = 1`).Scan(&stats.PinnedCount); err != nil {
		return nil, err
	}
	// Fading count — use stored strength as approximation (housekeeping materializes it)
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM nodes WHERE strength < 0.1 AND pinned = 0`).Scan(&stats.FadingCount); err != nil {
		return nil, err
	}
	// Recent (last 24h)
	cutoff := time.Now().Add(-24 * time.Hour).UnixMilli()
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM nodes WHERE updated_at > ?`, cutoff).Scan(&stats.RecentCount); err != nil {
		return nil, err
	}

	return stats, nil
}

// ListEvents returns recent audit log events.
func (s *Store) ListEvents(ctx context.Context, limit int) ([]MemoryEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, ts, op, target_id, actor, payload FROM memory_events ORDER BY ts DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []MemoryEvent
	for rows.Next() {
		var ev MemoryEvent
		var targetID, payload sql.NullString
		if err := rows.Scan(&ev.ID, &ev.TS, &ev.Op, &targetID, &ev.Actor, &payload); err != nil {
			return nil, err
		}
		ev.TargetID = targetID.String
		ev.Payload = payload.String
		events = append(events, ev)
	}
	if events == nil {
		events = []MemoryEvent{}
	}
	return events, rows.Err()
}
