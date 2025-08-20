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
	"github.com/rwa-platform/portfolio-service/internal/config"
	"github.com/rwa-platform/portfolio-service/internal/database"
	"github.com/rwa-platform/portfolio-service/internal/handlers"
	"github.com/rwa-platform/portfolio-service/internal/kafka"
	"github.com/rwa-platform/portfolio-service/internal/redis"
	"github.com/rwa-platform/portfolio-service/internal/services"
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

	logrus.Info("Starting RWA Portfolio Service...")

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

	kafkaConsumer, err := kafka.NewConsumer(cfg.KafkaBrokers, "portfolio-service-group")
	if err != nil {
		logrus.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer kafkaConsumer.Close()

	// 初始化服务
	portfolioService := services.NewPortfolioService(db, redisClient, kafkaProducer, cfg)
	aggregationService := services.NewAggregationService(db, redisClient, kafkaProducer, cfg)
	analyticsService := services.NewAnalyticsService(db, redisClient, kafkaProducer, cfg)
	reportService := services.NewReportService(db, redisClient, kafkaProducer, cfg)
	syncService := services.NewSyncService(db, redisClient, kafkaProducer, cfg)

	// 启动后台服务
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动持仓同步服务
	go syncService.StartPositionSync(ctx)
	
	// 启动数据聚合服务
	go aggregationService.StartAggregation(ctx)
	
	// 启动分析计算服务
	go analyticsService.StartAnalytics(ctx)
	
	// 启动报表生成服务
	go reportService.StartReportGeneration(ctx)

	// 启动Kafka消费者
	go startKafkaConsumers(ctx, kafkaConsumer, portfolioService, aggregationService, analyticsService)

	// 初始化HTTP服务器
	router := setupRouter(portfolioService, aggregationService, analyticsService, reportService, syncService)
	
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
	portfolioService *services.PortfolioService,
	aggregationService *services.AggregationService,
	analyticsService *services.AnalyticsService,
	reportService *services.ReportService,
	syncService *services.SyncService,
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
		// 投资组合接口
		portfolio := v1.Group("/portfolio")
		{
			portfolio.GET("/:user_id", handlers.GetPortfolio(portfolioService))
			portfolio.GET("/:user_id/summary", handlers.GetPortfolioSummary(portfolioService))
			portfolio.GET("/:user_id/positions", handlers.GetPositions(portfolioService))
			portfolio.GET("/:user_id/performance", handlers.GetPerformance(portfolioService))
			portfolio.GET("/:user_id/allocation", handlers.GetAllocation(portfolioService))
			portfolio.POST("/:user_id/sync", handlers.SyncPortfolio(syncService))
		}

		// 聚合数据接口
		aggregation := v1.Group("/aggregation")
		{
			aggregation.GET("/holdings/:user_id", handlers.GetAggregatedHoldings(aggregationService))
			aggregation.GET("/balances/:user_id", handlers.GetAggregatedBalances(aggregationService))
			aggregation.GET("/transactions/:user_id", handlers.GetAggregatedTransactions(aggregationService))
			aggregation.POST("/refresh/:user_id", handlers.RefreshAggregation(aggregationService))
		}

		// 分析接口
		analytics := v1.Group("/analytics")
		{
			analytics.GET("/returns/:user_id", handlers.GetReturnsAnalysis(analyticsService))
			analytics.GET("/risk/:user_id", handlers.GetRiskAnalysis(analyticsService))
			analytics.GET("/attribution/:user_id", handlers.GetAttributionAnalysis(analyticsService))
			analytics.GET("/benchmark/:user_id", handlers.GetBenchmarkComparison(analyticsService))
			analytics.GET("/correlation/:user_id", handlers.GetCorrelationAnalysis(analyticsService))
		}

		// 报表接口
		reports := v1.Group("/reports")
		{
			reports.GET("/:user_id", handlers.GetReports(reportService))
			reports.POST("/:user_id/generate", handlers.GenerateReport(reportService))
			reports.GET("/:user_id/download/:report_id", handlers.DownloadReport(reportService))
			reports.GET("/:user_id/tax", handlers.GetTaxReport(reportService))
			reports.GET("/:user_id/compliance", handlers.GetComplianceReport(reportService))
		}

		// 同步接口
		sync := v1.Group("/sync")
		{
			sync.POST("/manual/:user_id", handlers.ManualSync(syncService))
			sync.GET("/status/:user_id", handlers.GetSyncStatus(syncService))
			sync.GET("/history/:user_id", handlers.GetSyncHistory(syncService))
			sync.POST("/configure/:user_id", handlers.ConfigureSync(syncService))
		}

		// 管理接口
		admin := v1.Group("/admin")
		{
			admin.GET("/stats", handlers.GetSystemStats(portfolioService, aggregationService, analyticsService))
			admin.POST("/sync/all", handlers.SyncAllPortfolios(syncService))
			admin.GET("/health/detailed", handlers.DetailedHealthCheck(portfolioService))
			admin.GET("/metrics", handlers.GetMetrics(portfolioService))
		}
	}

	return router
}

func startKafkaConsumers(
	ctx context.Context,
	consumer *kafka.Consumer,
	portfolioService *services.PortfolioService,
	aggregationService *services.AggregationService,
	analyticsService *services.AnalyticsService,
) {
	topics := []string{
		"transaction-events",
		"balance-updates",
		"price-updates",
		"user-events",
		"asset-events",
	}

	for _, topic := range topics {
		go func(t string) {
			if err := consumer.Subscribe(t, func(message []byte) error {
				return handleKafkaMessage(t, message, portfolioService, aggregationService, analyticsService)
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
	portfolioService *services.PortfolioService,
	aggregationService *services.AggregationService,
	analyticsService *services.AnalyticsService,
) error {
	switch topic {
	case "transaction-events":
		return portfolioService.HandleTransactionEvent(message)
	case "balance-updates":
		return aggregationService.HandleBalanceUpdate(message)
	case "price-updates":
		return analyticsService.HandlePriceUpdate(message)
	case "user-events":
		return portfolioService.HandleUserEvent(message)
	case "asset-events":
		return analyticsService.HandleAssetEvent(message)
	default:
		logrus.Warnf("Unknown topic: %s", topic)
	}
	return nil
}
