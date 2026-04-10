// Package schema implements an in-memory, concurrency-safe cache for Postgres
// table schema metadata discovered during WAL replication.
//
// During logical replication Postgres emits relation messages that describe
// the column layout of every table touched by a transaction. The Cache type
// stores these descriptions so that downstream components (decoders,
// transformers, sinks) can look up column names and types without issuing
// additional queries to the database.
//
// Usage:
//
//	cache := schema.New()
//	cache.Set(&schema.TableSchema{
//		Schema:  "public",
//		Table:   "orders",
//		Columns: []schema.Column{{Name: "id", Type: "int4", Position: 1}},
//	})
//	ts, ok := cache.Get("public", "orders")
package schema
