package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	brokers []string
}

func NewProducer(brokers []string) *Producer {
	return &Producer{brokers: brokers}
}

// Push publishes an event to the DIGIT event bus for the egov-persister to process.
func (p *Producer) Push(topic string, message interface{}) error {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(p.brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	bytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = writer.WriteMessages(context.Background(),
		kafka.Message{
			Value: bytes,
		},
	)
	if err != nil {
		log.Printf("Failed to push message to topic %s: %v", topic, err)
		return err
	}

	log.Printf("Successfully pushed message to topic %s", topic)
	return nil
}
