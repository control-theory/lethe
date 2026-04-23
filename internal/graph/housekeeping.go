package graph

import (
	"context"
	"log/slog"
	"time"

	"github.com/control-theory/lethe/internal/events"
)

// RunHousekeeping starts a background loop that periodically materializes
// decayed strength values. It runs until ctx is cancelled.
// Only the leader process should call this.
func (s *Store) RunHousekeeping(ctx context.Context, interval time.Duration) {
	slog.Info("housekeeping started", "interval", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("housekeeping stopped")
			return
		case <-ticker.C:
			s.runSweep(ctx)
		}
	}
}

func (s *Store) runSweep(ctx context.Context) {
	updated, err := s.MaterializeStrength(ctx)
	if err != nil {
		slog.Error("housekeeping sweep failed", "error", err)
		return
	}

	if updated > 0 {
		slog.Info("housekeeping sweep complete", "nodes_updated", updated)
	}

	// Log the sweep event
	_ = events.Log(ctx, s.db, "decay_sweep", "", "system", map[string]int64{"updated": updated})
}
