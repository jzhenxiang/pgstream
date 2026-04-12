package partitioner

import "errors"

// Validate checks that the Config is well-formed.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("partitioner: config is nil")
	}
	if c.Partitions <= 0 {
		return errors.New("partitioner: partitions must be greater than zero")
	}
	switch c.Strategy {
	case "", StrategyTable, StrategyPK:
		// valid
	case StrategyCustom:
		if c.CustomField == "" {
			return errors.New("partitioner: custom_field required for custom strategy")
		}
	default:
		return errors.New("partitioner: unknown strategy")
	}
	return nil
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() Config {
	return Config{
		Strategy:   StrategyTable,
		Partitions: 8,
	}
}
