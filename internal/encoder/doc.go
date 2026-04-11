// Package encoder provides JSON encoding for WAL events.
//
// Events are wrapped in an Envelope that includes a schema version,
// a UTC timestamp, and optional metadata before being serialised to
// JSON. This gives downstream consumers (Kafka topics, webhook
// endpoints) a stable, self-describing message format.
//
// Basic usage:
//
//	enc := encoder.New()
//	b, err := enc.Encode(event)
//
// With metadata:
//
//	b, err := enc.EncodeWithMeta(event, map[string]any{
//		"source": "pgstream",
//		"slot":   "my_slot",
//	})
package encoder
