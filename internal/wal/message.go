package wal

import "github.com/jackc/pglogrepl"

// MessageType represents the type of WAL change.
type MessageType string

const (
	MessageTypeInsert MessageType = "INSERT"
	MessageTypeUpdate MessageType = "UPDATE"
	MessageTypeDelete MessageType = "DELETE"
	MessageTypeTruncate MessageType = "TRUNCATE"
	MessageTypeBegin  MessageType = "BEGIN"
	MessageTypeCommit MessageType = "COMMIT"
)

// Message is a raw WAL message received from Postgres.
type Message struct {
	LSN  pglogrepl.LSN
	Data []byte
}

// Event is a decoded, structured WAL change event.
type Event struct {
	Type      MessageType        `json:"type"`
	LSN       string             `json:"lsn"`
	Schema    string             `json:"schema"`
	Table     string             `json:"table"`
	Columns   map[string]any     `json:"columns,omitempty"`
	OldColumns map[string]any    `json:"old_columns,omitempty"`
}
