package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration for pgstream
type Config struct {
	Postgres PostgresConfig `yaml:"postgres"`
	Output   OutputConfig   `yaml:"output"`
	Server   ServerConfig   `yaml:"server"`
}

// PostgresConfig holds PostgreSQL connection settings
type PostgresConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	Database       string `yaml:"database"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	SlotName       string `yaml:"slot_name"`
	PublicationName string `yaml:"publication_name"`
}

// OutputConfig defines where to send CDC events
type OutputConfig struct {
	Type    string        `yaml:"type"` // "kafka" or "webhook"
	Kafka   *KafkaConfig  `yaml:"kafka,omitempty"`
	Webhook *WebhookConfig `yaml:"webhook,omitempty"`
}

// KafkaConfig holds Kafka-specific settings
type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
}

// WebhookConfig holds webhook-specific settings
type WebhookConfig struct {
	URL     string        `yaml:"url"`
	Timeout time.Duration `yaml:"timeout"`
}

// ServerConfig holds server settings
type ServerConfig struct {
	Port int `yaml:"port"`
}

// LoadConfig reads and parses the configuration file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Postgres.Host == "" {
		return fmt.Errorf("postgres host is required")
	}
	if c.Postgres.Database == "" {
		return fmt.Errorf("postgres database is required")
	}
	if c.Postgres.User == "" {
		return fmt.Errorf("postgres user is required")
	}
	if c.Output.Type != "kafka" && c.Output.Type != "webhook" {
		return fmt.Errorf("output type must be 'kafka' or 'webhook'")
	}
	if c.Output.Type == "kafka" && c.Output.Kafka == nil {
		return fmt.Errorf("kafka config is required when output type is 'kafka'")
	}
	if c.Output.Type == "webhook" && c.Output.Webhook == nil {
		return fmt.Errorf("webhook config is required when output type is 'webhook'")
	}
	return nil
}
