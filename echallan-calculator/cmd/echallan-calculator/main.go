package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/CDPI-HRSS/calci_sp/configs"
	rephttp "github.com/CDPI-HRSS/calci_sp/internal/repository/http"
	"github.com/CDPI-HRSS/calci_sp/internal/service"
	controllers "github.com/CDPI-HRSS/calci_sp/internal/transport/http"
	"github.com/CDPI-HRSS/calci_sp/internal/util"
	"github.com/CDPI-HRSS/calci_sp/internal/validator"
)

func main() {
	// 1. Load Configuration
	cfg := configs.LoadConfig()

	// 2. Initialize Repositories
	srRepo := rephttp.NewServiceRequestRepository()
	demandRepo := rephttp.NewDemandRepository(cfg, srRepo)

	// 3. Initialize Utils & Validator
	utils := util.NewCalculationUtils(cfg, srRepo)
	val := validator.NewCalculatorValidator()

	// 4. Initialize Services
	demandService := service.NewDemandService(cfg, utils, srRepo, demandRepo)
	calcService := service.NewCalculationService(cfg, utils, srRepo, demandService, val)

	// 5. Initialize Controllers
	ctrl := controllers.NewChallanCalController(calcService)

	// 6. Setup Gin Router
	r := gin.Default()

	// Setup health check
	r.GET("/echallan-calculator/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// Register business routes
	ctrl.RegisterRoutes(r, cfg.ContextPath)

	// 7. Start Server
	port := cfg.ServerPort
	if port == "" {
		port = "8078"
	}
	
	log.Printf("Starting echallan-calculator on port %s with context path %s", port, cfg.ContextPath)
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Printf("Failed to start server: %v", err)
		os.Exit(1)
	}
}
