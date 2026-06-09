package rowmapper

import (
	"encoding/json"
	"github.com/CDPI-HRSS/echallan-services/internal/domain"
	"github.com/jmoiron/sqlx"
)

type ChallanDBWrapper struct {
	domain.Challan
	AdditionalDetailRaw []byte  `db:"additionaldetail"`
	Description         *string `db:"description"`
	AccountId           *string `db:"accountid"`
	ReferenceId         *string `db:"referenceid"`
	DoorNo              *string `db:"doorno"`
	BuildingName        *string `db:"buildingname"`
	Street              *string `db:"street"`
	City                *string `db:"city"`
	Pincode             *string `db:"pincode"`
	Latitude            *float64 `db:"latitude"`
	Longitude           *float64 `db:"longitude"`
}

func MapChallanRow(rows *sqlx.Rows) (*domain.Challan, error) {
	var wrapper ChallanDBWrapper
	err := rows.StructScan(&wrapper)
	if err != nil {
		return nil, err
	}

	challan := wrapper.Challan
	if wrapper.Description != nil {
		challan.Description = *wrapper.Description
	}
	if wrapper.AccountId != nil {
		challan.AccountId = *wrapper.AccountId
	}
	if wrapper.ReferenceId != nil {
		challan.ReferenceId = *wrapper.ReferenceId
	}
	challan.Address = &domain.Address{
		TenantId: challan.TenantId,
	}
	if wrapper.DoorNo != nil {
		challan.Address.DoorNo = *wrapper.DoorNo
	}
	if wrapper.BuildingName != nil {
		challan.Address.BuildingName = *wrapper.BuildingName
	}
	if wrapper.Street != nil {
		challan.Address.Street = *wrapper.Street
	}
	if wrapper.City != nil {
		challan.Address.City = *wrapper.City
	}
	if wrapper.Pincode != nil {
		challan.Address.Pincode = *wrapper.Pincode
	}
	
	if wrapper.Latitude != nil || wrapper.Longitude != nil {
		challan.Address.GeoLocation = &domain.GeoLocation{}
		if wrapper.Latitude != nil {
			challan.Address.GeoLocation.Latitude = *wrapper.Latitude
		}
		if wrapper.Longitude != nil {
			challan.Address.GeoLocation.Longitude = *wrapper.Longitude
		}
	}
	
	if wrapper.AdditionalDetailRaw != nil {
		var addDetail interface{}
		_ = json.Unmarshal(wrapper.AdditionalDetailRaw, &addDetail)
		challan.AdditionalDetail = addDetail
	}

	return &challan, nil
}
