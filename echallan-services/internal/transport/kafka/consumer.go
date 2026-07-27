package kafka

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

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
	logger       *zap.Logger
}

func NewConsumer(brokers []string, paymentSvc PaymentHandler, filestoreSvc FilestoreHandler, challanSvc ChallanHandler, logger *zap.Logger) *Consumer {
	return &Consumer{
		brokers:      brokers,
		paymentSvc:   paymentSvc,
		filestoreSvc: filestoreSvc,
		challanSvc:   challanSvc,
		logger:       logger,
	}
}

func (c *Consumer) StartListening(ctx context.Context) {
	go c.listenTopic(ctx, "egov.collection.receipt-create", c.handleReceiptCreate, false)
	go c.listenTopic(ctx, "egov.collection.receipt-cancel", c.handleReceiptCancel, false)
	
	if c.filestoreSvc != nil {
		go c.listenTopic(ctx, "pdf-generated", c.handlePdfGenerated, true)
	}
	
	if c.challanSvc != nil {
		go c.listenTopic(ctx, "save-challan", c.handleSaveChallan, false)
		go c.listenTopic(ctx, "update-challan", c.handleUpdateChallan, false)
	}
}

func (c *Consumer) listenTopic(ctx context.Context, topic string, handler func([]byte, map[string]interface{}), passRaw bool) {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:  c.brokers,
		GroupID:  "echallan-services-group",
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer reader.Close()

	c.logger.Info("Started listening to Kafka topic", zap.String("topic", topic))
	
	sem := make(chan struct{}, 100)

	for {
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				c.logger.Info("Stopping kafka listener", zap.String("topic", topic))
				return
			}
			c.logger.Error("Error fetching message from topic", zap.String("topic", topic), zap.Error(err))
			continue
		}

		var record map[string]interface{}
		if !passRaw {
			if err := json.Unmarshal(m.Value, &record); err != nil {
				c.logger.Error("Error unmarshalling message from topic", zap.String("topic", topic), zap.Error(err))
				continue
			}
		}

		sem <- struct{}{}
		go func(msg kafkago.Message, raw []byte, parsed map[string]interface{}) {
			defer func() { <-sem }()
			defer func() {
				if r := recover(); r != nil {
					c.logger.Error("Recovered from panic in kafka handler", zap.String("topic", topic), zap.Any("panic", r))
				}
			}()

			handler(raw, parsed)

			if err := reader.CommitMessages(context.Background(), msg); err != nil {
				c.logger.Error("Failed to commit message", zap.String("topic", topic), zap.Error(err))
			}
		}(m, m.Value, record)
	}
}

func (c *Consumer) handleReceiptCreate(raw []byte, record map[string]interface{}) {
	if err := c.paymentSvc.ProcessPayment(record); err != nil {
		c.logger.Error("Failed to process payment receipt", zap.Error(err))
	}
}

func (c *Consumer) handleReceiptCancel(raw []byte, record map[string]interface{}) {
	if err := c.paymentSvc.ProcessCancel(record); err != nil {
		c.logger.Error("Failed to process cancelled receipt", zap.Error(err))
	}
}

func (c *Consumer) handlePdfGenerated(raw []byte, record map[string]interface{}) {
	if err := c.filestoreSvc.ProcessPdfGenerated(raw); err != nil {
		c.logger.Error("Failed to process pdf generated", zap.Error(err))
	}
}

func (c *Consumer) handleSaveChallan(raw []byte, record map[string]interface{}) {
	if err := c.challanSvc.ProcessSaveChallan(record); err != nil {
		c.logger.Error("Failed to process save challan event", zap.Error(err))
	}
}

func (c *Consumer) handleUpdateChallan(raw []byte, record map[string]interface{}) {
	if err := c.challanSvc.ProcessUpdateChallan(record); err != nil {
		c.logger.Error("Failed to process update challan event", zap.Error(err))
	}
}
