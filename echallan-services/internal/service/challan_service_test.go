package service

import (
	"testing"

	"github.com/CDPI-HRSS/echallan-services/internal/domain"
	"github.com/CDPI-HRSS/echallan-services/internal/transport/kafka"
	"github.com/CDPI-HRSS/echallan-services/internal/validator"
)

// MockRepo for testing
type mockRepo struct{}

func (m *mockRepo) Search(criteria domain.SearchCriteria) ([]*domain.Challan, int, error) {
	if criteria.ChallanNo == "MOCK-123" {
		return []*domain.Challan{{ChallanNo: "MOCK-123", TenantId: criteria.TenantId}}, 1, nil
	}
	return []*domain.Challan{}, 0, nil
}

func (m *mockRepo) Count(tenantId string) (map[string]int, error) {
	return map[string]int{"TOTAL": 1}, nil
}

func TestChallanService_Create(t *testing.T) {
	producer := kafka.NewProducer([]string{})
	svc := NewChallanService(producer, &mockRepo{}, validator.NewChallanValidator(nil, nil), nil, nil, nil, nil)

	req := &domain.ChallanRequest{
		RequestInfo: &domain.RequestInfo{
			UserInfo: &domain.UserInfo{Uuid: "test-uuid"},
		},
		Challan: &domain.Challan{
			TenantId: "pb.amritsar",
		},
	}

	challan, err := svc.Create(req)
	
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
	svc := NewChallanService(nil, &mockRepo{}, validator.NewChallanValidator(nil, nil), nil, nil, nil, nil)

	// Test happy path
	criteria := domain.SearchCriteria{TenantId: "pb", ChallanNo: "MOCK-123"}
	res, _, err := svc.Search(criteria, nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(res) != 1 {
		t.Errorf("Expected 1 result, got %d", len(res))
	}

	// Test empty
	criteria2 := domain.SearchCriteria{TenantId: "pb", ChallanNo: "INVALID"}
	res2, _, err := svc.Search(criteria2, nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(res2) != 0 {
		t.Errorf("Expected 0 results, got %d", len(res2))
	}
}
