package service

import (
	"testing"

	"github.com/CDPI-HRSS/calci_sp/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestGetCalculationInternal(t *testing.T) {
	calcService := &CalculationService{}

	tests := []struct {
		name          string
		criteriaList  []domain.CalculationCriteria
		expectedTotal float64
		expectError   bool
	}{
		{
			name: "Valid Challan with Multiple Amounts",
			criteriaList: []domain.CalculationCriteria{
				{
					TenantId: "tenant-1",
					Challan: &domain.Challan{
						ChallanNo:       "CH-123",
						BusinessService: "TEST",
						Amount: []domain.Amount{
							{TaxHeadCode: "TAX1", Amount: 100.0},
							{TaxHeadCode: "TAX2", Amount: 50.0},
						},
					},
				},
			},
			expectedTotal: 150.0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestInfo := &domain.RequestInfo{}
			
			calculations, err := calcService.getCalculationInternal(requestInfo, tt.criteriaList)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, calculations, 1)

				totalAmount := 0.0
				for _, est := range calculations[0].TaxHeadEstimates {
					totalAmount += est.EstimateAmount
				}

				assert.Equal(t, tt.expectedTotal, totalAmount)
			}
		})
	}
}
