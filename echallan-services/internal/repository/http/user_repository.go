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

type UserRepository struct {
	cfg    *config.Config
	client *http.Client
	logger *zap.Logger
}

func NewUserRepository(cfg *config.Config, logger *zap.Logger) *UserRepository {
	return &UserRepository{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
		logger: logger,
	}
}

func (r *UserRepository) SearchUsers(ctx context.Context, requestInfo *domain.RequestInfo, uuids []string) ([]domain.UserInfo, error) {
	if len(uuids) == 0 {
		return nil, nil
	}

	url := r.cfg.UserServiceHost + r.cfg.UserServiceSearchEndpoint

	reqBody := map[string]interface{}{
		"RequestInfo": requestInfo,
		"uuid":        uuids,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		r.logger.Error("failed to create http request", zap.Error(err))
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
		return nil, fmt.Errorf("user service returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		User []domain.UserInfo `json:"user"`
	}
	respBody, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(respBody, &result); err != nil {
		r.logger.Error("USER REPO JSON ERROR", zap.Error(err), zap.String("body", string(respBody)))
		return nil, err
	}

	return result.User, nil
}


func (r *UserRepository) CreateUser(ctx context.Context, requestInfo *domain.RequestInfo, user *domain.UserInfo) (*domain.UserInfo, error) {
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
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		r.logger.Error("failed to create http request", zap.Error(err))
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("user service create returned %d: %s", resp.StatusCode, string(respBody))
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
