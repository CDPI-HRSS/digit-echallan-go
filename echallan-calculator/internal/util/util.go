package util

import (
	"fmt"
	"net/url"
	"time"

	"github.com/CDPI-HRSS/calci_sp/internal/config"
	"github.com/CDPI-HRSS/calci_sp/internal/models"
	"github.com/CDPI-HRSS/calci_sp/internal/repository"
)

type CalculationUtils struct {
	cfg    *config.Config
	srRepo *repository.ServiceRequestRepository
}

func NewCalculationUtils(cfg *config.Config, srRepo *repository.ServiceRequestRepository) *CalculationUtils {
	return &CalculationUtils{
		cfg:    cfg,
		srRepo: srRepo,
	}
}

func (u *CalculationUtils) GetChallanSearchURL(tenantId, challanNo string) string {
	params := url.Values{}
	params.Set("tenantId", tenantId)
	params.Set("applicationNumber", challanNo)
	return fmt.Sprintf("%s%s%s?%s",
		u.cfg.ChallanHost,
		u.cfg.ChallanContextPath,
		u.cfg.ChallanSearchEndpoint,
		params.Encode(),
	)
}

func (u *CalculationUtils) GetDemandSearchURL(tenantId, businessService, consumerCodes string) string {
	params := url.Values{}
	params.Set("tenantId", tenantId)
	params.Set("businessService", businessService)
	params.Set("consumerCode", consumerCodes)
	return fmt.Sprintf("%s%s?%s",
		u.cfg.BillingHost,
		u.cfg.DemandSearchEndpoint,
		params.Encode(),
	)
}

func (u *CalculationUtils) GetBillGenerateURI(tenantId, consumerCode, businessService string) string {
	params := url.Values{}
	params.Set("tenantId", tenantId)
	params.Set("consumerCode", consumerCode)
	params.Set("businessService", businessService)
	return fmt.Sprintf("%s%s?%s",
		u.cfg.BillingHost,
		u.cfg.FetchBillEndpoint,
		params.Encode(),
	)
}

func (u *CalculationUtils) GetAuditDetails(by string, isCreate bool) *models.AuditDetails {
	nowMs := time.Now().UnixNano() / int64(time.Millisecond)
	if isCreate {
		return &models.AuditDetails{
			CreatedBy:        by,
			LastModifiedBy:   by,
			CreatedTime:      nowMs,
			LastModifiedTime: nowMs,
		}
	}
	return &models.AuditDetails{
		LastModifiedBy:   by,
		LastModifiedTime: nowMs,
	}
}

func (u *CalculationUtils) GetChallan(requestInfo *models.RequestInfo, challanNo, tenantId string) (*models.Challan, error) {
	url := u.GetChallanSearchURL(tenantId, challanNo)

	wrapper := models.RequestInfoWrapper{
		RequestInfo: requestInfo,
	}

	var response models.ChallanResponse
	err := u.srRepo.FetchResult(url, wrapper, &response)
	if err != nil {
		return nil, err
	}

	if len(response.Challans) == 0 {
		return nil, nil
	}

	return &response.Challans[0], nil
}

