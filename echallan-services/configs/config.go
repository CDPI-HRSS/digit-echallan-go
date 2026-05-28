package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort       string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	KafkaBrokers     string
	SaveChallanTopic string
}

func LoadConfig() *Config {
	_ = godotenv.Load("configs/.env", ".env")

	return &Config{
		ServerPort:       getEnv("SERVER_PORT", "8082"),
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBUser:           getEnv("DB_USER", "postgres"),
		DBPassword:       getEnv("DB_PASSWORD", "postgres"),
		DBName:           getEnv("DB_NAME", "echallan"),
		KafkaBrokers:     getEnv("KAFKA_BROKERS", "localhost:9092"),
		SaveChallanTopic: getEnv("PERSISTER_SAVE_CHALLAN_TOPIC", "save-challan"),
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
