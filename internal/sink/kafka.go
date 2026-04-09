package sink

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

// KafkaSink sends WAL events to a Kafka topic.
type KafkaSink struct {
	writer *kafka.Writer
	topic  string
}

// NewKafkaSink creates a new KafkaSink with the given brokers and topic.
func NewKafkaSink(brokers []string, topic string) (*KafkaSink, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka sink: at least one broker address is required")
	}
	if topic == "" {
		return nil, fmt.Errorf("kafka sink: topic is required")
	}

	w := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		AllowAutoTopicCreation: true,
		Balancer:               &kafka.LeastBytes{},
	}

	return &KafkaSink{writer: w, topic: topic}, nil
}

// Send encodes the event as JSON and publishes it to the Kafka topic.
func (k *KafkaSink) Send(ctx context.Context, event any) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("kafka sink: failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Value: payload,
	}

	if err := k.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("kafka sink: failed to write message: %w", err)
	}

	return nil
}

// Close closes the underlying Kafka writer.
func (k *KafkaSink) Close() error {
	if err := k.writer.Close(); err != nil {
		return fmt.Errorf("kafka sink: failed to close writer: %w", err)
	}
	return nil
}
