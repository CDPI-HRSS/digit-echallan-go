package service

import (
	"fmt"
	"time"

	"github.com/CDPI-HRSS/echallan-services/internal/domain"

	"github.com/CDPI-HRSS/echallan-services/configs"
	"github.com/CDPI-HRSS/echallan-services/internal/repository/postgres"
	"github.com/CDPI-HRSS/echallan-services/internal/repository/http"
	"github.com/CDPI-HRSS/echallan-services/internal/transport/kafka"
	"github.com/CDPI-HRSS/echallan-services/internal/validator"
	"github.com/google/uuid"
)

type challanServiceImpl struct {
	cfg       *config.Config
	userRepo  *http.UserRepository
	notifSvc  *NotificationService
	idgenRepo *http.IdGenRepository
	billRepo  *http.BillingRepository
	producer  *kafka.Producer
	repo      postgres.ChallanRepository
	validator *validator.ChallanValidator
}

func NewChallanService(cfg *config.Config, producer *kafka.Producer, repo postgres.ChallanRepository, val *validator.ChallanValidator, userRepo *http.UserRepository, notifSvc *NotificationService, idgenRepo *http.IdGenRepository, billRepo *http.BillingRepository) ChallanService {
	return &challanServiceImpl{
		cfg:       cfg,
		producer:  producer,
		repo:      repo,
		userRepo:  userRepo,
		notifSvc:  notifSvc,
		idgenRepo: idgenRepo,
		billRepo:  billRepo,
		validator: val,
	}
}

func (s *challanServiceImpl) Create(req *domain.ChallanRequest) (*domain.Challan, error) {
	if err := s.validator.ValidateCreateRequest(req); err != nil {
		return nil, err
	}

	// Enrichment
	req.Challan.Id = uuid.New().String()
	if req.Challan.Address != nil {
		req.Challan.Address.Id = uuid.New().String()
	}
	req.Challan.ApplicationStatus = "ACTIVE"
	
	// IDGen Call
	ids, err := s.idgenRepo.GenerateId(req.RequestInfo, req.Challan.TenantId, "echallan.number", "CH-[cy:yyyy-MM-dd]-[SEQ_EG_CHALLAN_NUM]", 1)
	if err != nil || len(ids) == 0 {
		return nil, fmt.Errorf("failed to generate Challan Number from IDGen service: %w", err)
	}
	req.Challan.ChallanNo = ids[0]
	
	// User Creation / UUID assignment
	if req.Challan.Citizen != nil {
		createdUser, err := s.userRepo.CreateUser(req.RequestInfo, req.Challan.Citizen)
		if err != nil || createdUser == nil {
			return nil, fmt.Errorf("failed to create/fetch citizen profile from User Service: %w", err)
		}
		req.Challan.Citizen.Uuid = createdUser.Uuid
		req.Challan.Citizen.Id = createdUser.Id
		req.Challan.AccountId = createdUser.Uuid
	} else if req.Challan.AccountId == "" && req.RequestInfo != nil && req.RequestInfo.UserInfo != nil {
		req.Challan.AccountId = req.RequestInfo.UserInfo.Uuid
	}

	// Audit Details
	var reqUuid string
	if req.RequestInfo != nil && req.RequestInfo.UserInfo != nil {
		reqUuid = req.RequestInfo.UserInfo.Uuid
	}
	req.Challan.AuditDetails = &domain.AuditDetails{
		CreatedBy:        reqUuid,
		LastModifiedBy:   reqUuid,
		CreatedTime:      time.Now().UnixMilli(),
		LastModifiedTime: time.Now().UnixMilli(),
	}

	// Calculation Call
	err = s.billRepo.GenerateBill(req.RequestInfo, req.Challan)
	if err != nil {
		return nil, fmt.Errorf("failed to generate calculation/demand: %w", err)
	}

	err = s.producer.Push(s.cfg.SaveChallanTopic, req)
	if err != nil {
		return nil, fmt.Errorf("failed to persist challan: %w", err)
	}

	go s.notifSvc.SendNotifications(req.RequestInfo, req.Challan, "CREATED")

	return req.Challan, nil
}

