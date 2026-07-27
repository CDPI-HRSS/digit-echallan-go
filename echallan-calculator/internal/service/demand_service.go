package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/CDPI-HRSS/echallan-calculator/configs"
	"github.com/CDPI-HRSS/echallan-calculator/internal/domain"
	"github.com/CDPI-HRSS/echallan-calculator/internal/repository/http"
	"github.com/CDPI-HRSS/echallan-calculator/internal/util"
	"github.com/shopspring/decimal"
)

const MDMSRoundOffTaxHead = "_ROUNDOFF"

type DemandService struct {
	cfg        *configs.Config
	utils      *util.CalculationUtils
	srRepo     *http.ServiceRequestRepository
	demandRepo *http.DemandRepository
}

func NewDemandService(cfg *configs.Config, utils *util.CalculationUtils, srRepo *http.ServiceRequestRepository, demandRepo *http.DemandRepository) *DemandService {
	return &DemandService{
		cfg:        cfg,
		utils:      utils,
		srRepo:     srRepo,
		demandRepo: demandRepo,
	}
}

func (s *DemandService) GenerateDemand(ctx context.Context, requestInfo *domain.RequestInfo, calculations []domain.Calculation, businessService string) error {
	if len(calculations) == 0 {
		return nil
	}

	var createCalculations []domain.Calculation
	var updateCalculations []domain.Calculation

	tenantId := calculations[0].TenantId
	var consumerCodes []string
	for _, calc := range calculations {
		if calc.Challan != nil {
			consumerCodes = append(consumerCodes, calc.Challan.ChallanNo)
		}
	}

	demands, err := s.SearchDemand(ctx, tenantId, consumerCodes, requestInfo, businessService)
	if err != nil {
		return fmt.Errorf("failed to search demands: %w", err)
	}

	appNoFromDemands := make(map[string]bool)
	for _, demand := range demands {
		appNoFromDemands[demand.ConsumerCode] = true
	}

	for _, calc := range calculations {
		var challanNo string
		if calc.Challan != nil {
			challanNo = calc.Challan.ChallanNo
		} else {
			challanNo = calc.ChallanNo
		}
		if !appNoFromDemands[challanNo] {
			createCalculations = append(createCalculations, calc)
		} else {
			updateCalculations = append(updateCalculations, calc)
		}
	}

	if len(createCalculations) > 0 {
		_, err := s.createDemand(ctx, requestInfo, createCalculations)
		if err != nil {
			return fmt.Errorf("failed to create demands: %w", err)
		}
	}

	if len(updateCalculations) > 0 {
		_, err := s.updateDemand(ctx, requestInfo, updateCalculations, businessService)
		if err != nil {
			return fmt.Errorf("failed to update demands: %w", err)
		}
	}

	for _, calc := range calculations {
		if calc.Challan != nil && calc.Challan.ApplicationStatus != "CANCELLED" {
			billCriteria := domain.GenerateBillCriteria{
				TenantId:        calc.TenantId,
				ConsumerCode:    calc.Challan.ChallanNo,
				BusinessService: businessService,
			}
			_, err := s.GenerateBill(ctx, requestInfo, billCriteria)
			if err != nil {
				log.Printf("Warning: failed to generate bill for %s: %v\n", calc.Challan.ChallanNo, err)
				return fmt.Errorf("failed to generate bill: %w", err)
			}
		}
	}

	return nil
}

