package graph

import (
	"context"
	"database/sql"
)

const ddl = `
CREATE TABLE IF NOT EXISTS nodes (
    id              TEXT PRIMARY KEY,
    kind            TEXT NOT NULL,
    custom_type     TEXT,
    name            TEXT NOT NULL,
    summary         TEXT,
    body            TEXT,
    tags            TEXT NOT NULL DEFAULT '[]',
    attributes      TEXT NOT NULL DEFAULT '{}',
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL,
    last_accessed_at INTEGER NOT NULL,
    access_count     INTEGER NOT NULL DEFAULT 0,
    strength         REAL    NOT NULL DEFAULT 1.0,
    pinned           INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS edges (
    id              TEXT PRIMARY KEY,
    src_id          TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    dst_id          TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    relation        TEXT NOT NULL,
    custom_relation TEXT,
    weight          REAL NOT NULL DEFAULT 1.0,
    attributes      TEXT NOT NULL DEFAULT '{}',
    created_at      INTEGER NOT NULL,
    last_accessed_at INTEGER NOT NULL,
    access_count     INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_nodes_kind ON nodes(kind);
CREATE INDEX IF NOT EXISTS idx_nodes_updated ON nodes(updated_at);
CREATE INDEX IF NOT EXISTS idx_nodes_strength ON nodes(strength);
CREATE INDEX IF NOT EXISTS idx_edges_src ON edges(src_id);
CREATE INDEX IF NOT EXISTS idx_edges_dst ON edges(dst_id);
CREATE INDEX IF NOT EXISTS idx_edges_relation ON edges(relation);

CREATE TABLE IF NOT EXISTS memory_events (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    ts         INTEGER NOT NULL,
    op         TEXT NOT NULL,
    target_id  TEXT,
    actor      TEXT NOT NULL,
    payload    TEXT
);

CREATE INDEX IF NOT EXISTS idx_events_ts ON memory_events(ts);

CREATE TABLE IF NOT EXISTS config (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
`

const ftsSetup = `
CREATE VIRTUAL TABLE IF NOT EXISTS nodes_fts USING fts5(
    name, summary, body, tags,
    content='nodes', content_rowid='rowid',
    tokenize='porter unicode61'
);

-- Triggers to keep FTS in sync with nodes table.
CREATE TRIGGER IF NOT EXISTS nodes_ai AFTER INSERT ON nodes BEGIN
    INSERT INTO nodes_fts(rowid, name, summary, body, tags)
    VALUES (new.rowid, new.name, COALESCE(new.summary,''), COALESCE(new.body,''), new.tags);
END;

CREATE TRIGGER IF NOT EXISTS nodes_ad AFTER DELETE ON nodes BEGIN
    INSERT INTO nodes_fts(nodes_fts, rowid, name, summary, body, tags)
    VALUES ('delete', old.rowid, old.name, COALESCE(old.summary,''), COALESCE(old.body,''), old.tags);
END;

CREATE TRIGGER IF NOT EXISTS nodes_au AFTER UPDATE ON nodes BEGIN
    INSERT INTO nodes_fts(nodes_fts, rowid, name, summary, body, tags)
    VALUES ('delete', old.rowid, old.name, COALESCE(old.summary,''), COALESCE(old.body,''), old.tags);
    INSERT INTO nodes_fts(rowid, name, summary, body, tags)
    VALUES (new.rowid, new.name, COALESCE(new.summary,''), COALESCE(new.body,''), new.tags);
END;
`

const defaultConfig = `
INSERT OR IGNORE INTO config (key, value) VALUES ('half_life_days', '30');
INSERT OR IGNORE INTO config (key, value) VALUES ('reinforcement_bump', '1.0');
INSERT OR IGNORE INTO config (key, value) VALUES ('strength_cap', '10.0');
`

// Migrate runs the schema DDL against the database.
func Migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, ftsSetup); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, defaultConfig); err != nil {
		return err
	}
	return nil
}
