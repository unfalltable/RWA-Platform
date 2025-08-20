package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	// 服务配置
	Port     int    `mapstructure:"PORT"`
	LogLevel string `mapstructure:"LOG_LEVEL"`
	
	// 数据库配置
	DatabaseURL string `mapstructure:"DATABASE_URL"`
	RedisURL    string `mapstructure:"REDIS_URL"`
	
	// Kafka配置
	KafkaBrokers []string `mapstructure:"KAFKA_BROKERS"`
	
	// 外部API配置
	CoinbaseAPIKey    string `mapstructure:"COINBASE_API_KEY"`
	CoinbaseAPISecret string `mapstructure:"COINBASE_API_SECRET"`
	BinanceAPIKey     string `mapstructure:"BINANCE_API_KEY"`
	BinanceAPISecret  string `mapstructure:"BINANCE_API_SECRET"`
	KrakenAPIKey      string `mapstructure:"KRAKEN_API_KEY"`
	KrakenAPISecret   string `mapstructure:"KRAKEN_API_SECRET"`
	
	// 撮合配置
	MatchingInterval      int     `mapstructure:"MATCHING_INTERVAL"`
	MaxMatchingResults    int     `mapstructure:"MAX_MATCHING_RESULTS"`
	MinMatchingScore      float64 `mapstructure:"MIN_MATCHING_SCORE"`
	RedirectExpiration    int     `mapstructure:"REDIRECT_EXPIRATION"`
	
	// 归因配置
	AttributionWindow     int `mapstructure:"ATTRIBUTION_WINDOW"`
	ConversionTimeout     int `mapstructure:"CONVERSION_TIMEOUT"`
	AttributionCacheTTL   int `mapstructure:"ATTRIBUTION_CACHE_TTL"`
	
	// 渠道同步配置
	ChannelSyncInterval   int `mapstructure:"CHANNEL_SYNC_INTERVAL"`
	ChannelCacheTTL       int `mapstructure:"CHANNEL_CACHE_TTL"`
	MaxConcurrentSyncs    int `mapstructure:"MAX_CONCURRENT_SYNCS"`
	
	// 费用计算配置
	DefaultTradingFee     float64 `mapstructure:"DEFAULT_TRADING_FEE"`
	DefaultWithdrawalFee  float64 `mapstructure:"DEFAULT_WITHDRAWAL_FEE"`
	FeeCalculationMethod  string  `mapstructure:"FEE_CALCULATION_METHOD"`
	
	// 安全配置
	JWTSecret             string `mapstructure:"JWT_SECRET"`
	APIRateLimit          int    `mapstructure:"API_RATE_LIMIT"`
	RequestTimeout        int    `mapstructure:"REQUEST_TIMEOUT"`
	
	// 监控配置
	MetricsEnabled        bool   `mapstructure:"METRICS_ENABLED"`
	MetricsPort          int    `mapstructure:"METRICS_PORT"`
	TracingEnabled       bool   `mapstructure:"TRACING_ENABLED"`
	TracingEndpoint      string `mapstructure:"TRACING_ENDPOINT"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")
	
	// 设置默认值
	setDefaults()
	
	// 自动读取环境变量
	viper.AutomaticEnv()
	
	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}
	
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}
	
	// 处理特殊的环境变量
	if kafkaBrokers := os.Getenv("KAFKA_BROKERS"); kafkaBrokers != "" {
		config.KafkaBrokers = strings.Split(kafkaBrokers, ",")
	}
	
	return &config, nil
}

func setDefaults() {
	// 服务配置
	viper.SetDefault("PORT", 8003)
	viper.SetDefault("LOG_LEVEL", "info")
	
	// 数据库配置
	viper.SetDefault("DATABASE_URL", "postgres://user:password@localhost:5432/rwa_platform?sslmode=disable")
	viper.SetDefault("REDIS_URL", "redis://localhost:6379")
	
	// Kafka配置
	viper.SetDefault("KAFKA_BROKERS", []string{"localhost:9092"})
	
	// 撮合配置
	viper.SetDefault("MATCHING_INTERVAL", 30)
	viper.SetDefault("MAX_MATCHING_RESULTS", 10)
	viper.SetDefault("MIN_MATCHING_SCORE", 0.6)
	viper.SetDefault("REDIRECT_EXPIRATION", 3600)
	
	// 归因配置
	viper.SetDefault("ATTRIBUTION_WINDOW", 86400)
	viper.SetDefault("CONVERSION_TIMEOUT", 1800)
	viper.SetDefault("ATTRIBUTION_CACHE_TTL", 300)
	
	// 渠道同步配置
	viper.SetDefault("CHANNEL_SYNC_INTERVAL", 300)
	viper.SetDefault("CHANNEL_CACHE_TTL", 600)
	viper.SetDefault("MAX_CONCURRENT_SYNCS", 5)
	
	// 费用计算配置
	viper.SetDefault("DEFAULT_TRADING_FEE", 0.001)
	viper.SetDefault("DEFAULT_WITHDRAWAL_FEE", 0.0005)
	viper.SetDefault("FEE_CALCULATION_METHOD", "percentage")
	
	// 安全配置
	viper.SetDefault("JWT_SECRET", "your-secret-key")
	viper.SetDefault("API_RATE_LIMIT", 100)
	viper.SetDefault("REQUEST_TIMEOUT", 30)
	
	// 监控配置
	viper.SetDefault("METRICS_ENABLED", true)
	viper.SetDefault("METRICS_PORT", 9003)
	viper.SetDefault("TRACING_ENABLED", false)
	viper.SetDefault("TRACING_ENDPOINT", "http://localhost:14268/api/traces")
}
