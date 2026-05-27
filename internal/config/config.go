package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort            string
	ContextPath           string
	UserHost              string
	UserSearchEndpoint    string
	IdGenHost             string
	IdGenPath             string
	LocationHost          string
	BillingHost           string
	FetchBillEndpoint     string
	DemandCreateEndpoint  string
	DemandUpdateEndpoint  string
	DemandSearchEndpoint  string
	CancelBillEndpoint    string
	ChallanHost           string
	ChallanSearchEndpoint string
	ChallanContextPath    string
}

func LoadConfig() *Config {
	// Attempt to load .env file if present
	_ = godotenv.Load()

	return &Config{
		ServerPort:            getEnv("SERVER_PORT", "8078"),
		ContextPath:           getEnv("SERVER_CONTEXT_PATH", "/echallan-calculator"),
		UserHost:              getEnv("EGOV_USER_HOST", "http://localhost:8085/"),
		UserSearchEndpoint:    getEnv("EGOV_USER_SEARCH_PATH", "/user/_search"),
		IdGenHost:             getEnv("EGOV_IDGEN_HOST", "http://localhost:8088"),
		IdGenPath:             getEnv("EGOV_IDGEN_PATH", "egov-idgen/id/_generate"),
		LocationHost:          getEnv("EGOV_LOCATION_HOST", "https://13.71.65.215.nip.io/"),
		BillingHost:           getEnv("EGOV_BILLINGSERVICE_HOST", "http://localhost:8081"),
		FetchBillEndpoint:     getEnv("EGOV_BILL_GEN_ENDPOINT", "/billing-service/bill/v2/_fetchbill"),
		DemandCreateEndpoint:  getEnv("EGOV_DEMAND_CREATE_ENDPOINT", "/billing-service/demand/_create"),
		DemandUpdateEndpoint:  getEnv("EGOV_DEMAND_UPDATE_ENDPOINT", "/billing-service/demand/_update"),
		DemandSearchEndpoint:  getEnv("EGOV_DEMAND_SEARCH_ENDPOINT", "/billing-service/demand/_search"),
		CancelBillEndpoint:    getEnv("EGOV_CANCEL_BILL_ENDPOINT", "/billing-service/bill/v2/_cancelbill"),
		ChallanHost:           getEnv("EGOV_CHALLAN_HOST", "http://localhost:8079"),
		ChallanSearchEndpoint: getEnv("EGOV_CHALLAN_SEARCH_ENDPOINT", "/_search"),
		ChallanContextPath:    getEnv("EGOV_CHALLAN_CONTEXT_PATH", "/echallan-services/v1"),
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

