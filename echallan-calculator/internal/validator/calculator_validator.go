package validator

import (
	"fmt"

	"github.com/CDPI-HRSS/calci_sp/internal/domain"
)

type CalculatorValidator struct{}

func NewCalculatorValidator() *CalculatorValidator {
	return &CalculatorValidator{}
}

func (v *CalculatorValidator) ValidateCalculationReq(req *domain.CalculationReq) error {
	if req == nil {
		return fmt.Errorf("CalculationReq is missing from request payload")
	}
	if len(req.CalculationCriteria) == 0 {
		return fmt.Errorf("CalculationCriteria is missing or empty")
	}
	
	for i, c := range req.CalculationCriteria {
		if c.TenantId == "" {
			return fmt.Errorf("TenantId is mandatory at criteria index %d", i)
		}
		if c.ChallanNo == "" && c.Challan == nil {
			return fmt.Errorf("Either ChallanNo or Challan object is mandatory at criteria index %d", i)
		}
	}
	return nil
}
