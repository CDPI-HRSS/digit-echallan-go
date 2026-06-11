package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	config "github.com/CDPI-HRSS/echallan-services/configs"
	"github.com/CDPI-HRSS/echallan-services/internal/domain"
)

type IdGenRepository struct {
	cfg    *config.Config
	client *http.Client
}

func NewIdGenRepository(cfg *config.Config) *IdGenRepository {
	return &IdGenRepository{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (r *IdGenRepository) GenerateId(reqInfo *domain.RequestInfo, tenantId string, idName string, format string, count int) ([]string, error) {
	url := r.cfg.IdGenServiceHost + r.cfg.IdGenServicePath

	idRequests := []domain.IdRequest{
		{
			IdName:   idName,
			TenantId: tenantId,
			Format:   format,
		},
	}

	reqBody := domain.IdGenerationRequest{
		RequestInfo: reqInfo,
		IdRequests:  idRequests,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("IdGen service returned %d", resp.StatusCode)
	}

	var result domain.IdResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var generatedIds []string
	for _, idRes := range result.IdResponses {
		generatedIds = append(generatedIds, idRes.Id)
	}

	if len(generatedIds) == 0 {
		return nil, fmt.Errorf("IdGen service returned empty ID list")
	}

	return generatedIds, nil
}
