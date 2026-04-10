package pipeline_test

import (
	"testing"

	"github.com/pgstream/pgstream/internal/config"
	"github.com/pgstream/pgstream/internal/pipeline"
)

func TestBuild_NilConfig(t *testing.T) {
	_, err := pipeline.Build(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestBuild_NoSinkConfigured(t *testing.T) {
	cfg := &config.Config{}
	cfg.Postgres.DSN = "postgres://localhost/test"
	// Neither Kafka nor Webhook configured.
	_, err := pipeline.Build(cfg, nil)
	if err == nil {
		t.Fatal("expected error when no sink is configured")
	}
}

func TestBuild_InvalidKafkaSink(t *testing.T) {
	cfg := &config.Config{}
	cfg.Postgres.DSN = "postgres://localhost/test"
	// Brokers slice is non-nil but topic is empty — NewKafkaSink should fail.
	cfg.Kafka.Brokers = []string{"localhost:9092"}
	cfg.Kafka.Topic = ""

	_, err := pipeline.Build(cfg, nil)
	if err == nil {
		t.Fatal("expected error for missing kafka topic")
	}
}

func TestBuild_InvalidWebhookSink(t *testing.T) {
	cfg := &config.Config{}
	cfg.Postgres.DSN = "postgres://localhost/test"
	// URL is non-empty but malformed — NewWebhookSink validates scheme.
	cfg.Webhook.URL = "not-a-url"

	_, err := pipeline.Build(cfg, nil)
	if err == nil {
		t.Fatal("expected error for invalid webhook URL")
	}
}
