package wal

import (
	"context"
	"fmt"
	"log"

	"github.com/pgstream/pgstream/internal/sink"
)

// Processor reads WAL messages and forwards decoded events to a sink.
type Processor struct {
	reader  *Reader
	decoder *Decoder
	sink    sink.Sink
}

// NewProcessor creates a new WAL processor.
func NewProcessor(reader *Reader, decoder *Decoder, s sink.Sink) *Processor {
	return &Processor{
		reader:  reader,
		decoder: decoder,
		sink:    s,
	}
}

// Run starts the processing loop, reading WAL messages until the context is cancelled.
func (p *Processor) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			log.Println("processor: context cancelled, stopping")
			return ctx.Err()
		default:
		}

		msg, err := p.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("processor: read message: %w", err)
		}

		if msg == nil {
			continue
		}

		event, err := p.decoder.Decode(msg)
		if err != nil {
			log.Printf("processor: decode message: %v", err)
			continue
		}

		if event == nil {
			continue
		}

		if err := p.sink.Send(ctx, event); err != nil {
			return fmt.Errorf("processor: send event: %w", err)
		}
	}
}
