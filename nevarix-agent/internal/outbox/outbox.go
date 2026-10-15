// Package outbox keeps collected snapshots in SQLite until Core acknowledges them.
package outbox

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

// Item is one queued snapshot.
type Item struct {
	ID      int64
	Payload []byte
}

const schema = `
CREATE TABLE IF NOT EXISTS outbox (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
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

func (s *Store) Enqueue(ctx context.Context, payload []byte) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO outbox (payload) VALUES (?)`, payload)
	return err
}

// Pending returns the oldest items first, so replay keeps the collection order.
func (s *Store) Pending(ctx context.Context, limit int) ([]Item, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, payload FROM outbox ORDER BY id LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Item
	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.ID, &it.Payload); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// Ack removes an item only after Core confirmed it is stored.
func (s *Store) Ack(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM outbox WHERE id = ?`, id)
	return err
}
