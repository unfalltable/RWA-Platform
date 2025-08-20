package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rwa-platform/data-collector/internal/config"
	"github.com/rwa-platform/data-collector/internal/kafka"
	"github.com/rwa-platform/data-collector/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PriceService struct {
	db       *gorm.DB
	redis    *redis.Client
	kafka    *kafka.Producer
	config   *config.Config
	client   *http.Client
	logger   *logrus.Logger
}

type CoinGeckoResponse struct {
	Data map[string]CoinGeckoPrice `json:"data"`
}

type CoinGeckoPrice struct {
	ID                string  `json:"id"`
	Symbol            string  `json:"symbol"`
	Name              string  `json:"name"`
	CurrentPrice      float64 `json:"current_price"`
	MarketCap         float64 `json:"market_cap"`
	TotalVolume       float64 `json:"total_volume"`
	PriceChange24h    float64 `json:"price_change_24h"`
	PriceChangePercentage24h float64 `json:"price_change_percentage_24h"`
	PriceChangePercentage7d  float64 `json:"price_change_percentage_7d_in_currency"`
	PriceChangePercentage30d float64 `json:"price_change_percentage_30d_in_currency"`
	LastUpdated       string  `json:"last_updated"`
}

func NewPriceService(db *gorm.DB, redisClient *redis.Client, kafkaProducer *kafka.Producer, cfg *config.Config) *PriceService {
	return &PriceService{
		db:     db,
		redis:  redisClient,
		kafka:  kafkaProducer,
		config: cfg,
		client: &http.Client{
			Timeout: time.Duration(cfg.RequestTimeout) * time.Second,
		},
		logger: logrus.New(),
	}
}

func (s *PriceService) StartPriceCollection(ctx context.Context) {
	s.logger.Info("Starting price collection service")
	
	ticker := time.NewTicker(time.Duration(s.config.PriceCollectionInterval) * time.Second)
	defer ticker.Stop()

	// 立即执行一次
	s.collectPrices(ctx)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Price collection service stopped")
			return
		case <-ticker.C:
			s.collectPrices(ctx)
		}
	}
}

func (s *PriceService) collectPrices(ctx context.Context) {
	s.logger.Info("Starting price collection cycle")

	// 获取需要采集价格的资产列表
	var assets []models.Asset
	if err := s.db.Where("is_active = ?", true).Find(&assets).Error; err != nil {
		s.logger.Errorf("Failed to fetch assets: %v", err)
		return
	}

	if len(assets) == 0 {
		s.logger.Warn("No active assets found for price collection")
		return
	}

	// 按数据源分组采集
	s.collectFromCoinGecko(ctx, assets)
	s.collectFromCoinMarketCap(ctx, assets)

	s.logger.Infof("Price collection cycle completed for %d assets", len(assets))
}

func (s *PriceService) collectFromCoinGecko(ctx context.Context, assets []models.Asset) {
	if s.config.CoinGeckoAPIKey == "" {
		s.logger.Debug("CoinGecko API key not configured, skipping")
		return
	}

	// 构建符号列表
	symbols := make([]string, 0, len(assets))
	assetMap := make(map[string]models.Asset)
	
	for _, asset := range assets {
		symbols = append(symbols, strings.ToLower(asset.Symbol))
		assetMap[strings.ToLower(asset.Symbol)] = asset
	}

	// 分批处理，CoinGecko API限制
	batchSize := 100
	for i := 0; i < len(symbols); i += batchSize {
		end := i + batchSize
		if end > len(symbols) {
			end = len(symbols)
		}

		batch := symbols[i:end]
		s.fetchCoinGeckoPrices(ctx, batch, assetMap)
		
		// 避免触发API限制
		time.Sleep(1 * time.Second)
	}
}

func (s *PriceService) fetchCoinGeckoPrices(ctx context.Context, symbols []string, assetMap map[string]models.Asset) {
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd&include_market_cap=true&include_24hr_vol=true&include_24hr_change=true&include_7d_change=true&include_30d_change=true",
		strings.Join(symbols, ","))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		s.logger.Errorf("Failed to create CoinGecko request: %v", err)
		return
	}

	if s.config.CoinGeckoAPIKey != "" {
		req.Header.Set("X-CG-Demo-API-Key", s.config.CoinGeckoAPIKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Errorf("Failed to fetch from CoinGecko: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Errorf("CoinGecko API returned status %d", resp.StatusCode)
		return
	}

	var priceData map[string]map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&priceData); err != nil {
		s.logger.Errorf("Failed to decode CoinGecko response: %v", err)
		return
	}

	// 处理价格数据
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, s.config.MaxConcurrentRequests)

	for symbol, data := range priceData {
		wg.Add(1)
		go func(symbol string, data map[string]interface{}) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			s.processPriceData(symbol, data, assetMap, "coingecko")
		}(symbol, data)
	}

	wg.Wait()
}

