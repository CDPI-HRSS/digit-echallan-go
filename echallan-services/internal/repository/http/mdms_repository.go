package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	config "github.com/CDPI-HRSS/echallan-services/configs"
	"github.com/CDPI-HRSS/echallan-services/internal/domain"
)

type MdmsRepository struct {
	cfg    *config.Config
	client *http.Client
}

func NewMdmsRepository(cfg *config.Config) *MdmsRepository {
	return &MdmsRepository{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (r *MdmsRepository) FetchMasterData(reqInfo *domain.RequestInfo, tenantId string) (map[string]interface{}, error) {
	url := r.cfg.MDMSHost + r.cfg.MDMSSearchEndpoint

	reqBody := map[string]interface{}{
		"RequestInfo": reqInfo,
		"MdmsCriteria": map[string]interface{}{
			"tenantId": tenantId,
			"moduleDetails": []map[string]interface{}{
				{
					"moduleName": "BillingService",
					"masterDetails": []map[string]interface{}{{"name": "TaxHeadMaster"}},
				},
				{
					"moduleName": "egov-location",
					"masterDetails": []map[string]interface{}{{"name": "TenantBoundary"}},
				},
			},
		},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("MDMS service returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if mdmsRes, ok := result["MdmsRes"].(map[string]interface{}); ok {
		return mdmsRes, nil
	}

	return result, nil
}
