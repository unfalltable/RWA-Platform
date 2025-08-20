package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rwa-platform/channel-service/internal/config"
	"github.com/rwa-platform/channel-service/internal/kafka"
	"github.com/rwa-platform/channel-service/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ChannelService struct {
	db     *gorm.DB
	redis  *redis.Client
	kafka  *kafka.Producer
	config *config.Config
	logger *logrus.Logger
}

type ChannelSyncResult struct {
	ChannelID string
	Success   bool
	Error     string
	UpdatedAt time.Time
}

func NewChannelService(db *gorm.DB, redisClient *redis.Client, kafkaProducer *kafka.Producer, cfg *config.Config) *ChannelService {
	return &ChannelService{
		db:     db,
		redis:  redisClient,
		kafka:  kafkaProducer,
		config: cfg,
		logger: logrus.New(),
	}
}

func (s *ChannelService) StartChannelSync(ctx context.Context) {
	s.logger.Info("Starting channel synchronization service")
	
	ticker := time.NewTicker(time.Duration(s.config.ChannelSyncInterval) * time.Second)
	defer ticker.Stop()

	// 立即执行一次
	s.syncAllChannels(ctx)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Channel sync service stopped")
			return
		case <-ticker.C:
			s.syncAllChannels(ctx)
		}
	}
}

func (s *ChannelService) syncAllChannels(ctx context.Context) {
	s.logger.Info("Starting channel synchronization cycle")

	// 获取所有活跃渠道
	var channels []models.Channel
	if err := s.db.Where("status = ? AND is_active = ?", "active", true).Find(&channels).Error; err != nil {
		s.logger.Errorf("Failed to fetch channels: %v", err)
		return
	}

	if len(channels) == 0 {
		s.logger.Warn("No active channels found for synchronization")
		return
	}

	// 并发同步渠道
	semaphore := make(chan struct{}, s.config.MaxConcurrentSyncs)
	var wg sync.WaitGroup
	results := make(chan ChannelSyncResult, len(channels))

	for _, channel := range channels {
		wg.Add(1)
		go func(ch models.Channel) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := s.syncChannel(ctx, ch)
			results <- result
		}(channel)
	}

	// 等待所有同步完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	successCount := 0
	errorCount := 0
	for result := range results {
		if result.Success {
			successCount++
		} else {
			errorCount++
			s.logger.Errorf("Channel sync failed for %s: %s", result.ChannelID, result.Error)
		}
	}

	s.logger.Infof("Channel sync completed: %d success, %d errors", successCount, errorCount)

	// 发布同步完成事件
	s.publishSyncEvent(successCount, errorCount)
}

func (s *ChannelService) syncChannel(ctx context.Context, channel models.Channel) ChannelSyncResult {
	result := ChannelSyncResult{
		ChannelID: channel.ID,
		UpdatedAt: time.Now(),
	}

	s.logger.Debugf("Syncing channel: %s (%s)", channel.Name, channel.Type)

	switch channel.Type {
	case "exchange":
		err := s.syncExchangeChannel(ctx, &channel)
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Success = true
		}
	case "broker":
		err := s.syncBrokerChannel(ctx, &channel)
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Success = true
		}
	case "dex":
		err := s.syncDEXChannel(ctx, &channel)
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Success = true
		}
	default:
		result.Error = fmt.Sprintf("unsupported channel type: %s", channel.Type)
	}

	// 更新渠道最后同步时间
	if result.Success {
		s.db.Model(&channel).Update("last_synced_at", time.Now())
		
		// 更新缓存
		s.updateChannelCache(&channel)
	}

	return result
}

func (s *ChannelService) syncExchangeChannel(ctx context.Context, channel *models.Channel) error {
	// 根据不同的交易所实现不同的同步逻辑
	switch channel.Name {
	case "coinbase":
		return s.syncCoinbaseChannel(ctx, channel)
	case "binance":
		return s.syncBinanceChannel(ctx, channel)
	case "kraken":
		return s.syncKrakenChannel(ctx, channel)
	default:
		return s.syncGenericExchangeChannel(ctx, channel)
	}
}

func (s *ChannelService) syncCoinbaseChannel(ctx context.Context, channel *models.Channel) error {
	// 实现Coinbase API同步逻辑
	s.logger.Debugf("Syncing Coinbase channel: %s", channel.ID)
	
	// 获取支持的资产列表
	assets, err := s.fetchCoinbaseAssets()
	if err != nil {
		return fmt.Errorf("failed to fetch Coinbase assets: %v", err)
	}

	// 获取费用信息
	fees, err := s.fetchCoinbaseFees()
	if err != nil {
		return fmt.Errorf("failed to fetch Coinbase fees: %v", err)
	}

	// 更新渠道信息
	channel.SupportedAssets = assets
	channel.Fees = fees
	channel.LastSyncedAt = time.Now()

	// 保存到数据库
	if err := s.db.Save(channel).Error; err != nil {
		return fmt.Errorf("failed to save channel: %v", err)
	}

	return nil
}

func (s *ChannelService) syncBinanceChannel(ctx context.Context, channel *models.Channel) error {
	// 实现Binance API同步逻辑
	s.logger.Debugf("Syncing Binance channel: %s", channel.ID)
	
	// 类似Coinbase的实现
	// ...
	
	return nil
}

func (s *ChannelService) syncKrakenChannel(ctx context.Context, channel *models.Channel) error {
	// 实现Kraken API同步逻辑
	s.logger.Debugf("Syncing Kraken channel: %s", channel.ID)
	
	// 类似Coinbase的实现
	// ...
	
	return nil
}

