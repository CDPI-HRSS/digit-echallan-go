package rowmapper

import (
	"github.com/CDPI-HRSS/echallan-services/internal/domain"
	"github.com/jmoiron/sqlx"
)

func MapChallanRow(rows *sqlx.Rows) (*domain.Challan, error) {
	var challan domain.Challan
	err := rows.StructScan(&challan)
	if err != nil {
		return nil, err
	}
	return &challan, nil
}
