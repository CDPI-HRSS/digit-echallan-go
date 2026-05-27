package service

import (
	"github.com/CDPI-HRSS/echallan-services/internal/domain"
)

type ChallanService interface {
	Create(req *domain.ChallanRequest) (*domain.Challan, error)
	Search(criteria domain.SearchCriteria, reqInfo *domain.RequestInfo) ([]*domain.Challan, error)
	Update(req *domain.ChallanRequest) (*domain.Challan, error)
	Count(tenantId string, reqInfo *domain.RequestInfo) (map[string]interface{}, error)
}
