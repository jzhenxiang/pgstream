package wal

import (
	"context"
	"fmt"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgconn"
)

// Reader reads WAL changes from a PostgreSQL replication slot.
type Reader struct {
	conn       *pgconn.PgConn
	slotName   string
	publication string
	lastLSN    pglogrepl.LSN
}

// NewReader creates a new WAL reader connected to the given DSN.
func NewReader(ctx context.Context, dsn, slotName, publication string) (*Reader, error) {
	conn, err := pgconn.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("wal: connect: %w", err)
	}
	return &Reader{
		conn:        conn,
		slotName:    slotName,
		publication: publication,
	}, nil
}

// Start begins logical replication and streams messages to the returned channel.
func (r *Reader) Start(ctx context.Context) (<-chan Message, <-chan error) {
	msgCh := make(chan Message, 64)
	errCh := make(chan error, 1)

	go func() {
		defer close(msgCh)
		defer close(errCh)
		if err := r.stream(ctx, msgCh); err != nil {
			errCh <- err
		}
	}()

	return msgCh, errCh
}

func (r *Reader) stream(ctx context.Context, msgCh chan<- Message) error {
	pluginArgs := []string{
		"proto_version '1'",
		fmt.Sprintf("publication_names '%s'", r.publication),
	}

	result, err := pglogrepl.CreateReplicationSlot(
		ctx, r.conn, r.slotName, "pgoutput",
		pglogrepl.CreateReplicationSlotOptions{Temporary: true},
	)
	if err != nil {
		return fmt.Errorf("wal: create slot: %w", err)
	}

	r.lastLSN, err = pglogrepl.ParseLSN(result.ConsistentPoint)
	if err != nil {
		return fmt.Errorf("wal: parse lsn: %w", err)
	}

	if err := pglogrepl.StartReplication(ctx, r.conn, r.slotName, r.lastLSN,
		pglogrepl.StartReplicationOptions{PluginArgs: pluginArgs}); err != nil {
		return fmt.Errorf("wal: start replication: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		rawMsg, err := r.conn.ReceiveMessage(ctx)
		if err != nil {
			return fmt.Errorf("wal: receive: %w", err)
		}

		if errMsg, ok := rawMsg.(*pgconn.PgError); ok {
			return fmt.Errorf("wal: postgres error: %s", errMsg.Message)
		}

		msg, ok := rawMsg.(*pgconn.CopyData)
		if !ok {
			continue
		}

		msgCh <- Message{LSN: r.lastLSN, Data: msg.Data}
	}
}

// Close closes the underlying connection.
func (r *Reader) Close(ctx context.Context) error {
	return r.conn.Close(ctx)
}
