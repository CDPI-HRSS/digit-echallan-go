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

type UserRepository struct {
	cfg    *config.Config
	client *http.Client
}

func NewUserRepository(cfg *config.Config) *UserRepository {
	return &UserRepository{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (r *UserRepository) SearchUsers(requestInfo *domain.RequestInfo, uuids []string) ([]domain.UserInfo, error) {
	if len(uuids) == 0 {
		return nil, nil
	}

	url := r.cfg.UserServiceHost + r.cfg.UserServiceSearchEndpoint

	reqBody := map[string]interface{}{
		"RequestInfo": requestInfo,
		"uuid":        uuids,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user service returned %d", resp.StatusCode)
	}

	var result struct {
		User []domain.UserInfo `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.User, nil
}


func (r *UserRepository) CreateUser(requestInfo *domain.RequestInfo, user *domain.UserInfo) (*domain.UserInfo, error) {
	if user == nil {
		return nil, nil
	}

	url := r.cfg.UserServiceHost + r.cfg.UserServiceCreateEndpoint

	if user.Type == "" {
		user.Type = "CITIZEN"
	}

	reqBody := map[string]interface{}{
		"requestInfo": requestInfo,
		"user":        user,
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
		return nil, fmt.Errorf("user service create returned %d", resp.StatusCode)
	}

	var result struct {
		User []domain.UserInfo `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.User) > 0 {
		return &result.User[0], nil
	}
	return nil, fmt.Errorf("user service returned empty array")
}
