package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

// Log writes an event to the memory_events table.
// If tx is nil, it executes directly against db.
func Log(ctx context.Context, db interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}, op, targetID, actor string, payload any) error {
	var payloadStr sql.NullString
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		payloadStr = sql.NullString{String: string(b), Valid: true}
	}
	_, err := db.ExecContext(ctx,
		`INSERT INTO memory_events (ts, op, target_id, actor, payload) VALUES (?, ?, ?, ?, ?)`,
		time.Now().UnixMilli(), op, targetID, actor, payloadStr,
	)
	return err
}
