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
	"github.com/rwa-platform/data-collector/internal/config"
	"github.com/rwa-platform/data-collector/internal/database"
	"github.com/rwa-platform/data-collector/internal/handlers"
	"github.com/rwa-platform/data-collector/internal/kafka"
	"github.com/rwa-platform/data-collector/internal/redis"
	"github.com/rwa-platform/data-collector/internal/services"
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

	logrus.Info("Starting RWA Data Collector Service...")

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

	// 初始化服务
	priceService := services.NewPriceService(db, redisClient, kafkaProducer, cfg)
	blockchainService := services.NewBlockchainService(db, redisClient, kafkaProducer, cfg)
	newsService := services.NewNewsService(db, redisClient, kafkaProducer, cfg)

	// 启动后台服务
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动价格数据采集
	go priceService.StartPriceCollection(ctx)
	
	// 启动区块链数据采集
	go blockchainService.StartBlockchainIndexing(ctx)
	
	// 启动新闻数据采集
	go newsService.StartNewsCollection(ctx)

	// 初始化HTTP服务器
	router := setupRouter(priceService, blockchainService, newsService)
	
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

func setupRouter(priceService *services.PriceService, blockchainService *services.BlockchainService, newsService *services.NewsService) *gin.Engine {
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
		// 价格相关接口
		prices := v1.Group("/prices")
		{
			prices.GET("/:symbol", handlers.GetPrice(priceService))
			prices.GET("/:symbol/history", handlers.GetPriceHistory(priceService))
		}

		// 区块链相关接口
		blockchain := v1.Group("/blockchain")
		{
			blockchain.GET("/assets/:address", handlers.GetAssetInfo(blockchainService))
			blockchain.GET("/transactions/:hash", handlers.GetTransaction(blockchainService))
		}

		// 新闻相关接口
		news := v1.Group("/news")
		{
			news.GET("/", handlers.GetNews(newsService))
			news.GET("/:id", handlers.GetNewsDetail(newsService))
		}

		// 管理接口
		admin := v1.Group("/admin")
		{
			admin.POST("/sync/prices", handlers.TriggerPriceSync(priceService))
			admin.POST("/sync/blockchain", handlers.TriggerBlockchainSync(blockchainService))
			admin.GET("/stats", handlers.GetStats(priceService, blockchainService, newsService))
		}
	}

	return router
}
