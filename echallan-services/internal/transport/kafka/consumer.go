package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

type PaymentHandler interface {
	ProcessPayment(record map[string]interface{}) error
	ProcessCancel(record map[string]interface{}) error
}

type Consumer struct {
	brokers      []string
	paymentSvc   PaymentHandler
}

func NewConsumer(brokers []string, paymentSvc PaymentHandler) *Consumer {
	return &Consumer{
		brokers:    brokers,
		paymentSvc: paymentSvc,
	}
}

func (c *Consumer) StartListening() {
	go c.listenTopic("egov.collection.receipt-create", c.handleReceiptCreate)
	go c.listenTopic("egov.collection.receipt-cancel", c.handleReceiptCancel)
}

func (c *Consumer) listenTopic(topic string, handler func(map[string]interface{})) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.brokers,
		GroupID:  "echallan-services-group",
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer reader.Close()

	log.Printf("Started listening to Kafka topic: %s", topic)
	
	// Bounded concurrency (semaphore)
	sem := make(chan struct{}, 100)

	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message from topic %s: %v", topic, err)
			continue
		}

		var record map[string]interface{}
		if err := json.Unmarshal(m.Value, &record); err != nil {
			log.Printf("Error unmarshalling message from topic %s: %v", topic, err)
			continue
		}

		// Process async with bounded concurrency
		sem <- struct{}{}
		go func(r map[string]interface{}) {
			defer func() { <-sem }()
			handler(r)
		}(record)
	}
}

func (c *Consumer) handleReceiptCreate(record map[string]interface{}) {
	if err := c.paymentSvc.ProcessPayment(record); err != nil {
		log.Printf("Failed to process payment receipt: %v", err)
	}
}

func (c *Consumer) handleReceiptCancel(record map[string]interface{}) {
	if err := c.paymentSvc.ProcessCancel(record); err != nil {
		log.Printf("Failed to process cancelled receipt: %v", err)
	}
}