func (s *ChannelService) syncBrokerChannel(ctx context.Context, channel *models.Channel) error {
	// 实现券商渠道同步逻辑
	s.logger.Debugf("Syncing broker channel: %s", channel.ID)
	
	// 券商通常需要不同的API调用
	// ...
	
	return nil
}

func (s *ChannelService) syncDEXChannel(ctx context.Context, channel *models.Channel) error {
	// 实现DEX渠道同步逻辑
	s.logger.Debugf("Syncing DEX channel: %s", channel.ID)
	
	// DEX通常通过智能合约查询
	// ...
	
	return nil
}

func (s *ChannelService) syncGenericExchangeChannel(ctx context.Context, channel *models.Channel) error {
	// 通用交易所同步逻辑
	s.logger.Debugf("Syncing generic exchange channel: %s", channel.ID)
	
	// 基础的同步逻辑
	// ...
	
	return nil
}

func (s *ChannelService) fetchCoinbaseAssets() ([]models.ChannelAsset, error) {
	// 模拟Coinbase API调用
	assets := []models.ChannelAsset{
		{
			AssetID:      "usdt",
			AssetType:    "stablecoin",
			TradingPairs: []string{"USDT/USD", "USDT/EUR"},
			MinimumOrder: 1.0,
			MaximumOrder: 1000000.0,
			IsActive:     true,
		},
		{
			AssetID:      "usdc",
			AssetType:    "stablecoin",
			TradingPairs: []string{"USDC/USD", "USDC/EUR"},
			MinimumOrder: 1.0,
			MaximumOrder: 1000000.0,
			IsActive:     true,
		},
	}
	
	return assets, nil
}

func (s *ChannelService) fetchCoinbaseFees() (models.ChannelFees, error) {
	// 模拟Coinbase费用信息
	fees := models.ChannelFees{
		Trading: models.TradingFees{
			Maker: 0.005,
			Taker: 0.005,
		},
		Deposit: models.DepositFees{
			Crypto: 0.0,
			Fiat:   0.0,
			Wire:   25.0,
		},
		Withdrawal: models.WithdrawalFees{
			Crypto: 0.0005,
			Fiat:   0.15,
			Wire:   25.0,
		},
	}
	
	return fees, nil
}

func (s *ChannelService) updateChannelCache(channel *models.Channel) {
	cacheKey := fmt.Sprintf("channel:%s", channel.ID)
	
	data, err := json.Marshal(channel)
	if err != nil {
		s.logger.Errorf("Failed to marshal channel for cache: %v", err)
		return
	}

	if err := s.redis.Set(context.Background(), cacheKey, data, time.Duration(s.config.ChannelCacheTTL)*time.Second).Err(); err != nil {
		s.logger.Errorf("Failed to update channel cache for %s: %v", channel.ID, err)
	}
}

func (s *ChannelService) publishSyncEvent(successCount, errorCount int) {
	event := map[string]interface{}{
		"type":          "channel_sync_completed",
		"success_count": successCount,
		"error_count":   errorCount,
		"timestamp":     time.Now().Unix(),
	}

	if err := s.kafka.PublishMessage("channel-events", "sync", event); err != nil {
		s.logger.Errorf("Failed to publish sync event: %v", err)
	}
}

func (s *ChannelService) GetChannels(filters map[string]interface{}, page, limit int) ([]models.Channel, int, error) {
	var channels []models.Channel
	var total int64

	query := s.db.Model(&models.Channel{})

	// 应用筛选条件
	if channelType, ok := filters["type"]; ok {
		query = query.Where("type = ?", channelType)
	}
	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if region, ok := filters["region"]; ok {
		query = query.Where("supported_regions @> ?", fmt.Sprintf(`["%s"]`, region))
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&channels).Error; err != nil {
		return nil, 0, err
	}

	return channels, int(total), nil
}

func (s *ChannelService) GetChannelByID(id string) (*models.Channel, error) {
	// 先从缓存获取
	cacheKey := fmt.Sprintf("channel:%s", id)
	cached, err := s.redis.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var channel models.Channel
		if err := json.Unmarshal([]byte(cached), &channel); err == nil {
			return &channel, nil
		}
	}

	// 从数据库获取
	var channel models.Channel
	if err := s.db.Where("id = ?", id).First(&channel).Error; err != nil {
		return nil, err
	}

	// 更新缓存
	s.updateChannelCache(&channel)

	return &channel, nil
}

func (s *ChannelService) CreateChannel(channel *models.Channel) error {
	channel.CreatedAt = time.Now()
	channel.UpdatedAt = time.Now()
	
	if err := s.db.Create(channel).Error; err != nil {
		return err
	}

	// 发布创建事件
	s.publishChannelEvent("channel_created", channel)

	return nil
}

func (s *ChannelService) UpdateChannel(id string, updates map[string]interface{}) error {
	if err := s.db.Model(&models.Channel{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return err
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("channel:%s", id)
	s.redis.Del(context.Background(), cacheKey)

	// 发布更新事件
	channel, _ := s.GetChannelByID(id)
	if channel != nil {
		s.publishChannelEvent("channel_updated", channel)
	}

	return nil
}

func (s *ChannelService) publishChannelEvent(eventType string, channel *models.Channel) {
	event := map[string]interface{}{
		"type":       eventType,
		"channel_id": channel.ID,
		"channel":    channel,
		"timestamp":  time.Now().Unix(),
	}

	if err := s.kafka.PublishMessage("channel-events", channel.ID, event); err != nil {
		s.logger.Errorf("Failed to publish channel event: %v", err)
	}
}
