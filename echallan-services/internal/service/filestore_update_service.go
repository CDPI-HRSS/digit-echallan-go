package service

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/CDPI-HRSS/echallan-services/internal/domain"
	"github.com/CDPI-HRSS/echallan-services/internal/repository/postgres"
	"github.com/CDPI-HRSS/echallan-services/internal/transport/kafka"
)

type FilestoreUpdateService struct {
	producer *kafka.Producer
	repo     postgres.ChallanRepository
}

func NewFilestoreUpdateService(producer *kafka.Producer, repo postgres.ChallanRepository) *FilestoreUpdateService {
	return &FilestoreUpdateService{
		producer: producer,
		repo:     repo,
	}
}

func (s *FilestoreUpdateService) ProcessPdfGenerated(payload []byte) error {
	var pdfEvent map[string]interface{}
	if err := json.Unmarshal(payload, &pdfEvent); err != nil {
		return err
	}

	jobName, ok := pdfEvent["jobName"].(string)
	if !ok || jobName != "mcollect-challan" {
		// Ignore PDFs not meant for echallan
		return nil
	}

	jobid, ok := pdfEvent["jobid"].(string)
	if !ok || jobid == "" {
		return fmt.Errorf("missing jobid (challanNo)")
	}

	filestoreid, ok := pdfEvent["filestoreid"].(string)
	if !ok || filestoreid == "" {
		return fmt.Errorf("missing filestoreid")
	}

	tenantId, _ := pdfEvent["tenantId"].(string)

	criteria := domain.SearchCriteria{
		ChallanNo: jobid,
		TenantId:  tenantId,
	}
	challans, _, err := s.repo.Search(criteria)
	if err != nil || len(challans) == 0 {
		return fmt.Errorf("challan not found for pdf link: %v", err)
	}

	challan := challans[0]
	challan.Filestoreid = filestoreid

	updateReq := domain.ChallanRequest{
		RequestInfo: &domain.RequestInfo{},
		Challan:     challan,
	}
	
	if err := s.producer.Push("update-challan", updateReq); err != nil {
		return fmt.Errorf("failed to push filestore update: %w", err)
	}

	log.Printf("Successfully linked filestore %s to challan %s", filestoreid, jobid)
	return nil
}
