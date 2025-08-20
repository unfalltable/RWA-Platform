package services

import (
	"context"
	"testing"
	"time"

	"github.com/rwa-platform/data-collector/internal/config"
	"github.com/rwa-platform/data-collector/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockRedisClient 模拟Redis客户端
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

// MockKafkaProducer 模拟Kafka生产者
type MockKafkaProducer struct {
	mock.Mock
}

func (m *MockKafkaProducer) PublishMessage(topic string, key string, message interface{}) error {
	args := m.Called(topic, key, message)
	return args.Error(0)
}

func (m *MockKafkaProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 自动迁移
	db.AutoMigrate(&models.Asset{}, &models.PriceData{})

	return db
}

func TestPriceService_GetPrice(t *testing.T) {
	db := setupTestDB()
	mockRedis := new(MockRedisClient)
	mockKafka := new(MockKafkaProducer)
	
	cfg := &config.Config{
		PriceCacheTTL: 300,
	}

	service := &PriceService{
		db:     db,
		redis:  mockRedis,
		kafka:  mockKafka,
		config: cfg,
	}

	// 创建测试资产
	asset := &models.Asset{
		ID:     "test-asset-id",
		Symbol: "TEST",
		Name:   "Test Asset",
		Type:   "stablecoin",
	}
	db.Create(asset)

	// 创建测试价格数据
	priceData := &models.PriceData{
		AssetID:   asset.ID,
		Symbol:    asset.Symbol,
		Price:     1.0,
		Currency:  "USD",
		Source:    "test",
		Timestamp: time.Now(),
	}
	db.Create(priceData)

	// 测试从缓存获取（缓存未命中）
	mockRedis.On("Get", mock.Anything, "price:TEST").Return("", assert.AnError)

	result, err := service.GetPrice("TEST")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "TEST", result.Symbol)
	assert.Equal(t, 1.0, result.Price)

	mockRedis.AssertExpectations(t)
}

func TestPriceService_ProcessPriceData(t *testing.T) {
	db := setupTestDB()
	mockRedis := new(MockRedisClient)
	mockKafka := new(MockKafkaProducer)
	
	cfg := &config.Config{
		PriceCacheTTL: 300,
	}

	service := &PriceService{
		db:     db,
		redis:  mockRedis,
		kafka:  mockKafka,
		config: cfg,
	}

	// 创建测试资产
	asset := models.Asset{
		ID:     "test-asset-id",
		Symbol: "TEST",
		Name:   "Test Asset",
		Type:   "stablecoin",
	}
	db.Create(&asset)

	assetMap := map[string]models.Asset{
		"test": asset,
	}

	// 模拟价格数据
	data := map[string]interface{}{
		"usd":            1.0,
		"usd_market_cap": 1000000.0,
		"usd_24h_vol":    50000.0,
		"usd_24h_change": 0.1,
	}

	// 设置mock期望
	mockRedis.On("Set", mock.Anything, "price:TEST", mock.Anything, time.Duration(300)*time.Second).Return(nil)
	mockKafka.On("PublishMessage", "price-updates", "TEST", mock.Anything).Return(nil)

	// 执行测试
	service.processPriceData("test", data, assetMap, "test-source")

	// 验证数据库中的记录
	var savedPriceData models.PriceData
	err := db.Where("symbol = ?", "TEST").First(&savedPriceData).Error
	assert.NoError(t, err)
	assert.Equal(t, 1.0, savedPriceData.Price)
	assert.Equal(t, "test-source", savedPriceData.Source)

	mockRedis.AssertExpectations(t)
	mockKafka.AssertExpectations(t)
}

func TestPriceService_GetPriceHistory(t *testing.T) {
	db := setupTestDB()
	mockRedis := new(MockRedisClient)
	mockKafka := new(MockKafkaProducer)
	
	cfg := &config.Config{}

	service := &PriceService{
		db:     db,
		redis:  mockRedis,
		kafka:  mockKafka,
		config: cfg,
	}

	// 创建测试价格历史数据
	now := time.Now()
	priceHistory := []models.PriceData{
		{
			Symbol:    "TEST",
			Price:     1.0,
			Currency:  "USD",
			Source:    "test",
			Timestamp: now.Add(-2 * time.Hour),
		},
		{
			Symbol:    "TEST",
			Price:     1.1,
			Currency:  "USD",
			Source:    "test",
			Timestamp: now.Add(-1 * time.Hour),
		},
		{
			Symbol:    "TEST",
			Price:     1.2,
			Currency:  "USD",
			Source:    "test",
			Timestamp: now,
		},
	}

	for _, price := range priceHistory {
		db.Create(&price)
	}

	// 测试获取价格历史
	from := now.Add(-3 * time.Hour)
	to := now.Add(1 * time.Hour)

	result, err := service.GetPriceHistory("TEST", from, to)
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, 1.0, result[0].Price)
	assert.Equal(t, 1.2, result[2].Price)
}

func TestPriceService_CategorizeNews(t *testing.T) {
	newsService := &NewsService{}

	tests := []struct {
		title       string
		description string
		keyword     string
		expected    string
	}{
		{
			title:       "USDT Stablecoin News",
			description: "Latest updates on USDT",
			keyword:     "stablecoin",
			expected:    "stablecoin",
		},
		{
			title:       "Treasury Bond Yields Rise",
			description: "Government bond yields increase",
			keyword:     "treasury",
			expected:    "treasury",
		},
		{
			title:       "DeFi Protocol Launch",
			description: "New decentralized finance protocol",
			keyword:     "defi",
			expected:    "defi",
		},
		{
			title:       "General Crypto News",
			description: "Some general cryptocurrency news",
			keyword:     "crypto",
			expected:    "general",
		},
	}

	for _, test := range tests {
		result := newsService.categorizeNews(test.title, test.description, test.keyword)
		assert.Equal(t, test.expected, result, "Failed for title: %s", test.title)
	}
}

func TestPriceService_CalculateRelevance(t *testing.T) {
	newsService := &NewsService{}

	tests := []struct {
		title       string
		description string
		keyword     string
		expected    float64
	}{
		{
			title:       "RWA Token Launch",
			description: "New real world assets token",
			keyword:     "rwa",
			expected:    0.9, // 标题+描述+相关术语
		},
		{
			title:       "Stablecoin News",
			description: "General news",
			keyword:     "stablecoin",
			expected:    0.6, // 标题+相关术语
		},
		{
			title:       "General News",
			description: "Some description with stablecoin",
			keyword:     "stablecoin",
			expected:    0.4, // 描述+相关术语
		},
	}

	for _, test := range tests {
		result := newsService.calculateRelevance(test.title, test.description, test.keyword)
		assert.InDelta(t, test.expected, result, 0.1, "Failed for title: %s", test.title)
	}
}
