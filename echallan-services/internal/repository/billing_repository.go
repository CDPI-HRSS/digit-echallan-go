package repository

import (
	config "github.com/CDPI-HRSS/echallan-services/configs"
)

type BillingRepository struct {
	cfg *config.Config
}

func NewBillingRepository(cfg *config.Config) *BillingRepository {
	return &BillingRepository{cfg: cfg}
}

// TODO: Implement FetchBill and CreateDemand methods matching the Java equivalent
