package pipeline

import (
	"errors"
	"fmt"

	"github.com/pgstream/pgstream/internal/config"
	"github.com/pgstream/pgstream/internal/healthcheck"
	"github.com/pgstream/pgstream/internal/sink"
	"github.com/pgstream/pgstream/internal/sink/kafka"
	"github.com/pgstream/pgstream/internal/sink/webhook"
	"github.com/pgstream/pgstream/internal/wal"
)

// Build constructs a Pipeline from the provided configuration, wiring together
// the WAL reader, decoder, sink, and optional health check server.
func Build(cfg *config.Config) (*Pipeline, error) {
	if cfg == nil {
		return nil, errors.New("pipeline: config must not be nil")
	}

	var s sink.Sink
	switch {
	case cfg.Kafka.Brokers != "":
		ks, err := kafka.NewKafkaSink(cfg.Kafka)
		if err != nil {
			return nil, fmt.Errorf("pipeline: kafka sink: %w", err)
		}
		s = ks
	case cfg.Webhook.URL != "":
		ws, err := webhook.NewWebhookSink(cfg.Webhook)
		if err != nil {
			return nil, fmt.Errorf("pipeline: webhook sink: %w", err)
		}
		s = ws
	default:
		return nil, errors.New("pipeline: no sink configured (set kafka.brokers or webhook.url)")
	}

	reader, err := wal.NewReader(cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("pipeline: wal reader: %w", err)
	}

	var hs *healthcheck.Server
	if cfg.HealthCheck.Addr != "" {
		hs = healthcheck.New(cfg.HealthCheck.Addr, cfg.Version)
	}

	return New(reader, s, hs)
}
