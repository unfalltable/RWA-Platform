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
	"github.com/rwa-platform/channel-service/internal/config"
	"github.com/rwa-platform/channel-service/internal/database"
	"github.com/rwa-platform/channel-service/internal/handlers"
	"github.com/rwa-platform/channel-service/internal/kafka"
	"github.com/rwa-platform/channel-service/internal/redis"
	"github.com/rwa-platform/channel-service/internal/services"
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

	logrus.Info("Starting RWA Channel Service...")

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
	channelService := services.NewChannelService(db, redisClient, kafkaProducer, cfg)
	matchingService := services.NewMatchingService(db, redisClient, kafkaProducer, cfg)
	attributionService := services.NewAttributionService(db, redisClient, kafkaProducer, cfg)

	// 启动后台服务
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动渠道数据同步
	go channelService.StartChannelSync(ctx)
	
	// 启动撮合引擎
	go matchingService.StartMatchingEngine(ctx)
	
	// 启动归因统计
	go attributionService.StartAttributionTracking(ctx)

	// 初始化HTTP服务器
	router := setupRouter(channelService, matchingService, attributionService)
	
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
	channelService *services.ChannelService,
	matchingService *services.MatchingService,
	attributionService *services.AttributionService,
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
		// 渠道相关接口
		channels := v1.Group("/channels")
		{
			channels.GET("/", handlers.GetChannels(channelService))
			channels.GET("/:id", handlers.GetChannel(channelService))
			channels.POST("/", handlers.CreateChannel(channelService))
			channels.PUT("/:id", handlers.UpdateChannel(channelService))
			channels.DELETE("/:id", handlers.DeleteChannel(channelService))
			channels.GET("/:id/assets", handlers.GetChannelAssets(channelService))
			channels.POST("/:id/sync", handlers.SyncChannel(channelService))
		}

		// 撮合相关接口
		matching := v1.Group("/matching")
		{
			matching.POST("/match", handlers.MatchChannels(matchingService))
			matching.GET("/quote", handlers.GetQuote(matchingService))
			matching.POST("/redirect", handlers.CreateRedirect(matchingService))
			matching.GET("/redirect/:id", handlers.GetRedirect(matchingService))
		}

		// 归因相关接口
		attribution := v1.Group("/attribution")
		{
			attribution.POST("/track", handlers.TrackAttribution(attributionService))
			attribution.GET("/stats", handlers.GetAttributionStats(attributionService))
			attribution.GET("/conversions", handlers.GetConversions(attributionService))
		}

		// 管理接口
		admin := v1.Group("/admin")
		{
			admin.GET("/stats", handlers.GetSystemStats(channelService, matchingService, attributionService))
			admin.POST("/sync/all", handlers.SyncAllChannels(channelService))
			admin.GET("/health/detailed", handlers.DetailedHealthCheck(channelService))
		}
	}

	return router
}
