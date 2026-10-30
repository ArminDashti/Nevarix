package store

import (
	"context"
	"strconv"
	"time"

	"github.com/ArminDashti/nevarix-core/core/model"
)

// ListApps returns installed apps across agents, optionally scoped to one agent.
func (s *Store) ListApps(ctx context.Context, agentID string) ([]model.App, error) {
	query := `
		SELECT agent_id, name, version, publisher, install_date, source
		FROM agent_apps`
	args := []any{}
	if agentID != "" {
		query += ` WHERE agent_id = $1`
		args = append(args, agentID)
	}
	query += ` ORDER BY name`
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.App{}
	for rows.Next() {
		var a model.App
		if err := rows.Scan(&a.AgentID, &a.Name, &a.Version, &a.Publisher, &a.InstallDate, &a.Source); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// ListEvents returns the latest agent events, newest first.
func (s *Store) ListEvents(ctx context.Context, agentID string, limit int) ([]model.Event, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	query := `
		SELECT agent_id, kind, message, occurred_at
		FROM agent_events`
	args := []any{}
	if agentID != "" {
		query += ` WHERE agent_id = $1`
		args = append(args, agentID)
	}
	query += ` ORDER BY occurred_at DESC, id DESC LIMIT ` + strconv.Itoa(limit)
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Event{}
	for rows.Next() {
		var e model.Event
		if err := rows.Scan(&e.AgentID, &e.Kind, &e.Message, &e.OccurredAt); err != nil {
			return nil, err
		}
		e.OccurredAt = e.OccurredAt.UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}

// ListActions returns the latest recorded dashboard actions, newest first.
func (s *Store) ListActions(ctx context.Context, limit int) ([]model.Action, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	rows, err := s.pool.Query(ctx, `
		SELECT kind, detail, created_at FROM core_actions ORDER BY created_at DESC, id DESC LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Action{}
	for rows.Next() {
		var a model.Action
		if err := rows.Scan(&a.Kind, &a.Detail, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.CreatedAt = a.CreatedAt.UTC()
		out = append(out, a)
	}
	return out, rows.Err()
}

// RecordAction stores one dashboard action audit row.
func (s *Store) RecordAction(ctx context.Context, kind, detail string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO core_actions (kind, detail, created_at) VALUES ($1, $2, $3)
	`, kind, detail, time.Now().UTC())
	return err
}