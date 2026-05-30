package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/CDPI-HRSS/echallan-services/configs"
	"github.com/CDPI-HRSS/echallan-services/internal/repository/postgres"
	"github.com/CDPI-HRSS/echallan-services/internal/service"
	httptransport "github.com/CDPI-HRSS/echallan-services/internal/transport/http"
	kafkatransport "github.com/CDPI-HRSS/echallan-services/internal/transport/kafka"
	"github.com/CDPI-HRSS/echallan-services/internal/validator"
)

func main() {
	cfg := config.LoadConfig()

	// 1. Initialize Kafka Producer
	producer := kafkatransport.NewProducer([]string{cfg.KafkaBrokers})

	// 2. Initialize PostgreSQL
	dbUri := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	
	db, err := sqlx.Connect("postgres", dbUri)
	if err != nil {
		log.Printf("Warning: Failed to connect to Postgres: %v", err)
	}

	// 3. Dependency Injection (Wiring)
	repo := postgres.NewChallanRepository(db)
	val := validator.NewChallanValidator()
	challanSvc := service.NewChallanService(producer, repo, val)
	challanCtrl := httptransport.NewChallanController(challanSvc)
	
	paymentSvc := service.NewPaymentUpdateService(producer)
	consumer := kafkatransport.NewConsumer([]string{cfg.KafkaBrokers}, paymentSvc)
	consumer.StartListening()

	// 4. Router Setup
	r := gin.Default()
	challanCtrl.RegisterRoutes(r)

	// Health Check
	r.GET("/echallan-services/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP"})
	})

	log.Printf("Starting eChallan Service on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
