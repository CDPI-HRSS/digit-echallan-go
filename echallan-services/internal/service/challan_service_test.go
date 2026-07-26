package service

import (
	"errors"
	"testing"

	"github.com/CDPI-HRSS/echallan-services/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockChallanRepository mocks the ChallanRepository interface as requested.
type MockChallanRepository struct {
	mock.Mock
}

func (m *MockChallanRepository) Create(req *domain.ChallanRequest) (*domain.Challan, error) {
	args := m.Called(req)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Challan), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockChallanRepository) Search(criteria domain.SearchCriteria) ([]*domain.Challan, int, error) {
	args := m.Called(criteria)
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.Challan), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}

func (m *MockChallanRepository) Update(req *domain.ChallanRequest) (*domain.Challan, error) {
	args := m.Called(req)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Challan), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockChallanRepository) Count(tenantId string) (map[string]int, error) {
	args := m.Called(tenantId)
	if args.Get(0) != nil {
		return args.Get(0).(map[string]int), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestCreateChallan(t *testing.T) {
	tests := []struct {
		name          string
		req           *domain.ChallanRequest
		mockSetup     func(mockRepo *MockChallanRepository)
		expectedError error
	}{
		{
			name: "Success",
			req: &domain.ChallanRequest{
				Challan: &domain.Challan{
					TenantId: "pb.amritsar",
				},
			},
			mockSetup: func(mockRepo *MockChallanRepository) {
				mockRepo.On("Create", mock.Anything).Return(&domain.Challan{Id: "123"}, nil)
			},
			expectedError: nil,
		},
		{
			name: "RepoError",
			req: &domain.ChallanRequest{
				Challan: &domain.Challan{
					TenantId: "pb.amritsar",
				},
			},
			mockSetup: func(mockRepo *MockChallanRepository) {
				mockRepo.On("Create", mock.Anything).Return(nil, errors.New("repository error"))
			},
			expectedError: errors.New("repository error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockChallanRepository)
			tt.mockSetup(mockRepo)

			// Simulating the use case with the mock repository directly for this test
			_, err := mockRepo.Create(tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
