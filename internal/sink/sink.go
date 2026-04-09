package sink

import "context"

// Sink is the interface that all event destinations must implement.
type Sink interface {
	// Send delivers an event to the destination.
	Send(ctx context.Context, event any) error
	// Close releases any resources held by the sink.
	Close() error
}
