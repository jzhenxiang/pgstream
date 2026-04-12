// Package pruner implements periodic replication-slot pruning for pgstream.
//
// Long-running or abandoned replication slots prevent Postgres from
// reclaiming WAL segments, which can exhaust disk space on the primary.
// The Pruner monitors a configured list of slot names and drops any that
// exceed a configurable lag threshold or that are explicitly listed for
// removal.
//
// Usage:
//
//	p, err := pruner.New(cfg, slotDropper, logger)
//	if err != nil { ... }
//	go p.Run(ctx)
package pruner
