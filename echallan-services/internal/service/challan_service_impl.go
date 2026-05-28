package service

import (
	"fmt"
	"time"

	"github.com/CDPI-HRSS/echallan-services/internal/domain"
	"github.com/CDPI-HRSS/echallan-services/internal/repository/postgres"
	"github.com/CDPI-HRSS/echallan-services/internal/transport/kafka"
	"github.com/CDPI-HRSS/echallan-services/internal/validator"
)

type challanServiceImpl struct {
	producer  *kafka.Producer
	repo      postgres.ChallanRepository
	validator *validator.ChallanValidator
}

func NewChallanService(producer *kafka.Producer, repo postgres.ChallanRepository, val *validator.ChallanValidator) ChallanService {
	return &challanServiceImpl{
		producer:  producer,
		repo:      repo,
		validator: val,
	}
}

func (s *challanServiceImpl) Create(req *domain.ChallanRequest) (*domain.Challan, error) {
	// 1. Validation Logic
	if err := s.validator.ValidateCreateRequest(req); err != nil {
		return nil, err
	}
	// This would typically involve validating against MDMS Master Data
	
	// 2. ID Generation Logic
	// In the real DIGIT setup, this makes an HTTP call to egov-idgen. 
	// Scaffolded here for structural integrity.
	req.Challan.ChallanNo = fmt.Sprintf("CH-%d", time.Now().Unix())
	req.Challan.Id = fmt.Sprintf("UUID-%d", time.Now().UnixNano())
	
	// 3. Audit Details Enrichment
	var uuid string
	if req.RequestInfo != nil && req.RequestInfo.UserInfo != nil {
		uuid = req.RequestInfo.UserInfo.Uuid
	}
	req.Challan.AuditDetails = &domain.AuditDetails{
		CreatedBy:        uuid,
		LastModifiedBy:   uuid,
		CreatedTime:      time.Now().UnixMilli(),
		LastModifiedTime: time.Now().UnixMilli(),
	}

	// 4. DIGIT Persister Pattern: Push to Kafka instead of direct DB insert
	err := s.producer.Push("save-challan", req)
	if err != nil {
		return nil, fmt.Errorf("failed to persist challan: %w", err)
	}

	return req.Challan, nil
}

func (s *challanServiceImpl) Update(req *domain.ChallanRequest) (*domain.Challan, error) {
	// 1. Validation and fetch existing (omitted)
	if err := s.validator.ValidateUpdateRequest(req); err != nil {
		return nil, err
	}

	// 2. Audit Details Enrichment
	var uuid string
	if req.RequestInfo != nil && req.RequestInfo.UserInfo != nil {
		uuid = req.RequestInfo.UserInfo.Uuid
	}
	req.Challan.AuditDetails = &domain.AuditDetails{
		LastModifiedBy:   uuid,
		LastModifiedTime: time.Now().UnixMilli(),
	}

	// 3. DIGIT Persister Pattern: Push to Kafka for egov-persister
	err := s.producer.Push("update-challan", req)
	if err != nil {
		return nil, fmt.Errorf("failed to persist challan update: %w", err)
	}

	return req.Challan, nil
}

func (s *challanServiceImpl) Search(criteria domain.SearchCriteria, reqInfo *domain.RequestInfo) ([]*domain.Challan, error) {
	return s.repo.Search(criteria)
}

func (s *challanServiceImpl) Count(tenantId string, reqInfo *domain.RequestInfo) (map[string]interface{}, error) {
	count, err := s.repo.Count(tenantId)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"totalCount": count,
	}, nil
}
