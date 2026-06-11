package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/CDPI-HRSS/echallan-services/internal/domain"
)

type mockChallanService struct{}

func (m *mockChallanService) Create(req *domain.ChallanRequest) (*domain.Challan, error) {
	return req.Challan, nil
}
func (m *mockChallanService) Search(criteria domain.SearchCriteria, reqInfo *domain.RequestInfo) ([]*domain.Challan, int, error) {
	return []*domain.Challan{}, 0, nil
}
func (m *mockChallanService) Update(req *domain.ChallanRequest) (*domain.Challan, error) {
	return req.Challan, nil
}
func (m *mockChallanService) Count(tenantId string, reqInfo *domain.RequestInfo) (map[string]interface{}, error) {
	return map[string]interface{}{"TOTAL": 0}, nil
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	ctrl := NewChallanController(&mockChallanService{}, nil)
	ctrl.RegisterRoutes(r)
	return r
}

func TestChallanController_Create_InvalidJSON(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/eChallan/v1/_create", bytes.NewBuffer([]byte("{invalid}")))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestChallanController_Search_Success(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	// Pass valid json for RequestInfoWrapper
	req, _ := http.NewRequest("POST", "/eChallan/v1/_search?tenantId=pb", bytes.NewBuffer([]byte(`{"RequestInfo": {}}`)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
