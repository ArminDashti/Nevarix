package store

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Store is Managment's local SQLite database. It holds only internal data: the outbox of
// SoftEther snapshots that have not yet been acknowledged by Hub. A row is deleted only
// after Hub returns 2xx, so nothing is lost while Hub is down.
type Store struct {
	db *sql.DB
}

// Row is one pending outbox item.
type Row struct {
	ID      int64
	Payload []byte
}

const schema = `
CREATE TABLE IF NOT EXISTS hub_outbox (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    kind       TEXT NOT NULL,
    payload    BLOB NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
`

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Enqueue(ctx context.Context, kind string, payload []byte) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO hub_outbox (kind, payload) VALUES (?, ?)`, kind, payload)
	return err
}

// Pending returns the oldest rows of one kind, in insertion order.
func (s *Store) Pending(ctx context.Context, kind string, limit int) ([]Row, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, payload FROM hub_outbox WHERE kind = ? ORDER BY id LIMIT ?`, kind, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Row
	for rows.Next() {
		var r Row
		if err := rows.Scan(&r.ID, &r.Payload); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// Ack deletes a row after Hub confirmed it is stored.
func (s *Store) Ack(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM hub_outbox WHERE id = ?`, id)
	return err
}
