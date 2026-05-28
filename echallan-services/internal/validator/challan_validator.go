package validator

import (
	"fmt"

	"github.com/CDPI-HRSS/echallan-services/internal/domain"
)

type ChallanValidator struct{}

func NewChallanValidator() *ChallanValidator {
	return &ChallanValidator{}
}

func (v *ChallanValidator) ValidateCreateRequest(req *domain.ChallanRequest) error {
	if req == nil || req.Challan == nil {
		return fmt.Errorf("Challan object is missing from request payload")
	}
	if req.Challan.TenantId == "" {
		return fmt.Errorf("TenantId is mandatory")
	}
	if req.Challan.BusinessService == "" {
		return fmt.Errorf("BusinessService is mandatory")
	}
	return nil
}

func (v *ChallanValidator) ValidateUpdateRequest(req *domain.ChallanRequest) error {
	if req == nil || req.Challan == nil {
		return fmt.Errorf("Challan object is missing from request payload")
	}
	if req.Challan.ChallanNo == "" {
		return fmt.Errorf("ChallanNo is mandatory for update")
	}
	return nil
}
