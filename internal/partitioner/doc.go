// Package partitioner provides deterministic partition assignment for WAL
// events. It supports three built-in strategies:
//
//   - table:  partition is derived from the event's table name.
//   - pk:     partition is derived from table + primary key, keeping rows
//             for the same key on the same partition (order preserved).
//   - custom: partition is derived from an arbitrary data field, useful
//             when a domain-level tenant or shard key is available.
//
// All strategies use FNV-32a hashing modulo the configured partition count,
// giving a uniform, deterministic distribution without external dependencies.
package partitioner
