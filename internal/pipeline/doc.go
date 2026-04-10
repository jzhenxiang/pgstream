// Package pipeline provides the top-level orchestration layer for pgstream.
//
// It wires together the individual components — WAL reader, event filter,
// column transformer, and output sink — into a single Run loop that can be
// started and stopped via context cancellation.
//
// Typical usage:
//
//	cfg, _ := config.LoadConfig("pgstream.yaml")
//	p, _ := pipeline.Build(cfg, slog.Default())
//	p.Run(ctx)
//
// The Build helper reads the application Config and selects the appropriate
// sink (Kafka or Webhook) automatically.  For more control, construct a
// pipeline.Config manually and call pipeline.New directly.
package pipeline
