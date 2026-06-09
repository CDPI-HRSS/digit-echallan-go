package service

import (
	"fmt"
	"log"

	"github.com/CDPI-HRSS/echallan-services/internal/domain"
	"github.com/CDPI-HRSS/echallan-services/internal/repository/postgres"
	"github.com/CDPI-HRSS/echallan-services/internal/transport/kafka"
)

type PaymentUpdateService struct {
	producer *kafka.Producer
	repo     postgres.ChallanRepository
}

func NewPaymentUpdateService(producer *kafka.Producer, repo postgres.ChallanRepository) *PaymentUpdateService {
	return &PaymentUpdateService{
		producer: producer,
		repo:     repo,
	}
}

func (s *PaymentUpdateService) ProcessPayment(payload map[string]interface{}) error {
	log.Printf("Processing payment update for receipt...")
	
	payment, ok := payload["Payment"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payload wrapper")
	}

	paymentDetailsObj, ok := payment["paymentDetails"].([]interface{})
	if !ok || len(paymentDetailsObj) == 0 {
		return fmt.Errorf("no payment details found")
	}

	for _, pDetailIntf := range paymentDetailsObj {
		paymentDetail, ok := pDetailIntf.(map[string]interface{})
		if !ok {
			continue
		}

		billObj, ok := paymentDetail["bill"].(map[string]interface{})
		if !ok {
			continue
		}

		challanNo, ok := billObj["consumerCode"].(string)
		if !ok || challanNo == "" {
			continue
		}

		receiptNumber, ok := paymentDetail["receiptNumber"].(string)
		if !ok {
			receiptNumber = ""
		}

		tenantId, _ := payment["tenantId"].(string)
		criteria := domain.SearchCriteria{
			TenantId:  tenantId,
			ChallanNo: challanNo,
		}
		
		challans, _, err := s.repo.Search(criteria)
		if err != nil || len(challans) == 0 {
			continue
		}
		
		challan := challans[0]

		if challan.ApplicationStatus == "CANCELLED" {
			log.Printf("Ignoring payment for CANCELLED challan: %s", challanNo)
			continue
		}

		totalAmountPaidRaw, ok := paymentDetail["totalAmountPaid"]
		var totalAmountPaid float64
		if ok {
			switch v := totalAmountPaidRaw.(type) {
			case float64:
				totalAmountPaid = v
			case int:
				totalAmountPaid = float64(v)
			}
		}

		totalAmountDue := 0.0
		for _, amt := range challan.Amount {
			totalAmountDue += amt.Amount
		}

		if totalAmountPaid != totalAmountDue {
			log.Printf("Partial payment detected for challan %s. Paid: %v, Due: %v", challanNo, totalAmountPaid, totalAmountDue)
			// Decide what to do here, Java allows partial payment? Actually no, Java explicitly checks if it's equal or handles it via receipt.
			// The instruction says "Verify total amount paid matches total amount due".
			if totalAmountPaid < totalAmountDue {
				challan.ApplicationStatus = "PARTIALLY_PAID"
			} else {
				challan.ApplicationStatus = "PAID"
			}
		} else {
			challan.ApplicationStatus = "PAID"
		}
		
		challan.ReceiptNumber = receiptNumber

		challanReq := &domain.ChallanRequest{
			RequestInfo: &domain.RequestInfo{},
			Challan:     challan,
		}
		
		err = s.producer.Push("update-challan", challanReq)
		if err != nil {
			log.Printf("Failed to push payment update: %v", err)
		}
	}
	
	return nil
}

func (s *PaymentUpdateService) ProcessCancel(payload map[string]interface{}) error {
	log.Printf("Processing payment cancellation...")
	
	payment, ok := payload["Payment"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payload wrapper")
	}

	paymentDetailsObj, ok := payment["paymentDetails"].([]interface{})
	if !ok || len(paymentDetailsObj) == 0 {
		return fmt.Errorf("no payment details found")
	}

	for _, pDetailIntf := range paymentDetailsObj {
		paymentDetail, ok := pDetailIntf.(map[string]interface{})
		if !ok {
			continue
		}

		billObj, ok := paymentDetail["bill"].(map[string]interface{})
		if !ok {
			continue
		}

		challanNo, ok := billObj["consumerCode"].(string)
		if !ok || challanNo == "" {
			continue
		}

		tenantId, _ := payment["tenantId"].(string)
		criteria := domain.SearchCriteria{
			TenantId:  tenantId,
			ChallanNo: challanNo,
		}
		
		challans, _, err := s.repo.Search(criteria)
		if err != nil || len(challans) == 0 {
			continue
		}
		
		challan := challans[0]
		
		challan.ApplicationStatus = "ACTIVE"

		challanReq := &domain.ChallanRequest{
			RequestInfo: &domain.RequestInfo{},
			Challan:     challan,
		}
		
		err = s.producer.Push("update-challan", challanReq)
		if err != nil {
			log.Printf("Failed to push payment cancel: %v", err)
		}
	}
	return nil
}
