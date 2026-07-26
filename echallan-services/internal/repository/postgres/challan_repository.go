package postgres

import (
	"fmt"

	"github.com/CDPI-HRSS/echallan-services/internal/domain"
	"gorm.io/gorm"
)

type ChallanRepository interface {
	Search(criteria domain.SearchCriteria) ([]*domain.Challan, int, error)
	Count(tenantId string) (map[string]int, error)
}

type challanRepositoryImpl struct {
	db *gorm.DB
}

func NewChallanRepository(db *gorm.DB) ChallanRepository {
	return &challanRepositoryImpl{db: db}
}

func (r *challanRepositoryImpl) Search(criteria domain.SearchCriteria) ([]*domain.Challan, int, error) {
	var dbChallans []domain.ChallanDB
	query := r.db.Preload("Amounts").Preload("AddressDB").Where("tenantid = ?", criteria.TenantId)

	if criteria.Ids != "" {
		query = query.Where("id IN ?", parseCommaSeparated(criteria.Ids))
	}
	if criteria.ChallanNo != "" {
		query = query.Where("challanno IN ?", parseCommaSeparated(criteria.ChallanNo))
	}
	if criteria.AccountId != "" {
		query = query.Where("accountid IN ?", parseCommaSeparated(criteria.AccountId))
	}
	// mobileNumber would need a join or in-memory resolution if user is separate, skipping for direct table query
	if criteria.BusinessService != "" {
		query = query.Where("businessservice IN ?", parseCommaSeparated(criteria.BusinessService))
	}
	if criteria.Status != "" {
		query = query.Where("applicationstatus IN ?", parseCommaSeparated(criteria.Status))
	}
	if criteria.ReceiptNumber != "" {
		query = query.Where("receiptnumber IN ?", parseCommaSeparated(criteria.ReceiptNumber))
	}

	var totalCount int64
	if err := query.Model(&domain.ChallanDB{}).Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to execute absolute count query: %w", err)
	}

	if criteria.Offset > 0 {
		query = query.Offset(criteria.Offset)
	}
	if criteria.Limit > 0 {
		query = query.Limit(criteria.Limit)
	} else {
		query = query.Limit(50) // Default DIGIT limit
	}

	if err := query.Find(&dbChallans).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to execute search query: %w", err)
	}

	var challans []*domain.Challan
	for _, dbC := range dbChallans {
		challans = append(challans, dbC.ToChallan())
	}

	return challans, int(totalCount), nil
}

func (r *challanRepositoryImpl) Count(tenantId string) (map[string]int, error) {
	var results []struct {
		Status string `gorm:"column:applicationstatus"`
		Count  int    `gorm:"column:count"`
	}

	err := r.db.Model(&domain.ChallanDB{}).
		Select("applicationstatus, count(*) as count").
		Where("tenantid = ?", tenantId).
		Group("applicationstatus").
		Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to execute count query: %w", err)
	}

	counts := make(map[string]int)
	for _, res := range results {
		counts[res.Status] = res.Count
	}

	return counts, nil
}

func parseCommaSeparated(s string) []string {
	var res []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			res = append(res, s[start:i])
			start = i + 1
		}
	}
	res = append(res, s[start:])
	return res
}
