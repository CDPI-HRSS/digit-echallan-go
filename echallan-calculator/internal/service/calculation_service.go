package service

import (
	"fmt"
	"strings"

	"github.com/CDPI-HRSS/calci_sp/configs"
	"github.com/CDPI-HRSS/calci_sp/internal/domain"
	"github.com/CDPI-HRSS/calci_sp/internal/repository/http"
	"github.com/CDPI-HRSS/calci_sp/internal/util"
	"github.com/CDPI-HRSS/calci_sp/internal/validator"
)

type CalculationService struct {
	cfg           *configs.Config
	utils         *util.CalculationUtils
	srRepo        *http.ServiceRequestRepository
	demandService *DemandService
	validator     *validator.CalculatorValidator
}

func NewCalculationService(cfg *configs.Config, utils *util.CalculationUtils, srRepo *http.ServiceRequestRepository, demandService *DemandService, val *validator.CalculatorValidator) *CalculationService {
	return &CalculationService{
		cfg:           cfg,
		utils:         utils,
		srRepo:        srRepo,
		demandService: demandService,
		validator:     val,
	}
}

func (s *CalculationService) GetCalculation(req *domain.CalculationReq) ([]domain.Calculation, error) {
	if err := s.validator.ValidateCalculationReq(req); err != nil {
		return nil, err
	}
	calculations, err := s.getCalculationInternal(req.RequestInfo, req.CalculationCriteria)
	if err != nil {
		return nil, err
	}

	for i := range req.CalculationCriteria {
		criteria := &req.CalculationCriteria[i]
		if criteria.Challan != nil && strings.EqualFold(criteria.Challan.ApplicationStatus, "CANCELLED") {
			if err := s.CancelBill(req.RequestInfo, criteria.Challan); err != nil {
				return nil, err
			}
		}
	}

	if len(calculations) > 0 {
		var businessService string
		if calculations[0].Challan != nil {
			businessService = calculations[0].Challan.BusinessService
		} else {
			// Find business service from criteria if challan is not populated yet
			for _, crit := range req.CalculationCriteria {
				if crit.Challan != nil {
					businessService = crit.Challan.BusinessService
					break
				}
			}
		}
		err = s.demandService.GenerateDemand(req.RequestInfo, calculations, businessService)
		if err != nil {
			return nil, err
		}
	}

	return calculations, nil
}

func (s *CalculationService) getCalculationInternal(requestInfo *domain.RequestInfo, criteriaList []domain.CalculationCriteria) ([]domain.Calculation, error) {
	var calculations []domain.Calculation

	for i := range criteriaList {
		criteria := &criteriaList[i]
		challan := criteria.Challan
		if challan == nil && criteria.ChallanNo != "" {
			var err error
			challan, err = s.utils.GetChallan(requestInfo, criteria.ChallanNo, criteria.TenantId)
			if err != nil {
				return nil, err
			}
			criteria.Challan = challan
		}

		if challan == nil {
			return nil, fmt.Errorf("INVALID APPLICATIONNUMBER: Challan does not exist for application number %s", criteria.ChallanNo)
		}

		var estimates []domain.TaxHeadEstimate
		for _, amt := range challan.Amount {
			estimates = append(estimates, domain.TaxHeadEstimate{
				EstimateAmount: amt.Amount,
				TaxHeadCode:    amt.TaxHeadCode,
				
			})
		}

		challanNo := criteria.ChallanNo
		if challanNo == "" {
			challanNo = challan.ChallanNo
		}

		calc := domain.Calculation{
			ChallanNo:        challanNo,
			Challan:          criteria.Challan,
			TenantId:         criteria.TenantId,
			TaxHeadEstimates: estimates,
		}
		calculations = append(calculations, calc)
	}

	return calculations, nil
}

func (s *CalculationService) CancelBill(requestInfo *domain.RequestInfo, challan *domain.Challan) error {
	if challan == nil {
		return nil
	}

	url := s.cfg.BillingHost + s.cfg.CancelBillEndpoint

	updateBillCriteria := map[string]interface{}{
		"tenantId":          challan.TenantId,
		"consumerCodes":     []string{challan.ChallanNo},
		"businessService":   challan.BusinessService,
		"additionalDetails": challan.AdditionalDetail,
	}

	requestBody := map[string]interface{}{
		"RequestInfo":        requestInfo,
		"UpdateBillCriteria": updateBillCriteria,
	}

	var responseTarget interface{}
	err := s.srRepo.FetchResult(url, requestBody, &responseTarget)
	if err != nil {
		return fmt.Errorf("billing service failed to cancel demand: %w", err)
	}
	return nil
}

