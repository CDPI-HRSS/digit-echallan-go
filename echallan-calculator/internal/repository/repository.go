package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/CDPI-HRSS/calci_sp/configs"
	"github.com/CDPI-HRSS/calci_sp/internal/models"
)

type ServiceRequestRepository struct {
	client *http.Client
}

func NewServiceRequestRepository() *ServiceRequestRepository {
	return &ServiceRequestRepository{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (r *ServiceRequestRepository) FetchResult(url string, requestPayload interface{}, responseTarget interface{}) error {
	bodyBytes, err := json.Marshal(requestPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// DEBUG: log every outgoing request
	log.Printf("[DEBUG] --> POST %s | body: %s", url, string(bodyBytes))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// 10MB max response size to prevent memory exhaustion / OOM from downstream failures
	limitReader := io.LimitReader(resp.Body, 10*1024*1024)
	respBody, err := io.ReadAll(limitReader)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// DEBUG: log every response
	log.Printf("[DEBUG] <-- %d from %s | body: %s", resp.StatusCode, url, string(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("external service returned status code %d: %s", resp.StatusCode, string(respBody))
	}

	if err := json.Unmarshal(respBody, responseTarget); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w. Body: %s", err, string(respBody))
	}

	return nil
}

type DemandRepository struct {
	cfg    *configs.Config
	srRepo *ServiceRequestRepository
}

func NewDemandRepository(cfg *configs.Config, srRepo *ServiceRequestRepository) *DemandRepository {
	return &DemandRepository{
		cfg:    cfg,
		srRepo: srRepo,
	}
}

func (dr *DemandRepository) SaveDemand(requestInfo *models.RequestInfo, demands []models.Demand) ([]models.Demand, error) {
	url := dr.cfg.BillingHost + dr.cfg.DemandCreateEndpoint
	reqPayload := models.DemandRequest{
		RequestInfo: requestInfo,
		Demands:     demands,
	}

	var respPayload models.DemandResponse
	err := dr.srRepo.FetchResult(url, reqPayload, &respPayload)
	if err != nil {
		return nil, err
	}
	return respPayload.Demands, nil
}

func (dr *DemandRepository) UpdateDemand(requestInfo *models.RequestInfo, demands []models.Demand) ([]models.Demand, error) {
	url := dr.cfg.BillingHost + dr.cfg.DemandUpdateEndpoint
	reqPayload := models.DemandRequest{
		RequestInfo: requestInfo,
		Demands:     demands,
	}

	var respPayload models.DemandResponse
	err := dr.srRepo.FetchResult(url, reqPayload, &respPayload)
	if err != nil {
		return nil, err
	}
	return respPayload.Demands, nil
}

