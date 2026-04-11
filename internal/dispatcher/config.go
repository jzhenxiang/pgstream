package dispatcher

import "fmt"

// RouteConfig is the serialisable form of a Route, used when loading
// dispatcher rules from a configuration file or environment variables.
type RouteConfig struct {
	// Table is the fully-qualified table name, e.g. "public.orders".
	Table string `yaml:"table" json:"table"`
	// Sinks lists the sink identifiers that should receive events for Table.
	// The identifiers must match sink names registered in the pipeline builder.
	Sinks []string `yaml:"sinks" json:"sinks"`
}

// Config holds the full dispatcher configuration.
type Config struct {
	Routes       []RouteConfig `yaml:"routes"        json:"routes"`
	DefaultSinks []string      `yaml:"default_sinks" json:"default_sinks"`
}

// Validate returns an error when the configuration is semantically invalid.
func (c *Config) Validate() error {
	if len(c.Routes) == 0 && len(c.DefaultSinks) == 0 {
		return fmt.Errorf("dispatcher config: at least one route or default_sink must be specified")
	}
	seen := make(map[string]struct{}, len(c.Routes))
	for i, r := range c.Routes {
		if r.Table == "" {
			return fmt.Errorf("dispatcher config: route[%d] table must not be empty", i)
		}
		if _, dup := seen[r.Table]; dup {
			return fmt.Errorf("dispatcher config: duplicate route for table %q", r.Table)
		}
		seen[r.Table] = struct{}{}
		if len(r.Sinks) == 0 {
			return fmt.Errorf("dispatcher config: route[%d] (%s) must reference at least one sink", i, r.Table)
		}
	}
	return nil
}
