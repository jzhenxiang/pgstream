package replication

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	defaultSlotPlugin = "pgoutput"
	defaultSlotTimeout = 10 * time.Second
)

// SlotConfig holds configuration for a replication slot.
type SlotConfig struct {
	DSN        string
	SlotName   string
	Plugin     string
	CreateSlot bool
}

// Slot manages a PostgreSQL logical replication slot.
type Slot struct {
	cfg SlotConfig
	conn *pgx.Conn
}

// NewSlot creates a new Slot instance and optionally creates the replication slot.
func NewSlot(ctx context.Context, cfg SlotConfig) (*Slot, error) {
	if cfg.SlotName == "" {
		return nil, fmt.Errorf("replication slot name is required")
	}
	if cfg.DSN == "" {
		return nil, fmt.Errorf("DSN is required")
	}
	if cfg.Plugin == "" {
		cfg.Plugin = defaultSlotPlugin
	}

	conn, err := pgx.Connect(ctx, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	s := &Slot{cfg: cfg, conn: conn}

	if cfg.CreateSlot {
		if err := s.create(ctx); err != nil {
			_ = conn.Close(ctx)
			return nil, err
		}
	}

	return s, nil
}

// create creates the replication slot if it does not already exist.
func (s *Slot) create(ctx context.Context) error {
	var name string
	err := s.conn.QueryRow(ctx,
		"SELECT slot_name FROM pg_replication_slots WHERE slot_name = $1",
		s.cfg.SlotName,
	).Scan(&name)

	if err == nil {
		// slot already exists
		return nil
	}

	_, err = s.conn.Exec(ctx,
		fmt.Sprintf("SELECT pg_create_logical_replication_slot('%s', '%s')",
			s.cfg.SlotName, s.cfg.Plugin),
	)
	if err != nil {
		return fmt.Errorf("create replication slot %q: %w", s.cfg.SlotName, err)
	}
	return nil
}

// Drop removes the replication slot from PostgreSQL.
func (s *Slot) Drop(ctx context.Context) error {
	_, err := s.conn.Exec(ctx,
		fmt.Sprintf("SELECT pg_drop_replication_slot('%s')", s.cfg.SlotName))
	if err != nil {
		return fmt.Errorf("drop replication slot %q: %w", s.cfg.SlotName, err)
	}
	return nil
}

// Close closes the underlying database connection.
func (s *Slot) Close(ctx context.Context) error {
	return s.conn.Close(ctx)
}

// Name returns the slot name.
func (s *Slot) Name() string {
	return s.cfg.SlotName
}
