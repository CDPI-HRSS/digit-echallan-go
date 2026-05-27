package service

import (
	"testing"

	"github.com/CDPI-HRSS/echallan-services/internal/domain"
	"github.com/CDPI-HRSS/echallan-services/internal/transport/kafka"
)

// MockRepo for testing
type mockRepo struct{}

func (m *mockRepo) Search(criteria domain.SearchCriteria) ([]*domain.Challan, error) {
	if len(criteria.ChallanNo) > 0 && criteria.ChallanNo[0] == "MOCK-123" {
		return []*domain.Challan{{ChallanNo: "MOCK-123", TenantId: criteria.TenantId}}, nil
	}
	return []*domain.Challan{}, nil
}

func (m *mockRepo) Count(tenantId string) (int, error) {
	return 1, nil
}

func TestChallanService_Create(t *testing.T) {
	// Mock producer with empty brokers
	producer := kafka.NewProducer([]string{})
	svc := NewChallanService(producer, &mockRepo{})

	req := &domain.ChallanRequest{
		RequestInfo: &domain.RequestInfo{
			UserInfo: &domain.UserInfo{Uuid: "test-uuid"},
		},
		Challan: &domain.Challan{
			TenantId: "pb.amritsar",
		},
	}

	// Attempt create (this will fail pushing to kafka because broker is empty, but logic executes)
	challan, err := svc.Create(req)
	
	// We expect a Kafka error here due to mock, but we want to verify struct enrichment
	if challan != nil {
		if challan.ChallanNo == "" {
			t.Errorf("Expected ChallanNo to be generated")
		}
		if challan.AuditDetails == nil {
			t.Errorf("Expected AuditDetails to be enriched")
		} else if challan.AuditDetails.CreatedBy != "test-uuid" {
			t.Errorf("Expected CreatedBy to be test-uuid, got %s", challan.AuditDetails.CreatedBy)
		}
	} else if err == nil {
		t.Errorf("Expected kafka failure, got nil error")
	}
}

func TestChallanService_Search(t *testing.T) {
	svc := NewChallanService(nil, &mockRepo{})

	// Test happy path
	criteria := domain.SearchCriteria{TenantId: "pb", ChallanNo: []string{"MOCK-123"}}
	res, err := svc.Search(criteria, nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(res) != 1 {
		t.Errorf("Expected 1 result, got %d", len(res))
	}

	// Test empty
	criteria2 := domain.SearchCriteria{TenantId: "pb", ChallanNo: []string{"INVALID"}}
	res2, err := svc.Search(criteria2, nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(res2) != 0 {
		t.Errorf("Expected 0 results, got %d", len(res2))
	}
}
