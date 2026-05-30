package service

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/CDPI-HRSS/calci_sp/configs"
	"github.com/CDPI-HRSS/calci_sp/internal/domain"
	"github.com/CDPI-HRSS/calci_sp/internal/repository"
	"github.com/CDPI-HRSS/calci_sp/internal/util"
)

const MDMSRoundOffTaxHead = "_ROUNDOFF"

type DemandService struct {
	cfg        *configs.Config
	utils      *util.CalculationUtils
	srRepo     *repository.ServiceRequestRepository
	demandRepo *repository.DemandRepository
}

func NewDemandService(cfg *configs.Config, utils *util.CalculationUtils, srRepo *repository.ServiceRequestRepository, demandRepo *repository.DemandRepository) *DemandService {
	return &DemandService{
		cfg:        cfg,
		utils:      utils,
		srRepo:     srRepo,
		demandRepo: demandRepo,
	}
}

func (s *DemandService) GenerateDemand(requestInfo *domain.RequestInfo, calculations []domain.Calculation, businessService string) error {
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

	demands, err := s.SearchDemand(tenantId, consumerCodes, requestInfo, businessService)
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
		_, err := s.createDemand(requestInfo, createCalculations)
		if err != nil {
			return fmt.Errorf("failed to create demands: %w", err)
		}
	}

	if len(updateCalculations) > 0 {
		_, err := s.updateDemand(requestInfo, updateCalculations, businessService)
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
			_, err := s.GenerateBill(requestInfo, billCriteria)
			if err != nil {
				log.Printf("Warning: failed to generate bill for %s (continuing): %v\n", calc.Challan.ChallanNo, err)
			}
		}
	}

	return nil
}

func (s *DemandService) createDemand(requestInfo *domain.RequestInfo, calculations []domain.Calculation) ([]domain.Demand, error) {
	var demands []domain.Demand
	for _, calc := range calculations {
		challan := calc.Challan
		if challan == nil && calc.ChallanNo != "" {
			var err error
			challan, err = s.utils.GetChallan(requestInfo, calc.ChallanNo, calc.TenantId)
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

	return s.demandRepo.SaveDemand(requestInfo, demands)
}

func (s *DemandService) updateDemand(requestInfo *domain.RequestInfo, calculations []domain.Calculation, businessService string) ([]domain.Demand, error) {
	var demands []domain.Demand
	for _, calc := range calculations {
		challan := calc.Challan
		if challan == nil {
			return nil, fmt.Errorf("challan is nil in update calculation")
		}

		searchResult, err := s.SearchDemand(calc.TenantId, []string{challan.ChallanNo}, requestInfo, businessService)
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

	return s.demandRepo.UpdateDemand(requestInfo, demands)
}

func (s *DemandService) SearchDemand(tenantId string, consumerCodes []string, requestInfo *domain.RequestInfo, businessService string) ([]domain.Demand, error) {
	if len(consumerCodes) == 0 {
		return nil, nil
	}

	codesStr := strings.Join(consumerCodes, ",")
	url := s.utils.GetDemandSearchURL(tenantId, businessService, codesStr)

	wrapper := domain.RequestInfoWrapper{
		RequestInfo: requestInfo,
	}

	var response domain.DemandResponse
	err := s.srRepo.FetchResult(url, wrapper, &response)
	if err != nil {
		return nil, err
	}

	return response.Demands, nil
}

func (s *DemandService) GenerateBill(requestInfo *domain.RequestInfo, billCriteria domain.GenerateBillCriteria) (*domain.BillResponse, error) {
	demands, err := s.SearchDemand(billCriteria.TenantId, []string{billCriteria.ConsumerCode}, requestInfo, billCriteria.BusinessService)
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
	err = s.srRepo.FetchResult(url, wrapper, &response)
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
			var total float64
			for _, d := range details {
				total += d.TaxAmount
			}
			diffInTaxAmount := estimate.EstimateAmount - total
			if diffInTaxAmount != 0 {
				newDemandDetails = append(newDemandDetails, domain.DemandDetail{
					TaxAmount:         diffInTaxAmount,
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
	var totalTax float64
	var prevRoundOffDemandDetail *domain.DemandDetail

	roundOffTaxHeadCode := businessService + MDMSRoundOffTaxHead

	for i := range *demandDetails {
		detail := &(*demandDetails)[i]
		if !strings.EqualFold(detail.TaxHeadMasterCode, roundOffTaxHeadCode) {
			totalTax += detail.TaxAmount
		} else {
			prevRoundOffDemandDetail = detail
		}
	}

	rounded := math.Round(totalTax)
	roundOff := rounded - totalTax

	if prevRoundOffDemandDetail != nil {
		roundOff -= prevRoundOffDemandDetail.TaxAmount
	}

	if roundOff != 0 {
		roundOffDemandDetail := domain.DemandDetail{
			TaxAmount:         roundOff,
			TaxHeadMasterCode: roundOffTaxHeadCode,
			TenantId:          tenantId,
			CollectionAmount:  0.0,
		}
		*demandDetails = append(*demandDetails, roundOffDemandDetail)
	}
}


