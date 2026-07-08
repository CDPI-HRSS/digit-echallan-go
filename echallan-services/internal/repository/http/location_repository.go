package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	config "github.com/CDPI-HRSS/echallan-services/configs"
	"github.com/CDPI-HRSS/echallan-services/internal/domain"
)

type LocationRepository struct {
	cfg    *config.Config
	client *http.Client
}

func NewLocationRepository(cfg *config.Config) *LocationRepository {
	return &LocationRepository{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (r *LocationRepository) GetLocalityCodes(tenantId string, requestInfo *domain.RequestInfo) ([]string, error) {
	// egov-location boundary endpoint
	url := fmt.Sprintf("%s/egov-location/location/v11/boundarys/_search?tenantId=%s&hierarchyTypeCode=REVENUE&boundaryType=Locality", r.cfg.MDMSHost, tenantId)

	req, err := http.NewRequest("POST", url, nil)
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
		return nil, fmt.Errorf("location service returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		TenantBoundary []struct {
			Boundary []struct {
				Code string `json:"code"`
			} `json:"boundary"`
		} `json:"TenantBoundary"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var codes []string
	if len(result.TenantBoundary) > 0 {
		for _, b := range result.TenantBoundary[0].Boundary {
			codes = append(codes, b.Code)
		}
	}
	return codes, nil
}
