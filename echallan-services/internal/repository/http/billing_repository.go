package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"log"
	"io"
	"time"

	config "github.com/CDPI-HRSS/echallan-services/configs"
	"github.com/CDPI-HRSS/echallan-services/internal/domain"
)

type BillingRepository struct {
	cfg    *config.Config
	client *http.Client
}

func NewBillingRepository(cfg *config.Config) *BillingRepository {
	return &BillingRepository{
		cfg:    cfg,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (r *BillingRepository) GenerateBill(reqInfo *domain.RequestInfo, challan *domain.Challan) error {
	url := fmt.Sprintf("%s%s", r.cfg.CalculatorServiceHost, r.cfg.CalculatorServicePath)
	log.Printf("CALLING CALCULATOR AT URL: %s", url)

	reqBody := domain.CalculationReq{
		RequestInfo: reqInfo,
		CalculationCriteria: []domain.CalculationCriteria{
			{
				TenantId:  challan.TenantId,
				ChallanNo: challan.ChallanNo,
				Challan:   challan,
			},
		},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("CALCULATOR RETURNED %d: %s", resp.StatusCode, string(respBody))
		return fmt.Errorf("Calculator service returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
