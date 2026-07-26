package validator

import (
	"fmt"
	"strings"

	"github.com/CDPI-HRSS/echallan-services/internal/domain"
	"github.com/CDPI-HRSS/echallan-services/internal/repository/http"
)

type ChallanValidator struct {
	mdmsRepo *http.MdmsRepository
	locRepo  *http.LocationRepository
}

func NewChallanValidator(mdmsRepo *http.MdmsRepository, locRepo *http.LocationRepository) *ChallanValidator {
	return &ChallanValidator{
		mdmsRepo: mdmsRepo,
		locRepo:  locRepo,
	}
}

func (v *ChallanValidator) ValidateCreateRequest(req *domain.ChallanRequest) error {
	challan := req.Challan

	// 1. Amount validations
	totalAmt := 0.0
	for _, amt := range challan.Amount {
		totalAmt += amt.Amount
		if amt.Amount < 0 {
			return fmt.Errorf("Amount cannot be negative")
		}
	}
	if totalAmt <= 0 {
		return fmt.Errorf("Challan cannot be generated for zero amount")
	}

	// 2. MDMS Fetch (Financial Year)
	mdmsData, err := v.mdmsRepo.FetchMasterData(req.RequestInfo, challan.TenantId)
	if err != nil {
		return fmt.Errorf("failed to fetch MDMS data: %w", err)
	}

	validFinancialYear := false
	if egovCommon, ok := mdmsData["egov-common-masters"].(map[string]interface{}); ok {
		if fyList, ok := egovCommon["FinancialYear"].([]interface{}); ok {
			for _, fyIntf := range fyList {
				if fy, ok := fyIntf.(map[string]interface{}); ok {
					startDateRaw, ok1 := fy["startingDate"].(float64)
					endDateRaw, ok2 := fy["endingDate"].(float64)
					if ok1 && ok2 {
						startDate := int64(startDateRaw)
						endDate := int64(endDateRaw)
						if challan.TaxPeriodFrom < challan.TaxPeriodTo && challan.TaxPeriodFrom >= startDate && challan.TaxPeriodTo <= endDate {
							validFinancialYear = true
							break
						}
					}
				}
			}
		}
	}
	if !validFinancialYear {
		return fmt.Errorf("Tax period details are invalid compared to MDMS")
	}

	// TaxHeadMaster validation
	if billingService, ok := mdmsData["BillingService"].(map[string]interface{}); ok {
		if taxHeadMasterList, ok := billingService["TaxHeadMaster"].([]interface{}); ok {
			validTaxHeads := make(map[string]bool)
			for _, thIntf := range taxHeadMasterList {
				if th, ok := thIntf.(map[string]interface{}); ok {
					if code, ok := th["code"].(string); ok {
						validTaxHeads[code] = true
					}
				}
			}
			for _, amt := range challan.Amount {
				if amt.TaxHeadCode != "" && !validTaxHeads[amt.TaxHeadCode] {
					return fmt.Errorf("Invalid TaxHeadCode: %s", amt.TaxHeadCode)
				}
			}
		}
	}

	// 3. Location & Date basic validation
	if challan.Address == nil || challan.Address.Locality == nil || challan.Address.Locality.Code == "" {
		return fmt.Errorf("Locality code is mandatory")
	}

	localityCodes, err := v.locRepo.GetLocalityCodes(challan.TenantId, req.RequestInfo)
	if err != nil {
		return fmt.Errorf("failed to fetch locality boundaries from Location Service: %w", err)
	}
	if len(localityCodes) == 0 {
		return fmt.Errorf("no locality boundaries found for tenant %s", challan.TenantId)
	}

	found := false
	for _, code := range localityCodes {
		if code == challan.Address.Locality.Code {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("Locality details are invalid")
	}

	return nil
}

func (v *ChallanValidator) ValidateUpdateRequest(req *domain.ChallanRequest, searchResult []*domain.Challan) error {
	challan := req.Challan

	if len(searchResult) == 0 {
		return fmt.Errorf("The Challan to be updated is not in database")
	}

	searchChallan := searchResult[0]

	if !strings.EqualFold(challan.BusinessService, searchChallan.BusinessService) {
		return fmt.Errorf("The business service is not matching with the Search result")
	}
	if !strings.EqualFold(challan.ChallanNo, searchChallan.ChallanNo) {
		return fmt.Errorf("The Challan Number is not matching with the Search result")
	}
	if challan.Address != nil && searchChallan.Address != nil {
		if !strings.EqualFold(challan.Address.Id, searchChallan.Address.Id) {
			return fmt.Errorf("Address is not matching with the Search result")
		}
	}
	if challan.Citizen != nil && searchChallan.Citizen != nil {
		if !strings.EqualFold(challan.Citizen.Uuid, searchChallan.Citizen.Uuid) {
			return fmt.Errorf("User Details not matching with the Search result")
		}
		if !strings.EqualFold(challan.Citizen.UserName, searchChallan.Citizen.UserName) {
			return fmt.Errorf("User Details not matching with the Search result")
		}
		if !strings.EqualFold(challan.Citizen.MobileNumber, searchChallan.Citizen.MobileNumber) {
			return fmt.Errorf("User Details not matching with the Search result")
		}
	}

	if searchChallan.ApplicationStatus != "ACTIVE" {
		return fmt.Errorf("Challan cannot be updated/cancelled")
	}

	// Re-run standard validation
	return v.ValidateCreateRequest(req)
}
