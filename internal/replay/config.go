package replay

import "fmt"

// Config holds configuration for the Replayer.
type Config struct {
	// StartLSN is the LSN to start replaying from. If empty, replay starts
	// from the last committed offset stored on disk.
	StartLSN string `yaml:"start_lsn" env:"PGSTREAM_REPLAY_START_LSN"`

	// OffsetFile is the path to the file used to persist the committed offset.
	OffsetFile string `yaml:"offset_file" env:"PGSTREAM_REPLAY_OFFSET_FILE"`
}

// Validate returns an error if the Config is invalid.
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("replay: config is required")
	}
	if c.OffsetFile == "" {
		return fmt.Errorf("replay: offset_file is required")
	}
	return nil
}
