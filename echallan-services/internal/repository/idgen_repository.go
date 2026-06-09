package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	config "github.com/CDPI-HRSS/echallan-services/configs"
	"github.com/CDPI-HRSS/echallan-services/internal/domain"
)

type IdGenRepository struct {
	cfg *config.Config
}

func NewIdGenRepository(cfg *config.Config) *IdGenRepository {
	return &IdGenRepository{cfg: cfg}
}

func (r *IdGenRepository) GenerateId(reqInfo *domain.RequestInfo, tenantId string, idName string, format string) (string, error) {
	idReq := domain.IdGenerationRequest{
		RequestInfo: reqInfo,
		IdRequests: []domain.IdRequest{
			{
				IdName:   idName,
				TenantId: tenantId,
				Format:   format,
			},
		},
	}

	body, _ := json.Marshal(idReq)
	url := r.cfg.IdGenServiceHost + r.cfg.IdGenServicePath
	res, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil || res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to generate ID")
	}
	defer res.Body.Close()

	var idResp domain.IdResponse
	if err := json.NewDecoder(res.Body).Decode(&idResp); err != nil || len(idResp.IdResponses) == 0 {
		return "", fmt.Errorf("invalid idgen response")
	}
	return idResp.IdResponses[0].Id, nil
}