func (s *DemandService) createDemand(ctx context.Context, requestInfo *domain.RequestInfo, calculations []domain.Calculation) ([]domain.Demand, error) {
	var demands []domain.Demand
	for _, calc := range calculations {
		challan := calc.Challan
		if challan == nil && calc.ChallanNo != "" {
			var err error
			challan, err = s.utils.GetChallan(ctx, requestInfo, calc.ChallanNo, calc.TenantId)
			if err != nil {
				return nil, err
			}
		}

		if challan == nil {
			return nil, fmt.Errorf("INVALID APPLICATIONNUMBER: Demand cannot be generated for applicationNumber %s; challan with this number does not exist", calc.ChallanNo)
		}

		tenantId := calc.TenantId
		consumerCode := challan.ChallanNo

		var payer *domain.User
		if challan.Citizen != nil {
			payer = challan.Citizen.ToCommonUser()
		}

		var demandDetails []domain.DemandDetail
		for _, estimate := range calc.TaxHeadEstimates {
			demandDetails = append(demandDetails, domain.DemandDetail{
				TaxAmount:         estimate.EstimateAmount,
				TaxHeadMasterCode: estimate.TaxHeadCode,
				CollectionAmount:  0.0,
				TenantId:          tenantId,
			})
		}

		taxPeriodFrom := challan.TaxPeriodFrom
		taxPeriodTo := challan.TaxPeriodTo
		businessService := challan.BusinessService

		s.addRoundOffTaxHead(calc.TenantId, &demandDetails, businessService)

		singleDemand := domain.Demand{
			ConsumerCode:         consumerCode,
			DemandDetails:        demandDetails,
			Payer:                payer,
			TenantId:             tenantId,
			TaxPeriodFrom:        taxPeriodFrom,
			TaxPeriodTo:          taxPeriodTo,
			ConsumerType:         "challan",
			BusinessService:      businessService,
			MinimumAmountPayable: 0.0,
		}

		demands = append(demands, singleDemand)
	}

	return s.demandRepo.SaveDemand(ctx, requestInfo, demands)
}

func (s *DemandService) updateDemand(ctx context.Context, requestInfo *domain.RequestInfo, calculations []domain.Calculation, businessService string) ([]domain.Demand, error) {
	var demands []domain.Demand
	for _, calc := range calculations {
		challan := calc.Challan
		if challan == nil {
			return nil, fmt.Errorf("challan is nil in update calculation")
		}

		searchResult, err := s.SearchDemand(ctx, calc.TenantId, []string{challan.ChallanNo}, requestInfo, businessService)
		if err != nil {
			return nil, err
		}

		if len(searchResult) == 0 {
			return nil, fmt.Errorf("INVALID UPDATE: No demand exists for applicationNumber: %s", challan.ChallanNo)
		}

		demand := searchResult[0]
		if challan.ApplicationStatus == "CANCELLED" {
			demand.Status = "CANCELLED"
		}

		updatedDemandDetails := s.getUpdatedDemandDetails(calc, demand.DemandDetails)
		demand.DemandDetails = updatedDemandDetails
		demands = append(demands, demand)
	}

	return s.demandRepo.UpdateDemand(ctx, requestInfo, demands)
}

func (s *DemandService) SearchDemand(ctx context.Context, tenantId string, consumerCodes []string, requestInfo *domain.RequestInfo, businessService string) ([]domain.Demand, error) {
	if len(consumerCodes) == 0 {
		return nil, nil
	}

	codesStr := strings.Join(consumerCodes, ",")
	url := s.utils.GetDemandSearchURL(tenantId, businessService, codesStr)

	wrapper := domain.RequestInfoWrapper{
		RequestInfo: requestInfo,
	}

	var response domain.DemandResponse
	err := s.srRepo.FetchResult(ctx, url, wrapper, &response)
	if err != nil {
		return nil, err
	}

	return response.Demands, nil
}

