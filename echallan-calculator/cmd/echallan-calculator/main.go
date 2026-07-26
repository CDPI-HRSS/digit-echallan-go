package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	config "github.com/CDPI-HRSS/calci_sp/configs"
	"github.com/CDPI-HRSS/calci_sp/internal/middleware"
	httprepo "github.com/CDPI-HRSS/calci_sp/internal/repository/http"
	"github.com/CDPI-HRSS/calci_sp/internal/service"
	httptransport "github.com/CDPI-HRSS/calci_sp/internal/transport/http"
	"github.com/CDPI-HRSS/calci_sp/internal/util"
	"github.com/CDPI-HRSS/calci_sp/internal/validator"
)

func main() {
	// 1. Initialize Zap Logger
	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()
	zap.ReplaceGlobals(logger)

	cfg := config.LoadConfig()
	zap.L().Info("Loaded calculator configuration", zap.String("port", cfg.ServerPort))

	// 2. Dependency Injection (Wiring)
	srRepo := httprepo.NewServiceRequestRepository()
	utils := util.NewCalculationUtils(cfg, srRepo)
	demandRepo := httprepo.NewDemandRepository(cfg, srRepo)
	demandSvc := service.NewDemandService(cfg, utils, srRepo, demandRepo)
	val := validator.NewCalculatorValidator()
	calcSvc := service.NewCalculationService(cfg, utils, srRepo, demandSvc, val)
	calcCtrl := httptransport.NewChallanCalController(calcSvc)

	// 3. Router Setup & Middleware
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 4. Keycloak Auth Middleware
	keycloakCfg := middleware.KeycloakConfig{
		URL:      cfg.KeycloakURL,
		Realm:    cfg.KeycloakRealm,
		ClientID: cfg.KeycloakClientID,
	}
	authMw := middleware.KeycloakAuthMiddleware(keycloakCfg)

	// Health check (unprotected)
	r.GET("/echallan-calculator/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP"})
	})

	// Register business routes (RESTful + legacy)
	calcCtrl.RegisterRoutes(r, cfg.ContextPath, authMw)

	// 5. Graceful Shutdown
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ServerPort),
		Handler: r,
	}

	go func() {
		zap.L().Info("Starting eChallan Calculator", zap.String("port", cfg.ServerPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.L().Info("Graceful Shutdown initialized...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server forced to shutdown", zap.Error(err))
	}
	
	zap.L().Info("Calculator exiting gracefully")
}
