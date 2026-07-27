package http

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CDPI-HRSS/echallan-services/internal/domain"
	"github.com/gin-gonic/gin"
)

type mockChallanService struct{}

func (m *mockChallanService) Create(ctx context.Context, req *domain.ChallanRequest) (*domain.Challan, error) {
	return req.Challan, nil
}
func (m *mockChallanService) Search(ctx context.Context, criteria domain.SearchCriteria, reqInfo *domain.RequestInfo) ([]*domain.Challan, int, error) {
	return []*domain.Challan{}, 0, nil
}
func (m *mockChallanService) Update(ctx context.Context, req *domain.ChallanRequest) (*domain.Challan, error) {
	return req.Challan, nil
}
func (m *mockChallanService) Count(ctx context.Context, tenantId string, reqInfo *domain.RequestInfo) (map[string]interface{}, error) {
	return map[string]interface{}{"TOTAL": 0}, nil
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	// Kafka producer is nil for mock testing
	ctrl := NewChallanController(&mockChallanService{}, nil)
	ctrl.RegisterRoutes(r, func(c *gin.Context) { c.Next() })
	return r
}

func TestChallanController_Create(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name       string
		payload    string
		wantStatus int
	}{
		{
			name:       "Invalid JSON",
			payload:    `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Valid JSON",
			payload:    `{"RequestInfo": {}, "Challan": {"tenantId": "pb.amritsar"}}`,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/challans", bytes.NewBufferString(tt.payload))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestChallanController_Search(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name       string
		path       string
		payload    string
		wantStatus int
	}{
		{
			name:       "Invalid JSON",
			path:       "/api/v1/challans/search?tenantId=pb",
			payload:    `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Missing TenantId",
			path:       "/api/v1/challans/search",
			payload:    `{"RequestInfo": {}}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Valid Search",
			path:       "/api/v1/challans/search?tenantId=pb.amritsar",
			payload:    `{"RequestInfo": {}}`,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", tt.path, bytes.NewBufferString(tt.payload))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}
