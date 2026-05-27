package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/CDPI-HRSS/echallan-services/internal/domain"
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
	query := `SELECT id, tenantid, businessservice, challanno, referenceid, description, accountid, source, taxperiodfrom, taxperiodto, applicationstatus FROM eg_echallan WHERE tenantid = ?`
	args := []interface{}{criteria.TenantId}

	if len(criteria.ChallanNo) > 0 {
		query += ` AND challanno IN (?)`
		args = append(args, criteria.ChallanNo)
	}
	
	query, args, err := sqlx.In(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to build IN query: %w", err)
	}
	query = r.db.Rebind(query)

	if criteria.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", criteria.Offset)
	}
	if criteria.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", criteria.Limit)
	} else {
		query += " LIMIT 50" // Default DIGIT limit
	}

	var challans []*domain.Challan
	err = r.db.Select(&challans, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}

	return challans, nil
}

func (r *challanRepositoryImpl) Count(tenantId string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM eg_echallan WHERE tenantid = $1`
	err := r.db.Get(&count, query, tenantId)
	if err != nil {
		return 0, fmt.Errorf("failed to execute count query: %w", err)
	}
	return count, nil
}
