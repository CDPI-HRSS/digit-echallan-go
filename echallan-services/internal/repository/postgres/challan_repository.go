package postgres

import (
	"fmt"

	"github.com/CDPI-HRSS/echallan-services/internal/domain"
	"github.com/CDPI-HRSS/echallan-services/internal/repository/query"
	"github.com/CDPI-HRSS/echallan-services/internal/repository/rowmapper"
	"github.com/jmoiron/sqlx"
)

type ChallanRepository interface {
	Search(criteria domain.SearchCriteria) ([]*domain.Challan, error)
	Count(tenantId string) (int, error)
}

type challanRepositoryImpl struct {
	db *sqlx.DB
}

func NewChallanRepository(db *sqlx.DB) ChallanRepository {
	return &challanRepositoryImpl{db: db}
}

func (r *challanRepositoryImpl) Search(criteria domain.SearchCriteria) ([]*domain.Challan, error) {
	q := query.ChallanSearchQuery
	args := []interface{}{criteria.TenantId}

	if len(criteria.ChallanNo) > 0 {
		q += ` AND challanno IN (?)`
		args = append(args, criteria.ChallanNo)
	}
	
	q, args, err := sqlx.In(q, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to build IN query: %w", err)
	}
	q = r.db.Rebind(q)

	if criteria.Offset > 0 {
		q += fmt.Sprintf(" OFFSET %d", criteria.Offset)
	}
	if criteria.Limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", criteria.Limit)
	} else {
		q += " LIMIT 50" // Default DIGIT limit
	}

	rows, err := r.db.Queryx(q, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var challans []*domain.Challan
	for rows.Next() {
		challan, err := rowmapper.MapChallanRow(rows)
		if err != nil {
			return nil, err
		}
		challans = append(challans, challan)
	}

	return challans, nil
}

func (r *challanRepositoryImpl) Count(tenantId string) (int, error) {
	var count int
	err := r.db.Get(&count, query.ChallanCountQuery, tenantId)
	if err != nil {
		return 0, fmt.Errorf("failed to execute count query: %w", err)
	}
	return count, nil
}
