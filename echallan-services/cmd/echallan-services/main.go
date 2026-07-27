package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"go.uber.org/zap"

	config "github.com/CDPI-HRSS/echallan-services/configs"
	"github.com/CDPI-HRSS/echallan-services/internal/repository/postgres"

	httpclient "github.com/CDPI-HRSS/echallan-services/internal/repository/http"
	"github.com/CDPI-HRSS/echallan-services/internal/service"
	"github.com/CDPI-HRSS/echallan-services/internal/middleware"
	httptransport "github.com/CDPI-HRSS/echallan-services/internal/transport/http"
	kafkatransport "github.com/CDPI-HRSS/echallan-services/internal/transport/kafka"
	"github.com/CDPI-HRSS/echallan-services/internal/validator"
)

func main() {
	// 1. Initialize Zap Logger
	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()
	zap.ReplaceGlobals(logger)

	cfg := config.LoadConfig()
	zap.L().Info("Loaded configuration", zap.String("port", cfg.ServerPort))

	// 2. Initialize Kafka Producer
	producer := kafkatransport.NewProducer([]string{cfg.KafkaBrokers})

	// 3. Initialize PostgreSQL with Connection Pooling
	dbUri := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dbUri), &gorm.Config{})
	var sqlDB *sql.DB
	if err != nil {
		zap.L().Warn("Failed to connect to Postgres", zap.Error(err))
	} else {
		// Set Golden Standard Connection Pool Limits
		sqlDB, _ = db.DB()
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(25)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
		zap.L().Info("Connected to PostgreSQL successfully")
	}

	// 4. Dependency Injection (Wiring)
	repo := postgres.NewChallanRepository(db, logger, cfg)
	userRepo := httpclient.NewUserRepository(cfg, logger)
	mdmsRepo := httpclient.NewMdmsRepository(cfg)
	locRepo := httpclient.NewLocationRepository(cfg)
	val := validator.NewChallanValidator(mdmsRepo, locRepo)
	notifSvc := service.NewNotificationService(cfg, producer, mdmsRepo, logger)
	idgenRepo := httpclient.NewIdGenRepository(cfg)
	billRepo := httpclient.NewBillingRepository(cfg, logger)

	challanSvc := service.NewChallanService(cfg, producer, repo, val, userRepo, notifSvc, idgenRepo, billRepo, logger)
	challanCtrl := httptransport.NewChallanController(challanSvc, producer)

	paymentSvc := service.NewPaymentUpdateService(producer, repo)
	fsSvc := service.NewFilestoreUpdateService(producer, repo)

	consumer := kafkatransport.NewConsumer([]string{cfg.KafkaBrokers}, paymentSvc, fsSvc, notifSvc, logger)
	
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	consumer.StartListening(ctx)

	// 5. Router Setup & Middleware
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Use Zap logger
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))

	// CORS Middleware
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// In production, validate against exact domain or regex
			return origin == "http://localhost:3000" || origin == "https://digit.org" || origin == "http://localhost:8080"
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	keycloakCfg := middleware.KeycloakConfig{
		URL:      cfg.KeycloakURL,
		Realm:    cfg.KeycloakRealm,
		ClientID: cfg.KeycloakClientID,
	}
	authMw := middleware.KeycloakAuthMiddleware(keycloakCfg)
	challanCtrl.RegisterRoutes(r, authMw)

	r.GET("/echallan-services/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP"})
	})

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	// 6. Graceful Shutdown
	go func() {
		zap.L().Info("Starting eChallan Service", zap.String("port", cfg.ServerPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	<-ctx.Done()
	zap.L().Info("Graceful Shutdown initialized...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		zap.L().Fatal("Server forced to shutdown", zap.Error(err))
	}
	
	// Clean up resources
	if sqlDB != nil {
		sqlDB.Close()
	}
	zap.L().Info("Server exiting gracefully")
}
