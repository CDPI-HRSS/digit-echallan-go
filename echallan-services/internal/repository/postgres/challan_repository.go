package postgres

import (
	"fmt"

	"github.com/CDPI-HRSS/echallan-services/internal/domain"
	"github.com/CDPI-HRSS/echallan-services/internal/repository/query"
	"github.com/CDPI-HRSS/echallan-services/internal/repository/rowmapper"
	"github.com/jmoiron/sqlx"
)

type ChallanRepository interface {
	Search(criteria domain.SearchCriteria) ([]*domain.Challan, int, error)
	Count(tenantId string) (map[string]int, error)
}

type challanRepositoryImpl struct {
	db *sqlx.DB
}

func NewChallanRepository(db *sqlx.DB) ChallanRepository {
	return &challanRepositoryImpl{db: db}
}

func (r *challanRepositoryImpl) Search(criteria domain.SearchCriteria) ([]*domain.Challan, int, error) {
	q, args := query.BuildSearchQuery(criteria)
	
	q, args, err := sqlx.In(q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build IN query: %w", err)
	}

	q = r.db.Rebind(q)

	countQuery := "SELECT COUNT(*) FROM (" + q + ") as sub"
	var totalCount int
	if err := r.db.Get(&totalCount, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to execute absolute count query: %w", err)
	}


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
		return nil, 0, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var challans []*domain.Challan
	for rows.Next() {
		challan, err := rowmapper.MapChallanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		challans = append(challans, challan)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if len(challans) > 0 {
		var ids []string
		for _, c := range challans {
			ids = append(ids, c.Id)
		}
		
		q, args, err := sqlx.In("SELECT echallanid, taxheadcode, amount FROM eg_challan_amount WHERE echallanid IN (?)", ids)
		if err == nil {
			q = r.db.Rebind(q)
			
			type amountRow struct {
				ChallanId   string  `db:"echallanid"`
				TaxHeadCode string  `db:"taxheadcode"`
				Amount      float64 `db:"amount"`
			}
			var amounts []amountRow
			if err := r.db.Select(&amounts, q, args...); err == nil {
				for i, c := range challans {
					for _, a := range amounts {
						if a.ChallanId == c.Id {
							challans[i].Amount = append(challans[i].Amount, domain.Amount{
								TaxHeadCode: a.TaxHeadCode,
								Amount:      a.Amount,
							})
						}
					}
				}
			}
		}
	}

	return challans, totalCount, nil
}

func (r *challanRepositoryImpl) Count(tenantId string) (map[string]int, error) {
	var results []struct {
		Status string `db:"applicationstatus"`
		Count  int    `db:"count"`
	}
	
	err := r.db.Select(&results, query.ChallanCountQuery, tenantId)
	if err != nil {
		return nil, fmt.Errorf("failed to execute count query: %w", err)
	}

	counts := make(map[string]int)
	for _, res := range results {
		counts[res.Status] = res.Count
	}
	
	return counts, nil
}
