package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rwa-platform/risk-engine/internal/config"
	"github.com/rwa-platform/risk-engine/internal/database"
	"github.com/rwa-platform/risk-engine/internal/handlers"
	"github.com/rwa-platform/risk-engine/internal/kafka"
	"github.com/rwa-platform/risk-engine/internal/redis"
	"github.com/rwa-platform/risk-engine/internal/services"
	"github.com/sirupsen/logrus"
)

func main() {
	// 初始化配置
	cfg, err := config.Load()
	if err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	setupLogger(cfg.LogLevel)

	logrus.Info("Starting RWA Risk Engine Service...")

	// 初始化数据库
	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		logrus.Fatalf("Failed to connect to database: %v", err)
	}

	// 初始化Redis
	redisClient, err := redis.NewClient(cfg.RedisURL)
	if err != nil {
		logrus.Fatalf("Failed to connect to Redis: %v", err)
	}

	// 初始化Kafka
	kafkaProducer, err := kafka.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		logrus.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	kafkaConsumer, err := kafka.NewConsumer(cfg.KafkaBrokers, "risk-engine-group")
	if err != nil {
		logrus.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer kafkaConsumer.Close()

	// 初始化服务
	riskService := services.NewRiskService(db, redisClient, kafkaProducer, cfg)
	ratingService := services.NewRatingService(db, redisClient, kafkaProducer, cfg)
	complianceService := services.NewComplianceService(db, redisClient, kafkaProducer, cfg)
	alertService := services.NewAlertService(db, redisClient, kafkaProducer, cfg)

	// 启动后台服务
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动风险监控引擎
	go riskService.StartRiskMonitoring(ctx)
	
	// 启动评分计算引擎
	go ratingService.StartRatingEngine(ctx)
	
	// 启动合规检查引擎
	go complianceService.StartComplianceEngine(ctx)
	
	// 启动预警系统
	go alertService.StartAlertSystem(ctx)

	// 启动Kafka消费者
	go startKafkaConsumers(ctx, kafkaConsumer, riskService, ratingService, complianceService)

	// 初始化HTTP服务器
	router := setupRouter(riskService, ratingService, complianceService, alertService)
	
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	// 启动HTTP服务器
	go func() {
		logrus.Infof("HTTP server starting on port %d", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logrus.Errorf("Server forced to shutdown: %v", err)
	}

	logrus.Info("Server exited")
}

func setupLogger(level string) {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	
	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func setupRouter(
	riskService *services.RiskService,
	ratingService *services.RatingService,
	complianceService *services.ComplianceService,
	alertService *services.AlertService,
) *gin.Engine {
	if gin.Mode() == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 健康检查
	router.GET("/health", handlers.HealthCheck)

	// API路由组
	v1 := router.Group("/api/v1")
	{
		// 风险评估接口
		risk := v1.Group("/risk")
		{
			risk.POST("/assess", handlers.AssessRisk(riskService))
			risk.GET("/profile/:id", handlers.GetRiskProfile(riskService))
			risk.POST("/profile", handlers.CreateRiskProfile(riskService))
			risk.PUT("/profile/:id", handlers.UpdateRiskProfile(riskService))
			risk.GET("/metrics", handlers.GetRiskMetrics(riskService))
		}

		// 评分接口
		rating := v1.Group("/rating")
		{
			rating.POST("/calculate", handlers.CalculateRating(ratingService))
			rating.GET("/asset/:id", handlers.GetAssetRating(ratingService))
			rating.GET("/channel/:id", handlers.GetChannelRating(ratingService))
			rating.GET("/history/:id", handlers.GetRatingHistory(ratingService))
			rating.POST("/update", handlers.UpdateRating(ratingService))
		}

		// 合规检查接口
		compliance := v1.Group("/compliance")
		{
			compliance.POST("/check", handlers.CheckCompliance(complianceService))
			compliance.GET("/rules", handlers.GetComplianceRules(complianceService))
			compliance.POST("/rules", handlers.CreateComplianceRule(complianceService))
			compliance.PUT("/rules/:id", handlers.UpdateComplianceRule(complianceService))
			compliance.GET("/violations", handlers.GetViolations(complianceService))
		}

		// 预警接口
		alerts := v1.Group("/alerts")
		{
			alerts.GET("/", handlers.GetAlerts(alertService))
			alerts.POST("/", handlers.CreateAlert(alertService))
			alerts.PUT("/:id", handlers.UpdateAlert(alertService))
			alerts.DELETE("/:id", handlers.DeleteAlert(alertService))
			alerts.POST("/:id/acknowledge", handlers.AcknowledgeAlert(alertService))
			alerts.GET("/rules", handlers.GetAlertRules(alertService))
			alerts.POST("/rules", handlers.CreateAlertRule(alertService))
		}

		// 管理接口
		admin := v1.Group("/admin")
		{
			admin.GET("/stats", handlers.GetSystemStats(riskService, ratingService, complianceService, alertService))
			admin.POST("/recalculate", handlers.RecalculateRatings(ratingService))
			admin.GET("/health/detailed", handlers.DetailedHealthCheck(riskService))
		}
	}

	return router
}

func startKafkaConsumers(
	ctx context.Context,
	consumer *kafka.Consumer,
	riskService *services.RiskService,
	ratingService *services.RatingService,
	complianceService *services.ComplianceService,
) {
	topics := []string{
		"asset-events",
		"channel-events",
		"user-events",
		"transaction-events",
		"market-events",
	}

	for _, topic := range topics {
		go func(t string) {
			if err := consumer.Subscribe(t, func(message []byte) error {
				return handleKafkaMessage(t, message, riskService, ratingService, complianceService)
			}); err != nil {
				logrus.Errorf("Failed to subscribe to topic %s: %v", t, err)
			}
		}(topic)
	}

	<-ctx.Done()
	logrus.Info("Kafka consumers stopped")
}

func handleKafkaMessage(
	topic string,
	message []byte,
	riskService *services.RiskService,
	ratingService *services.RatingService,
	complianceService *services.ComplianceService,
) error {
	switch topic {
	case "asset-events":
		return ratingService.HandleAssetEvent(message)
	case "channel-events":
		return ratingService.HandleChannelEvent(message)
	case "user-events":
		return riskService.HandleUserEvent(message)
	case "transaction-events":
		return riskService.HandleTransactionEvent(message)
	case "market-events":
		return riskService.HandleMarketEvent(message)
	default:
		logrus.Warnf("Unknown topic: %s", topic)
	}
	return nil
}
