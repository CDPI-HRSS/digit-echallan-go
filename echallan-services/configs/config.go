package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	UserServiceHost           string
	UserServiceSearchEndpoint string
	UserServiceCreateEndpoint string
	IdGenServiceHost          string
	IdGenServicePath          string
	BillingServiceHost        string
	BillingServiceCreatePath  string
	CalculatorServiceHost     string
	CalculatorServicePath     string
	ServerPort                string
	DBHost                    string
	DBPort                    string
	DBUser                    string
	DBPassword                string
	DBName                    string
	MDMSHost                  string
	MDMSSearchEndpoint        string
	KafkaBrokers              string
	SaveChallanTopic          string
	SMSTopic                  string
	EmailTopic                string
	UserEventTopic            string
	PdfGenerateTopic          string
}

func LoadConfig() *Config {
	_ = godotenv.Load("configs/.env", ".env")

	return &Config{
		UserServiceHost:           getEnv("USER_SERVICE_HOST", "http://localhost:8080"),
		UserServiceSearchEndpoint: getEnv("USER_SERVICE_SEARCH_ENDPOINT", "/user/_search"),
		UserServiceCreateEndpoint: getEnv("USER_SERVICE_CREATE_ENDPOINT", "/user/users/_createnovalidate"),
		IdGenServiceHost:          getEnv("IDGEN_SERVICE_HOST", "http://localhost:8080"),
		IdGenServicePath:          getEnv("IDGEN_SERVICE_PATH", "/egov-idgen/id/_generate"),
		BillingServiceHost:        getEnv("BILLING_SERVICE_HOST", "http://localhost:8080"),
		BillingServiceCreatePath:  getEnv("BILLING_SERVICE_CREATE_PATH", "/billing-service/bill/v2/_create"),
		CalculatorServiceHost:     getEnv("CALCULATOR_SERVICE_HOST", "http://localhost:8078"),
		CalculatorServicePath:     getEnv("CALCULATOR_SERVICE_PATH", "/echallan-calculator/v1/_calculate"),
		ServerPort:                getEnv("SERVER_PORT", "8082"),
		DBHost:                    getEnv("DB_HOST", "localhost"),
		DBPort:                    getEnv("DB_PORT", "5432"),
		DBUser:                    getEnv("DB_USER", "postgres"),
		DBPassword:                getEnv("DB_PASSWORD", "postgres"),
		DBName:                    getEnv("DB_NAME", "echallan"),
		MDMSHost:                  getEnv("MDMS_HOST", "http://localhost:8080"),
		MDMSSearchEndpoint:        getEnv("MDMS_SEARCH_ENDPOINT", "/egov-mdms-service/v1/_search"),
		KafkaBrokers:              getEnv("KAFKA_BROKERS", "localhost:9092"),
		SaveChallanTopic:          getEnv("PERSISTER_SAVE_CHALLAN_TOPIC", "save-challan"),
		SMSTopic:                  getEnv("KAFKA_TOPIC_SMS", "egov.core.notification.sms"),
		EmailTopic:                getEnv("KAFKA_TOPIC_EMAIL", "egov.core.notification.email"),
		UserEventTopic:            getEnv("KAFKA_TOPIC_USER_EVENT", "persist-user-events-async"),
		PdfGenerateTopic:          getEnv("KAFKA_TOPIC_PDF", "pdf-generated"),
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
