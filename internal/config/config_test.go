package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	configContent := `
postgres:
  host: localhost
  port: 5432
  database: testdb
  user: testuser
  password: testpass
  slot_name: pgstream_slot
  publication_name: pgstream_pub

output:
  type: kafka
  kafka:
    brokers:
      - localhost:9092
    topic: cdc_events

server:
  port: 8080
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Postgres.Host != "localhost" {
		t.Errorf("expected host 'localhost', got '%s'", cfg.Postgres.Host)
	}
	if cfg.Postgres.Port != 5432 {
		t.Errorf("expected port 5432, got %d", cfg.Postgres.Port)
	}
	if cfg.Output.Type != "kafka" {
		t.Errorf("expected output type 'kafka', got '%s'", cfg.Output.Type)
	}
	if cfg.Output.Kafka == nil {
		t.Fatal("kafka config should not be nil")
	}
	if len(cfg.Output.Kafka.Brokers) != 1 {
		t.Errorf("expected 1 broker, got %d", len(cfg.Output.Kafka.Brokers))
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid kafka config",
			config: Config{
				Postgres: PostgresConfig{Host: "localhost", Database: "db", User: "user"},
				Output:   OutputConfig{Type: "kafka", Kafka: &KafkaConfig{Brokers: []string{"localhost:9092"}, Topic: "topic"}},
			},
			wantErr: false,
		},
		{
			name: "valid webhook config",
			config: Config{
				Postgres: PostgresConfig{Host: "localhost", Database: "db", User: "user"},
				Output:   OutputConfig{Type: "webhook", Webhook: &WebhookConfig{URL: "http://example.com", Timeout: 5 * time.Second}},
			},
			wantErr: false,
		},
		{
			name:    "missing postgres host",
			config:  Config{Postgres: PostgresConfig{Database: "db", User: "user"}},
			wantErr: true,
		},
		{
			name:    "invalid output type",
			config:  Config{Postgres: PostgresConfig{Host: "localhost", Database: "db", User: "user"}, Output: OutputConfig{Type: "invalid"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
