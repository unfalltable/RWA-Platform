package config

import (
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

	// 区块链RPC配置
	EthereumRPC string `mapstructure:"ETHEREUM_RPC_URL"`
	ArbitrumRPC string `mapstructure:"ARBITRUM_RPC_URL"`
	BaseRPC     string `mapstructure:"BASE_RPC_URL"`
	SolanaRPC   string `mapstructure:"SOLANA_RPC_URL"`
	BSCRPC      string `mapstructure:"BSC_RPC_URL"`
	PolygonRPC  string `mapstructure:"POLYGON_RPC_URL"`

	// 外部API配置
	CoinGeckoAPIKey     string `mapstructure:"COINGECKO_API_KEY"`
	CoinMarketCapAPIKey string `mapstructure:"COINMARKETCAP_API_KEY"`
	MessariAPIKey       string `mapstructure:"MESSARI_API_KEY"`
	DuneAPIKey          string `mapstructure:"DUNE_API_KEY"`
	NewsAPIKey          string `mapstructure:"NEWS_API_KEY"`

	// 数据采集配置
	PriceCollectionInterval      int `mapstructure:"PRICE_COLLECTION_INTERVAL"`      // 秒
	BlockchainSyncInterval       int `mapstructure:"BLOCKCHAIN_SYNC_INTERVAL"`       // 秒
	NewsCollectionInterval       int `mapstructure:"NEWS_COLLECTION_INTERVAL"`       // 秒
	MaxConcurrentRequests        int `mapstructure:"MAX_CONCURRENT_REQUESTS"`
	RequestTimeout               int `mapstructure:"REQUEST_TIMEOUT"`                // 秒
	RetryAttempts                int `mapstructure:"RETRY_ATTEMPTS"`
	RetryDelay                   int `mapstructure:"RETRY_DELAY"`                    // 秒

	// 缓存配置
	CacheTTL           int `mapstructure:"CACHE_TTL"`            // 秒
	PriceCacheTTL      int `mapstructure:"PRICE_CACHE_TTL"`      // 秒
	BlockchainCacheTTL int `mapstructure:"BLOCKCHAIN_CACHE_TTL"` // 秒

	// 监控配置
	MetricsEnabled bool   `mapstructure:"METRICS_ENABLED"`
	MetricsPort    int    `mapstructure:"METRICS_PORT"`
	TracingEnabled bool   `mapstructure:"TRACING_ENABLED"`
	TracingEndpoint string `mapstructure:"TRACING_ENDPOINT"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/data-collector")

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

	// 处理Kafka brokers
	if brokers := viper.GetString("KAFKA_BROKERS"); brokers != "" {
		viper.Set("KAFKA_BROKERS", strings.Split(brokers, ","))
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func setDefaults() {
	// 服务默认配置
	viper.SetDefault("PORT", 8080)
	viper.SetDefault("LOG_LEVEL", "info")

	// 数据库默认配置
	viper.SetDefault("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/rwa_platform")
	viper.SetDefault("REDIS_URL", "redis://localhost:6379")

	// Kafka默认配置
	viper.SetDefault("KAFKA_BROKERS", []string{"localhost:9092"})

	// 区块链RPC默认配置
	viper.SetDefault("ETHEREUM_RPC", "https://eth-mainnet.alchemyapi.io/v2/demo")
	viper.SetDefault("ARBITRUM_RPC", "https://arb-mainnet.g.alchemy.com/v2/demo")
	viper.SetDefault("BASE_RPC", "https://mainnet.base.org")
	viper.SetDefault("SOLANA_RPC", "https://api.mainnet-beta.solana.com")
	viper.SetDefault("BSC_RPC", "https://bsc-dataseed.binance.org")
	viper.SetDefault("POLYGON_RPC", "https://polygon-mainnet.g.alchemy.com/v2/demo")

	// 数据采集默认配置
	viper.SetDefault("PRICE_COLLECTION_INTERVAL", 60)      // 1分钟
	viper.SetDefault("BLOCKCHAIN_SYNC_INTERVAL", 300)      // 5分钟
	viper.SetDefault("NEWS_COLLECTION_INTERVAL", 1800)     // 30分钟
	viper.SetDefault("MAX_CONCURRENT_REQUESTS", 10)
	viper.SetDefault("REQUEST_TIMEOUT", 30)
	viper.SetDefault("RETRY_ATTEMPTS", 3)
	viper.SetDefault("RETRY_DELAY", 5)

	// 缓存默认配置
	viper.SetDefault("CACHE_TTL", 3600)           // 1小时
	viper.SetDefault("PRICE_CACHE_TTL", 300)      // 5分钟
	viper.SetDefault("BLOCKCHAIN_CACHE_TTL", 600) // 10分钟

	// 监控默认配置
	viper.SetDefault("METRICS_ENABLED", true)
	viper.SetDefault("METRICS_PORT", 9090)
	viper.SetDefault("TRACING_ENABLED", false)
}
