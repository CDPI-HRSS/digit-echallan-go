package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	brokers []string
	writers sync.Map
}

func NewProducer(brokers []string) *Producer {
	return &Producer{brokers: brokers}
}

// Push publishes an event to the DIGIT event bus for the egov-persister to process.
func (p *Producer) Push(topic string, message interface{}) error {
	var writer *kafka.Writer
	
	if val, ok := p.writers.Load(topic); ok {
		writer = val.(*kafka.Writer)
	} else {
		writer = &kafka.Writer{
			Addr:     kafka.TCP(p.brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		}
		p.writers.Store(topic, writer)
	}

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
