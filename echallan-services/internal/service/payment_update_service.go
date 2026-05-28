package service

import (
	"fmt"
	"log"

	"github.com/CDPI-HRSS/echallan-services/internal/domain"
	"github.com/CDPI-HRSS/echallan-services/internal/transport/kafka"
)

type PaymentUpdateService interface {
	ProcessPayment(record map[string]interface{}) error
	ProcessCancel(record map[string]interface{}) error
}

type paymentUpdateServiceImpl struct {
	producer *kafka.Producer
}

func NewPaymentUpdateService(producer *kafka.Producer) PaymentUpdateService {
	return &paymentUpdateServiceImpl{
		producer: producer,
	}
}

func (s *paymentUpdateServiceImpl) ProcessPayment(record map[string]interface{}) error {
	// 1. Extract receipt payload
	// 2. Map receipt to Challan status update (PAID)
	// 3. Push to update-challan topic for persister to save
	
	log.Printf("Processing payment update for receipt...")
	
	// Mocked implementation to satisfy architectural parity
	challanReq := &domain.ChallanRequest{
		RequestInfo: &domain.RequestInfo{},
		Challan: &domain.Challan{
			ApplicationStatus: "PAID",
		},
	}
	
	err := s.producer.Push("update-challan", challanReq)
	if err != nil {
		return fmt.Errorf("failed to push payment update: %w", err)
	}
	return nil
}

func (s *paymentUpdateServiceImpl) ProcessCancel(record map[string]interface{}) error {
	log.Printf("Processing payment cancellation...")
	
	challanReq := &domain.ChallanRequest{
		RequestInfo: &domain.RequestInfo{},
		Challan: &domain.Challan{
			ApplicationStatus: "CANCELLED",
		},
	}
	
	err := s.producer.Push("update-challan", challanReq)
	if err != nil {
		return fmt.Errorf("failed to push payment cancel: %w", err)
	}
	return nil
}
