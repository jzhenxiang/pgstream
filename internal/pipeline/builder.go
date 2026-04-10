package pipeline

import (
	"fmt"
	"log/slog"

	"github.com/pgstream/pgstream/internal/config"
	"github.com/pgstream/pgstream/internal/filter"
	"github.com/pgstream/pgstream/internal/metrics"
	"github.com/pgstream/pgstream/internal/sink"
	"github.com/pgstream/pgstream/internal/transform"
	"github.com/pgstream/pgstream/internal/wal"
)

// Build constructs a fully wired Pipeline from application config.
func Build(cfg *config.Config, logger *slog.Logger) (*Pipeline, error) {
	if cfg == nil {
		return nil, fmt.Errorf("builder: config must not be nil")
	}

	reader, err := wal.NewReader(cfg.Postgres.DSN)
	if err != nil {
		return nil, fmt.Errorf("builder: wal reader: %w", err)
	}

	f := filter.New(cfg.Filter)
	tr := transform.New(cfg.Transform)
	m := metrics.New()

	var s sink.Sink
	switch {
	case cfg.Kafka.Brokers != nil:
		s, err = sink.NewKafkaSink(cfg.Kafka)
		if err != nil {
			return nil, fmt.Errorf("builder: kafka sink: %w", err)
		}
	case cfg.Webhook.URL != "":
		s, err = sink.NewWebhookSink(cfg.Webhook)
		if err != nil {
			return nil, fmt.Errorf("builder: webhook sink: %w", err)
		}
	default:
		return nil, fmt.Errorf("builder: no sink configured; set kafka or webhook options")
	}

	return New(Config{
		Reader:    reader,
		Filter:    f,
		Transform: tr,
		Sink:      s,
		Metrics:   m,
		Logger:    logger,
	})
}
