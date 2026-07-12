package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// EventLog represents a single event persisted to PostgreSQL.
type EventLog struct {
	ID        int64           `json:"id"`
	Server    string          `json:"server"`
	Container string          `json:"container"`
	Image     string          `json:"image"`
	Action    string          `json:"action"`
	EventType string          `json:"event_type"`
	Status    string          `json:"status"`
	Message   string          `json:"message"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}

// InsertEvent persists a new event to the event_logs table.
func (db *DB) InsertEvent(ctx context.Context, e *EventLog) error {
	payload := e.Payload
	if payload == nil {
		payload = json.RawMessage("{}")
	}

	query := `
		INSERT INTO event_logs (server, container, image, action, event_type, status, message, payload, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	return db.Pool.QueryRow(ctx, query,
		e.Server,
		e.Container,
		e.Image,
		e.Action,
		e.EventType,
		e.Status,
		e.Message,
		payload,
		time.Now(),
	).Scan(&e.ID)
}

// ListEvents returns paginated event logs with optional filters.
func (db *DB) ListEvents(ctx context.Context, filter EventFilter) ([]EventLog, error) {
	query := `
		SELECT id, server, container, image, action, event_type, status, message, payload, created_at
		FROM event_logs
		WHERE ($1 = '' OR server = $1)
		  AND ($2 = '' OR container = $2)
		  AND ($3 = '' OR action = $3)
		  AND ($4 = '' OR event_type = $4)
		ORDER BY created_at DESC
		LIMIT $5 OFFSET $6`

	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	rows, err := db.Pool.Query(ctx, query,
		filter.Server,
		filter.Container,
		filter.Action,
		filter.EventType,
		limit,
		filter.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("event_log: list events: %w", err)
	}
	defer rows.Close()

	var events []EventLog
	for rows.Next() {
		var e EventLog
		if err := rows.Scan(
			&e.ID, &e.Server, &e.Container, &e.Image,
			&e.Action, &e.EventType, &e.Status, &e.Message,
			&e.Payload, &e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("event_log: scan row: %w", err)
		}
		events = append(events, e)
	}

	return events, rows.Err()
}

// EventFilter defines query parameters for listing events.
type EventFilter struct {
	Server    string
	Container string
	Action    string
	EventType string
	Limit     int
	Offset    int
}