func (s *DemandService) GenerateBill(ctx context.Context, requestInfo *domain.RequestInfo, billCriteria domain.GenerateBillCriteria) (*domain.BillResponse, error) {
	demands, err := s.SearchDemand(ctx, billCriteria.TenantId, []string{billCriteria.ConsumerCode}, requestInfo, billCriteria.BusinessService)
	if err != nil {
		return nil, err
	}

	if len(demands) == 0 {
		return nil, fmt.Errorf("INVALID CONSUMERCODE: Bill cannot be generated. No demand exists for the given consumerCode")
	}

	url := s.utils.GetBillGenerateURI(billCriteria.TenantId, billCriteria.ConsumerCode, billCriteria.BusinessService)

	wrapper := domain.RequestInfoWrapper{
		RequestInfo: requestInfo,
	}

	var response domain.BillResponse
	err = s.srRepo.FetchResult(ctx, url, wrapper, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (s *DemandService) getUpdatedDemandDetails(calculation domain.Calculation, demandDetails []domain.DemandDetail) []domain.DemandDetail {
	var newDemandDetails []domain.DemandDetail
	taxHeadToDemandDetails := make(map[string][]domain.DemandDetail)

	for _, detail := range demandDetails {
		taxHeadToDemandDetails[detail.TaxHeadMasterCode] = append(taxHeadToDemandDetails[detail.TaxHeadMasterCode], detail)
	}

	for _, estimate := range calculation.TaxHeadEstimates {
		details, exists := taxHeadToDemandDetails[estimate.TaxHeadCode]
		if !exists {
			newDemandDetails = append(newDemandDetails, domain.DemandDetail{
				TaxAmount:         estimate.EstimateAmount,
				TaxHeadMasterCode: estimate.TaxHeadCode,
				TenantId:          calculation.TenantId,
				CollectionAmount:  0.0,
			})
		} else {
			total := decimal.NewFromFloat(0)
			for _, d := range details {
				total = total.Add(decimal.NewFromFloat(d.TaxAmount))
			}
			estAmt := decimal.NewFromFloat(estimate.EstimateAmount)
			diffInTaxAmount := estAmt.Sub(total)
			if !diffInTaxAmount.IsZero() {
				diffFloat, _ := diffInTaxAmount.Float64()
				newDemandDetails = append(newDemandDetails, domain.DemandDetail{
					TaxAmount:         diffFloat,
					TaxHeadMasterCode: estimate.TaxHeadCode,
					TenantId:          calculation.TenantId,
					CollectionAmount:  0.0,
				})
			}
		}
	}

	combined := append([]domain.DemandDetail(nil), demandDetails...)
	combined = append(combined, newDemandDetails...)

	s.addRoundOffTaxHead(calculation.TenantId, &combined, calculation.Challan.BusinessService)
	return combined
}

func (s *DemandService) addRoundOffTaxHead(tenantId string, demandDetails *[]domain.DemandDetail, businessService string) {
	totalTax := decimal.NewFromFloat(0)
	var prevRoundOffDemandDetail *domain.DemandDetail

	roundOffTaxHeadCode := businessService + MDMSRoundOffTaxHead

	for i := range *demandDetails {
		detail := &(*demandDetails)[i]
		if !strings.EqualFold(detail.TaxHeadMasterCode, roundOffTaxHeadCode) {
			totalTax = totalTax.Add(decimal.NewFromFloat(detail.TaxAmount))
		} else {
			prevRoundOffDemandDetail = detail
		}
	}

	one := decimal.NewFromInt(1)
	decimalValue := totalTax.Sub(decimal.NewFromInt(totalTax.IntPart()))
	midVal := decimal.NewFromFloat(0.5)
	roundOff := decimal.NewFromFloat(0)

	if decimalValue.Cmp(midVal) >= 0 {
		roundOff = one.Sub(decimalValue)
	} else if decimalValue.Cmp(midVal) < 0 {
		roundOff = decimalValue.Neg()
	}

	if totalTax.Cmp(decimal.NewFromInt(0)) < 0 && roundOff.Cmp(decimal.NewFromInt(0)) > 0 {
		roundOff = roundOff.Sub(one)
	}

	if prevRoundOffDemandDetail != nil {
		prev := decimal.NewFromFloat(prevRoundOffDemandDetail.TaxAmount)
		roundOff = roundOff.Sub(prev)
	}

	if !roundOff.IsZero() {
		roundOffFloat, _ := roundOff.Float64()
		roundOffDemandDetail := domain.DemandDetail{
			TaxAmount:         roundOffFloat,
			TaxHeadMasterCode: roundOffTaxHeadCode,
			TenantId:          tenantId,
			CollectionAmount:  0.0,
		}
		*demandDetails = append(*demandDetails, roundOffDemandDetail)
	}
}
