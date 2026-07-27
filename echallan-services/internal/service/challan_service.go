package service

import (
	"context"
	"github.com/CDPI-HRSS/echallan-services/internal/domain"
)

type ChallanService interface {
	Create(ctx context.Context, req *domain.ChallanRequest) (*domain.Challan, error)
	Search(ctx context.Context, criteria domain.SearchCriteria, reqInfo *domain.RequestInfo) ([]*domain.Challan, int, error)
	Update(ctx context.Context, req *domain.ChallanRequest) (*domain.Challan, error)
	Count(ctx context.Context, tenantId string, reqInfo *domain.RequestInfo) (map[string]interface{}, error)
}