func (s *challanServiceImpl) Update(req *domain.ChallanRequest) (*domain.Challan, error) {
	// 1. Validation
	searchCriteria := domain.SearchCriteria{
		TenantId:  req.Challan.TenantId,
		ChallanNo: req.Challan.ChallanNo,
	}
	searchResult, _, err := s.Search(searchCriteria, req.RequestInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to search for existing challan during update: %w", err)
	}

	if err := s.validator.ValidateUpdateRequest(req, searchResult); err != nil {
		return nil, err
	}

	// 2. FileStore Soft-Delete Enrichment
	if req.Challan.ApplicationStatus == "CANCELLED" && len(searchResult) > 0 {
		searchChallan := searchResult[0]
		if searchChallan.Filestoreid != "" {
			req.Challan.Filestoreid = "_INACTIVE_" // Custom flag for filestore consumer to delete or ignore
		}
	}

	// 3. Audit Details Enrichment
	var reqUuid string
	if req.RequestInfo != nil && req.RequestInfo.UserInfo != nil {
		reqUuid = req.RequestInfo.UserInfo.Uuid
	}
	if len(searchResult) > 0 && searchResult[0].AuditDetails != nil {
		req.Challan.AuditDetails = searchResult[0].AuditDetails
	} else {
		req.Challan.AuditDetails = &domain.AuditDetails{}
	}
	req.Challan.AuditDetails.LastModifiedBy = reqUuid
	req.Challan.AuditDetails.LastModifiedTime = time.Now().UnixMilli()

	// 4. Calculation Call
	err = s.billRepo.GenerateBill(req.RequestInfo, req.Challan)
	if err != nil {
		return nil, fmt.Errorf("failed to update calculation/demand: %w", err)
	}

	err = s.producer.Push(s.cfg.UpdateChallanTopic, req)
	if err != nil {
		return nil, fmt.Errorf("failed to persist challan update: %w", err)
	}

	go s.notifSvc.SendNotifications(req.RequestInfo, req.Challan, "UPDATED")

	return req.Challan, nil
}

func (s *challanServiceImpl) Search(criteria domain.SearchCriteria, reqInfo *domain.RequestInfo) ([]*domain.Challan, int, error) {
	// RBAC logic
	if reqInfo != nil && reqInfo.UserInfo != nil && reqInfo.UserInfo.Type != "EMPLOYEE" {
		criteria.AccountId = reqInfo.UserInfo.Uuid
	}

	challans, count, err := s.repo.Search(criteria)
	if err != nil || len(challans) == 0 {
		return challans, count, err
	}

	var uuids []string
	uuidMap := make(map[string]bool)
	for _, c := range challans {
		if c.AccountId != "" && !uuidMap[c.AccountId] {
			uuidMap[c.AccountId] = true
			uuids = append(uuids, c.AccountId)
		}
	}

	fmt.Printf("SEARCH uuids: %v\n", uuids)
	users, err := s.userRepo.SearchUsers(reqInfo, uuids)
	fmt.Printf("SEARCH users returned: %v, err: %v\n", len(users), err)
	userObjMap := make(map[string]*domain.UserInfo)
	for i := range users {
		userObjMap[users[i].Uuid] = &users[i]
	}

	for _, c := range challans {
		if user, ok := userObjMap[c.AccountId]; ok {
			c.Citizen = user
		}
	}

	return challans, count, nil
}

func (s *challanServiceImpl) Count(tenantId string, reqInfo *domain.RequestInfo) (map[string]interface{}, error) {
	// Dynamic Dashboard Data logic
	countMap, err := s.repo.Count(tenantId)
	if err != nil {
		return nil, err
	}
	
	res := make(map[string]interface{})
	for k, v := range countMap {
		res[k] = v
	}
	
	// Simulated Dynamic Data
	res["totalServices"] = countMap["totalChallans"]
	res["totalCollection"] = countMap["totalAmount"]
	
	return res, nil
}
