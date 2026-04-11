// Package fanout implements a multi-sink dispatcher for pgstream.
//
// It allows a single stream of WAL events to be forwarded to multiple
// downstream sinks (e.g. Kafka and a webhook) concurrently. All sinks
// receive every event; if any sink fails, the combined error is returned
// while the remaining sinks still process the event.
//
// Usage:
//
//	f, err := fanout.New(kafkaSink, webhookSink)
//	if err != nil {
//		log.Fatal(err)
//	}
//	if err := f.Send(ctx, event); err != nil {
//		log.Println("one or more sinks failed:", err)
//	}
package fanout
