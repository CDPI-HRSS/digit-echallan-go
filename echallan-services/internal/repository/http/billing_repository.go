package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	config "github.com/CDPI-HRSS/echallan-services/configs"
	"github.com/CDPI-HRSS/echallan-services/internal/domain"
	"go.uber.org/zap"
)

type BillingRepository struct {
	cfg    *config.Config
	client *http.Client
	logger *zap.Logger
}

func NewBillingRepository(cfg *config.Config, logger *zap.Logger) *BillingRepository {
	return &BillingRepository{
		cfg:    cfg,
		client: &http.Client{Timeout: 15 * time.Second},
		logger: logger,
	}
}

func (r *BillingRepository) GenerateBill(ctx context.Context, reqInfo *domain.RequestInfo, challan *domain.Challan) error {
	url := fmt.Sprintf("%s%s", r.cfg.CalculatorServiceHost, r.cfg.CalculatorServicePath)
	r.logger.Info("CALLING CALCULATOR", zap.String("url", url))

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
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		r.logger.Error("CALCULATOR RETURNED ERROR", zap.Int("status", resp.StatusCode), zap.String("body", string(respBody)))
		return fmt.Errorf("Calculator service returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
