package sink

import "context"

// Sink defines the interface for WAL event destinations.
type Sink interface {
	// Send delivers a WAL event payload to the destination.
	Send(ctx context.Context, event Event) error
	// Close releases any resources held by the sink.
	Close() error
}

// Event represents a decoded WAL change ready for delivery.
type Event struct {
	// LSN is the log sequence number of the WAL record.
	LSN string
	// Table is the relation name the change belongs to.
	Table string
	// Action is one of INSERT, UPDATE, DELETE.
	Action string
	// Data holds the column key/value pairs for the change.
	Data map[string]interface{}
}

// Type enumerates supported sink backends.
type Type string

const (
	TypeKafka   Type = "kafka"
	TypeWebhook Type = "webhook"
)