func (s *PriceService) processPriceData(symbol string, data map[string]interface{}, assetMap map[string]models.Asset, source string) {
	asset, exists := assetMap[symbol]
	if !exists {
		return
	}

	// 提取价格数据
	price, ok := data["usd"].(float64)
	if !ok {
		s.logger.Warnf("Invalid price data for %s", symbol)
		return
	}

	priceData := &models.PriceData{
		AssetID:   asset.ID,
		Symbol:    asset.Symbol,
		Price:     price,
		Currency:  "USD",
		Source:    source,
		Timestamp: time.Now(),
	}

	// 可选字段
	if marketCap, ok := data["usd_market_cap"].(float64); ok {
		priceData.MarketCap = &marketCap
	}
	if volume, ok := data["usd_24h_vol"].(float64); ok {
		priceData.Volume24h = &volume
	}
	if change24h, ok := data["usd_24h_change"].(float64); ok {
		priceData.Change24h = &change24h
	}

	// 保存到数据库
	if err := s.db.Create(priceData).Error; err != nil {
		s.logger.Errorf("Failed to save price data for %s: %v", symbol, err)
		return
	}

	// 更新缓存
	s.updatePriceCache(asset.Symbol, priceData)

	// 发送到Kafka
	s.publishPriceUpdate(priceData)

	s.logger.Debugf("Updated price for %s: $%.4f", asset.Symbol, price)
}

func (s *PriceService) updatePriceCache(symbol string, priceData *models.PriceData) {
	cacheKey := fmt.Sprintf("price:%s", symbol)
	
	data, err := json.Marshal(priceData)
	if err != nil {
		s.logger.Errorf("Failed to marshal price data for cache: %v", err)
		return
	}

	if err := s.redis.Set(context.Background(), cacheKey, data, time.Duration(s.config.PriceCacheTTL)*time.Second).Err(); err != nil {
		s.logger.Errorf("Failed to update price cache for %s: %v", symbol, err)
	}
}

func (s *PriceService) publishPriceUpdate(priceData *models.PriceData) {
	message := map[string]interface{}{
		"type":      "price_update",
		"asset_id":  priceData.AssetID,
		"symbol":    priceData.Symbol,
		"price":     priceData.Price,
		"currency":  priceData.Currency,
		"source":    priceData.Source,
		"timestamp": priceData.Timestamp.Unix(),
	}

	if priceData.Change24h != nil {
		message["change_24h"] = *priceData.Change24h
	}

	if err := s.kafka.PublishMessage("price-updates", priceData.Symbol, message); err != nil {
		s.logger.Errorf("Failed to publish price update for %s: %v", priceData.Symbol, err)
	}
}

func (s *PriceService) collectFromCoinMarketCap(ctx context.Context, assets []models.Asset) {
	if s.config.CoinMarketCapAPIKey == "" {
		s.logger.Debug("CoinMarketCap API key not configured, skipping")
		return
	}

	// 构建符号列表
	symbols := make([]string, 0, len(assets))
	assetMap := make(map[string]models.Asset)

	for _, asset := range assets {
		symbols = append(symbols, strings.ToUpper(asset.Symbol))
		assetMap[strings.ToUpper(asset.Symbol)] = asset
	}

	// CoinMarketCap API调用
	url := "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		s.logger.Errorf("Failed to create CoinMarketCap request: %v", err)
		return
	}

	// 设置请求参数
	q := req.URL.Query()
	q.Add("symbol", strings.Join(symbols[:min(len(symbols), 100)], ",")) // 限制100个符号
	q.Add("convert", "USD")
	req.URL.RawQuery = q.Encode()

	req.Header.Set("X-CMC_PRO_API_KEY", s.config.CoinMarketCapAPIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Errorf("Failed to fetch from CoinMarketCap: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Errorf("CoinMarketCap API returned status %d", resp.StatusCode)
		return
	}

	var response struct {
		Data map[string]struct {
			Symbol string `json:"symbol"`
			Quote  map[string]struct {
				Price            float64 `json:"price"`
				Volume24h        float64 `json:"volume_24h"`
				PercentChange24h float64 `json:"percent_change_24h"`
				PercentChange7d  float64 `json:"percent_change_7d"`
				PercentChange30d float64 `json:"percent_change_30d"`
				MarketCap        float64 `json:"market_cap"`
			} `json:"quote"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		s.logger.Errorf("Failed to decode CoinMarketCap response: %v", err)
		return
	}

	// 处理响应数据
	for _, data := range response.Data {
		if usdQuote, exists := data.Quote["USD"]; exists {
			priceData := map[string]interface{}{
				"usd":                usdQuote.Price,
				"usd_market_cap":     usdQuote.MarketCap,
				"usd_24h_vol":        usdQuote.Volume24h,
				"usd_24h_change":     usdQuote.PercentChange24h,
				"usd_7d_change":      usdQuote.PercentChange7d,
				"usd_30d_change":     usdQuote.PercentChange30d,
			}
			s.processPriceData(strings.ToLower(data.Symbol), priceData, assetMap, "coinmarketcap")
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *PriceService) GetPrice(symbol string) (*models.PriceData, error) {
	// 先从缓存获取
	cacheKey := fmt.Sprintf("price:%s", symbol)
	cached, err := s.redis.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var priceData models.PriceData
		if err := json.Unmarshal([]byte(cached), &priceData); err == nil {
			return &priceData, nil
		}
	}

	// 从数据库获取最新价格
	var priceData models.PriceData
	if err := s.db.Where("symbol = ?", symbol).Order("timestamp DESC").First(&priceData).Error; err != nil {
		return nil, err
	}

	return &priceData, nil
}

func (s *PriceService) GetPriceHistory(symbol string, from, to time.Time) ([]models.PriceData, error) {
	var priceHistory []models.PriceData
	
	query := s.db.Where("symbol = ? AND timestamp BETWEEN ? AND ?", symbol, from, to).Order("timestamp ASC")
	
	if err := query.Find(&priceHistory).Error; err != nil {
		return nil, err
	}

	return priceHistory, nil
}

func (s *PriceService) TriggerSync() error {
	go s.collectPrices(context.Background())
	return nil
}
