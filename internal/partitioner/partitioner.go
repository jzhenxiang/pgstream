// Package partitioner provides key-based partitioning for WAL events,
// allowing events to be routed to consistent Kafka partitions or sink
// shards based on table name, primary key, or a custom field.
package partitioner

import (
	"errors"
	"fmt"
	"hash/fnv"

	"github.com/pgstream/pgstream/internal/wal"
)

// Strategy determines how a partition key is derived from an event.
type Strategy string

const (
	StrategyTable  Strategy = "table"
	StrategyPK     Strategy = "pk"
	StrategyCustom Strategy = "custom"
)

// Config holds configuration for the Partitioner.
type Config struct {
	Strategy    Strategy
	CustomField string // used when Strategy == StrategyCustom
	Partitions  int    // total number of partitions (> 0)
}

// Partitioner assigns WAL events to partition indices.
type Partitioner struct {
	cfg Config
}

// New creates a new Partitioner from the given Config.
func New(cfg Config) (*Partitioner, error) {
	if cfg.Partitions <= 0 {
		return nil, errors.New("partitioner: partitions must be greater than zero")
	}
	if cfg.Strategy == StrategyCustom && cfg.CustomField == "" {
		return nil, errors.New("partitioner: custom_field is required for custom strategy")
	}
	if cfg.Strategy == "" {
		cfg.Strategy = StrategyTable
	}
	return &Partitioner{cfg: cfg}, nil
}

// Assign returns the zero-based partition index for the given event.
func (p *Partitioner) Assign(event *wal.Event) (int, error) {
	if event == nil {
		return 0, errors.New("partitioner: nil event")
	}
	key, err := p.extractKey(event)
	if err != nil {
		return 0, fmt.Errorf("partitioner: extract key: %w", err)
	}
	return p.hash(key), nil
}

func (p *Partitioner) extractKey(event *wal.Event) (string, error) {
	switch p.cfg.Strategy {
	case StrategyTable:
		return event.Table, nil
	case StrategyPK:
		if event.PrimaryKey == "" {
			return event.Table, nil
		}
		return event.Table + ":" + event.PrimaryKey, nil
	case StrategyCustom:
		val, ok := event.Data[p.cfg.CustomField]
		if !ok {
			return event.Table, nil
		}
		return fmt.Sprintf("%v", val), nil
	default:
		return "", fmt.Errorf("unknown strategy %q", p.cfg.Strategy)
	}
}

func (p *Partitioner) hash(key string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return int(h.Sum32()) % p.cfg.Partitions
}
