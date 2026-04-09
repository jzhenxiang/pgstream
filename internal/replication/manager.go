package replication

import (
	"context"
	"fmt"
	"log"
)

// Manager coordinates replication slot lifecycle operations.
type Manager struct {
	slot   *Slot
	logger *log.Logger
}

// NewManager creates a Manager with the given slot.
func NewManager(slot *Slot, logger *log.Logger) (*Manager, error) {
	if slot == nil {
		return nil, fmt.Errorf("slot must not be nil")
	}
	if logger == nil {
		logger = log.Default()
	}
	return &Manager{slot: slot, logger: logger}, nil
}

// EnsureSlot verifies the slot is ready, logging its name.
func (m *Manager) EnsureSlot(ctx context.Context) error {
	if m.slot.Name() == "" {
		return fmt.Errorf("replication slot has no name")
	}
	m.logger.Printf("replication: slot %q is ready", m.slot.Name())
	return nil
}

// Teardown drops the replication slot and closes the connection.
func (m *Manager) Teardown(ctx context.Context) error {
	m.logger.Printf("replication: dropping slot %q", m.slot.Name())
	if err := m.slot.Drop(ctx); err != nil {
		return fmt.Errorf("teardown slot: %w", err)
	}
	return m.slot.Close(ctx)
}

// SlotName returns the managed slot name.
func (m *Manager) SlotName() string {
	return m.slot.Name()
}
