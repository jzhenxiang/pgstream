// Package snapshot provides initial table snapshot functionality,
// allowing pgstream to capture the current state of tables before
// streaming incremental WAL changes.
package snapshot

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/your-org/pgstream/internal/wal"
)

// Config holds configuration for the snapshot process.
type Config struct {
	DSN    string
	Tables []string
	BatchSize int
}

// Snapshot reads existing rows from Postgres tables and emits them
// as synthetic WAL events so they flow through the normal pipeline.
type Snapshot struct {
	cfg Config
	db  *sql.DB
}

// New creates a new Snapshot. It validates the config and opens a
// database connection.
func New(cfg Config) (*Snapshot, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("snapshot: DSN is required")
	}
	if len(cfg.Tables) == 0 {
		return nil, fmt.Errorf("snapshot: at least one table is required")
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 500
	}

	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("snapshot: open db: %w", err)
	}
	return &Snapshot{cfg: cfg, db: db}, nil
}

// Run iterates over configured tables, reads rows in batches and calls
// emit for each synthetic event. It stops early if ctx is cancelled.
func (s *Snapshot) Run(ctx context.Context, emit func(context.Context, *wal.Event) error) error {
	defer s.db.Close()

	for _, table := range s.cfg.Tables {
		if err := s.snapshotTable(ctx, table, emit); err != nil {
			return err
		}
	}
	return nil
}

func (s *Snapshot) snapshotTable(ctx context.Context, table string, emit func(context.Context, *wal.Event) error) error {
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf("SELECT * FROM %s", table))
	if err != nil {
		return fmt.Errorf("snapshot: query %s: %w", table, err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("snapshot: columns %s: %w", table, err)
	}

	for rows.Next() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return fmt.Errorf("snapshot: scan %s: %w", table, err)
		}

		data := make(map[string]interface{}, len(cols))
		for i, col := range cols {
			data[col] = vals[i]
		}

		event := &wal.Event{
			Type:      "snapshot",
			Schema:    "public",
			Table:     table,
			Timestamp: time.Now().UTC(),
			Data:      data,
		}
		if err := emit(ctx, event); err != nil {
			return fmt.Errorf("snapshot: emit %s: %w", table, err)
		}
	}
	return rows.Err()
}
