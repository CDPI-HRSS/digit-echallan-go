package kafka

import (
	"context"
	"encoding/json"
	"log"

	kafkago "github.com/segmentio/kafka-go"
)

type PaymentHandler interface {
	ProcessPayment(record map[string]interface{}) error
	ProcessCancel(record map[string]interface{}) error
}

type FilestoreHandler interface {
	ProcessPdfGenerated(payload []byte) error
}

type ChallanHandler interface {
	ProcessSaveChallan(payload map[string]interface{}) error
	ProcessUpdateChallan(payload map[string]interface{}) error
}

type Consumer struct {
	brokers      []string
	paymentSvc   PaymentHandler
	filestoreSvc FilestoreHandler
	challanSvc   ChallanHandler
}

func NewConsumer(brokers []string, paymentSvc PaymentHandler, filestoreSvc FilestoreHandler, challanSvc ChallanHandler) *Consumer {
	return &Consumer{
		brokers:      brokers,
		paymentSvc:   paymentSvc,
		filestoreSvc: filestoreSvc,
		challanSvc:   challanSvc,
	}
}

func (c *Consumer) StartListening() {
	go c.listenTopic("egov.collection.receipt-create", c.handleReceiptCreate, false)
	go c.listenTopic("egov.collection.receipt-cancel", c.handleReceiptCancel, false)
	
	if c.filestoreSvc != nil {
		go c.listenTopic("pdf-generated", c.handlePdfGenerated, true)
	}
	
	if c.challanSvc != nil {
		go c.listenTopic("save-challan", c.handleSaveChallan, false)
		go c.listenTopic("update-challan", c.handleUpdateChallan, false)
	}
}

func (c *Consumer) listenTopic(topic string, handler func([]byte, map[string]interface{}), passRaw bool) {
	ctx := context.Background() // TODO: pass in global shutdown context
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:  c.brokers,
		GroupID:  "echallan-services-group",
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer reader.Close()

	log.Printf("Started listening to Kafka topic: %s", topic)
	
	sem := make(chan struct{}, 100)

	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading message from topic %s: %v", topic, err)
			continue
		}

		var record map[string]interface{}
		if !passRaw {
			if err := json.Unmarshal(m.Value, &record); err != nil {
				log.Printf("Error unmarshalling message from topic %s: %v", topic, err)
				continue
			}
		}

		sem <- struct{}{}
		go func(raw []byte, parsed map[string]interface{}) {
			defer func() { <-sem }()
			handler(raw, parsed)
		}(m.Value, record)
	}
}

func (c *Consumer) handleReceiptCreate(raw []byte, record map[string]interface{}) {
	if err := c.paymentSvc.ProcessPayment(record); err != nil {
		log.Printf("Failed to process payment receipt: %v", err)
	}
}

func (c *Consumer) handleReceiptCancel(raw []byte, record map[string]interface{}) {
	if err := c.paymentSvc.ProcessCancel(record); err != nil {
		log.Printf("Failed to process cancelled receipt: %v", err)
	}
}

func (c *Consumer) handlePdfGenerated(raw []byte, record map[string]interface{}) {
	if err := c.filestoreSvc.ProcessPdfGenerated(raw); err != nil {
		log.Printf("Failed to process pdf generated: %v", err)
	}
}

func (c *Consumer) handleSaveChallan(raw []byte, record map[string]interface{}) {
	if err := c.challanSvc.ProcessSaveChallan(record); err != nil {
		log.Printf("Failed to process save challan event: %v", err)
	}
}

func (c *Consumer) handleUpdateChallan(raw []byte, record map[string]interface{}) {
	if err := c.challanSvc.ProcessUpdateChallan(record); err != nil {
		log.Printf("Failed to process update challan event: %v", err)
	}
}
