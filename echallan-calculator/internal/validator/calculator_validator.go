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
	
	for i, c := range req.CalculationCriteria {
		if c.ChallanNo == "" && c.Challan == nil {
			return fmt.Errorf("Either ChallanNo or Challan object is mandatory at criteria index %d", i)
		}
	}
	return nil
}
