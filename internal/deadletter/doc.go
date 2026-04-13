// Package deadletter provides a thread-safe in-memory dead-letter queue for
// WAL events that could not be delivered to a sink after all retry attempts
// have been exhausted.
//
// Usage:
//
//	q := deadletter.New(500)
//	q.Push(ctx, event, err, attempts)
//	entries := q.Drain()
//
// The queue evicts the oldest entry when capacity is exceeded, ensuring that
// memory usage remains bounded at the cost of potentially losing the earliest
// failures.
package deadletter
