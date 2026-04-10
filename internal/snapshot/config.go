package snapshot

import "fmt"

// Validate checks the Config for required fields and applies defaults.
func (c *Config) Validate() error {
	if c.DSN == "" {
		return fmt.Errorf("snapshot config: DSN must not be empty")
	}
	if len(c.Tables) == 0 {
		return fmt.Errorf("snapshot config: Tables must not be empty")
	}
	for i, t := range c.Tables {
		if t == "" {
			return fmt.Errorf("snapshot config: Tables[%d] must not be empty", i)
		}
	}
	if c.BatchSize < 0 {
		return fmt.Errorf("snapshot config: BatchSize must be non-negative")
	}
	if c.BatchSize == 0 {
		c.BatchSize = 500
	}
	return nil
}
