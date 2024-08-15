package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/health-analytics-service/api-gateway-health-analytics/config"
	"github.com/segmentio/kafka-go"
)

// Producer produces Kafka messages.
type Producer struct {
	writer *kafka.Writer
	Cfg    config.Config
}

// NewProducer creates a new Producer instance.
func NewProducer(cfg config.Config) *Producer {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.KafkaBrokers...),
		AllowAutoTopicCreation: true,
		RequiredAcks:           kafka.RequireOne,
		Balancer:               &kafka.LeastBytes{},
	}
	return &Producer{writer: writer, Cfg: cfg}
}

// ProduceMessage produces a message to the specified topic with the given key and value.
func (p *Producer) ProduceMessage(ctx context.Context, topic, key string, message interface{}) error {
	value, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: value,
	})
	if err != nil {
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}

	return nil
}

// CreateTopic creates a Kafka topic if it doesn't exist.
func CreateTopic(cfg config.Config, topic string) error {
	// Check if the topic already exists (implementation omitted for brevity)

	conn, err := kafka.DialLeader(context.Background(), "tcp", cfg.KafkaBrokers[0], topic, 0)
	if err != nil {
		return fmt.Errorf("failed to dial leader: %w", err)
	}
	defer conn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = conn.CreateTopics(topicConfigs...)
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	return nil
}
