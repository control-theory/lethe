package graph

import (
	"context"
	"database/sql"
	"sync"

	"modernc.org/sqlite"
)

var registerOnce sync.Once

func registerPragmaHook() {
	registerOnce.Do(func() {
		sqlite.RegisterConnectionHook(func(conn sqlite.ExecQuerierContext, dsn string) error {
			_, err := conn.ExecContext(context.Background(), `
				PRAGMA journal_mode = WAL;
				PRAGMA synchronous = NORMAL;
				PRAGMA busy_timeout = 5000;
				PRAGMA foreign_keys = ON;
				PRAGMA wal_autocheckpoint = 1000;
			`, nil)
			return err
		})
	})
}

// OpenDB opens (or creates) the SQLite database at path, runs migrations,
// and returns the configured *sql.DB.
func OpenDB(path string) (*sql.DB, error) {
	registerPragmaHook()

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(0)

	if err := Migrate(context.Background(), db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// OpenMemoryDB opens an in-memory SQLite database for testing.
func OpenMemoryDB() (*sql.DB, error) {
	return OpenDB(":memory:")
}
