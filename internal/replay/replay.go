// Package replay provides WAL event replay from a persisted offset position.
// It allows pgstream to resume streaming from a known LSN after a restart
// or failure, avoiding duplicate or missed events.
package replay

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/pgstream/pgstream/internal/offset"
	"github.com/pgstream/pgstream/internal/wal"
)

// Reader is the interface for reading WAL messages.
type Reader interface {
	Read(ctx context.Context) (*wal.Message, error)
}

// Sink is the interface for writing replayed events.
type Sink interface {
	Send(ctx context.Context, event *wal.Event) error
}

// Replayer replays WAL events starting from the last committed offset.
type Replayer struct {
	reader Reader
	sink   Sink
	offset *offset.Offset
	logger *slog.Logger
}

// New creates a new Replayer. Returns an error if reader, sink, or offset is nil.
func New(reader Reader, sink Sink, off *offset.Offset, logger *slog.Logger) (*Replayer, error) {
	if reader == nil {
		return nil, fmt.Errorf("replay: reader is required")
	}
	if sink == nil {
		return nil, fmt.Errorf("replay: sink is required")
	}
	if off == nil {
		return nil, fmt.Errorf("replay: offset is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Replayer{
		reader: reader,
		sink:   sink,
		offset: off,
		logger: logger,
	}, nil
}

// Run reads WAL messages and replays them to the sink until ctx is cancelled.
func (r *Replayer) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, err := r.reader.Read(ctx)
		if err != nil {
			return fmt.Errorf("replay: read error: %w", err)
		}
		if msg == nil {
			continue
		}

		if err := r.sink.Send(ctx, msg.Event); err != nil {
			r.logger.Error("replay: sink send failed", "lsn", msg.LSN, "error", err)
			return fmt.Errorf("replay: sink error: %w", err)
		}

		if err := r.offset.Commit(ctx, msg.LSN); err != nil {
			r.logger.Warn("replay: failed to commit offset", "lsn", msg.LSN, "error", err)
		}
	}
}
