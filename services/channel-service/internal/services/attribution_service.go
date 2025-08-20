package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/rwa-platform/channel-service/internal/config"
	"github.com/rwa-platform/channel-service/internal/kafka"
	"github.com/rwa-platform/channel-service/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AttributionService struct {
	db     *gorm.DB
	redis  *redis.Client
	kafka  *kafka.Producer
	config *config.Config
	logger *logrus.Logger
}

type AttributionEvent struct {
	ID            string                 `json:"id"`
	UserID        string                 `json:"user_id"`
	SessionID     string                 `json:"session_id"`
	EventType     string                 `json:"event_type"`
	ChannelID     string                 `json:"channel_id"`
	AssetID       string                 `json:"asset_id"`
	Amount        float64                `json:"amount"`
	RedirectID    string                 `json:"redirect_id"`
	IPAddress     string                 `json:"ip_address"`
	UserAgent     string                 `json:"user_agent"`
	Referrer      string                 `json:"referrer"`
	UTMSource     string                 `json:"utm_source"`
	UTMMedium     string                 `json:"utm_medium"`
	UTMCampaign   string                 `json:"utm_campaign"`
	Metadata      map[string]interface{} `json:"metadata"`
	Timestamp     time.Time              `json:"timestamp"`
}

type ConversionEvent struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	ChannelID       string    `json:"channel_id"`
	AssetID         string    `json:"asset_id"`
	Amount          float64   `json:"amount"`
	Fee             float64   `json:"fee"`
	ConversionType  string    `json:"conversion_type"` // purchase, deposit, trade
	AttributionPath []string  `json:"attribution_path"`
	Revenue         float64   `json:"revenue"`
	Timestamp       time.Time `json:"timestamp"`
}

type AttributionStats struct {
	ChannelID       string  `json:"channel_id"`
	TotalClicks     int64   `json:"total_clicks"`
	TotalConversions int64  `json:"total_conversions"`
	ConversionRate  float64 `json:"conversion_rate"`
	TotalRevenue    float64 `json:"total_revenue"`
	AverageOrderValue float64 `json:"average_order_value"`
	Period          string  `json:"period"`
}

func NewAttributionService(db *gorm.DB, redisClient *redis.Client, kafkaProducer *kafka.Producer, cfg *config.Config) *AttributionService {
	return &AttributionService{
		db:     db,
		redis:  redisClient,
		kafka:  kafkaProducer,
		config: cfg,
		logger: logrus.New(),
	}
}

func (s *AttributionService) StartAttributionTracking(ctx context.Context) {
	s.logger.Info("Starting attribution tracking service")
	
	// 启动事件处理器
	go s.processAttributionEvents(ctx)
	
	// 启动转化事件处理器
	go s.processConversionEvents(ctx)
	
	// 启动统计计算器
	go s.calculateAttributionStats(ctx)
	
	<-ctx.Done()
	s.logger.Info("Attribution tracking service stopped")
}

func (s *AttributionService) processAttributionEvents(ctx context.Context) {
	queueKey := "attribution:events"
	
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 从队列中获取事件
			result, err := s.redis.BLPop(ctx, time.Second, queueKey).Result()
			if err != nil {
				if err == redis.Nil {
					continue
				}
				s.logger.Errorf("Failed to pop from attribution events queue: %v", err)
				continue
			}

			if len(result) < 2 {
				continue
			}

			// 解析事件
			var event AttributionEvent
			if err := json.Unmarshal([]byte(result[1]), &event); err != nil {
				s.logger.Errorf("Failed to unmarshal attribution event: %v", err)
				continue
			}

			// 处理事件
			s.handleAttributionEvent(&event)
		}
	}
}

func (s *AttributionService) processConversionEvents(ctx context.Context) {
	queueKey := "attribution:conversions"
	
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 从队列中获取转化事件
			result, err := s.redis.BLPop(ctx, time.Second, queueKey).Result()
			if err != nil {
				if err == redis.Nil {
					continue
				}
				s.logger.Errorf("Failed to pop from conversion events queue: %v", err)
				continue
			}

			if len(result) < 2 {
				continue
			}

			// 解析转化事件
			var event ConversionEvent
			if err := json.Unmarshal([]byte(result[1]), &event); err != nil {
				s.logger.Errorf("Failed to unmarshal conversion event: %v", err)
				continue
			}

			// 处理转化事件
			s.handleConversionEvent(&event)
		}
	}
}

func (s *AttributionService) calculateAttributionStats(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour) // 每小时计算一次统计数据
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.updateAttributionStats()
		}
	}
}

func (s *AttributionService) TrackEvent(event *AttributionEvent) error {
	// 生成事件ID
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	
	// 设置时间戳
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 验证必要字段
	if event.UserID == "" || event.EventType == "" {
		return fmt.Errorf("user_id and event_type are required")
	}

	// 存储到数据库
	attributionRecord := &models.AttributionEvent{
		ID:          event.ID,
		UserID:      event.UserID,
		SessionID:   event.SessionID,
		EventType:   event.EventType,
		ChannelID:   event.ChannelID,
		AssetID:     event.AssetID,
		Amount:      event.Amount,
		RedirectID:  event.RedirectID,
		IPAddress:   event.IPAddress,
		UserAgent:   event.UserAgent,
		Referrer:    event.Referrer,
		UTMSource:   event.UTMSource,
		UTMMedium:   event.UTMMedium,
		UTMCampaign: event.UTMCampaign,
		Metadata:    event.Metadata,
		Timestamp:   event.Timestamp,
	}

	if err := s.db.Create(attributionRecord).Error; err != nil {
		return fmt.Errorf("failed to save attribution event: %v", err)
	}

	// 更新用户归因路径
	s.updateUserAttributionPath(event)

	// 发布事件到Kafka
	s.publishAttributionEvent(event)

	s.logger.Debugf("Tracked attribution event: %s for user %s", event.EventType, event.UserID)
	return nil
}

func (s *AttributionService) TrackConversion(conversion *ConversionEvent) error {
	// 生成转化ID
	if conversion.ID == "" {
		conversion.ID = uuid.New().String()
	}
	
	// 设置时间戳
	if conversion.Timestamp.IsZero() {
		conversion.Timestamp = time.Now()
	}

	// 获取用户的归因路径
	attributionPath := s.getUserAttributionPath(conversion.UserID)
	conversion.AttributionPath = attributionPath

	// 存储到数据库
	conversionRecord := &models.ConversionEvent{
		ID:              conversion.ID,
		UserID:          conversion.UserID,
		ChannelID:       conversion.ChannelID,
		AssetID:         conversion.AssetID,
		Amount:          conversion.Amount,
		Fee:             conversion.Fee,
		ConversionType:  conversion.ConversionType,
		AttributionPath: attributionPath,
		Revenue:         conversion.Revenue,
		Timestamp:       conversion.Timestamp,
	}

	if err := s.db.Create(conversionRecord).Error; err != nil {
		return fmt.Errorf("failed to save conversion event: %v", err)
	}

	// 更新渠道统计
	s.updateChannelStats(conversion)

	// 发布转化事件到Kafka
	s.publishConversionEvent(conversion)

	s.logger.Debugf("Tracked conversion: %s for user %s, amount %f", conversion.ConversionType, conversion.UserID, conversion.Amount)
	return nil
}

func (s *AttributionService) handleAttributionEvent(event *AttributionEvent) {
	// 处理不同类型的归因事件
	switch event.EventType {
	case "click":
		s.handleClickEvent(event)
	case "view":
		s.handleViewEvent(event)
	case "redirect":
		s.handleRedirectEvent(event)
	case "signup":
		s.handleSignupEvent(event)
	default:
		s.logger.Warnf("Unknown attribution event type: %s", event.EventType)
	}
}

func (s *AttributionService) handleClickEvent(event *AttributionEvent) {
	// 记录点击事件
	clickKey := fmt.Sprintf("clicks:%s:%s", event.ChannelID, time.Now().Format("2006-01-02"))
	s.redis.Incr(context.Background(), clickKey)
	s.redis.Expire(context.Background(), clickKey, 30*24*time.Hour) // 保留30天
}

func (s *AttributionService) handleViewEvent(event *AttributionEvent) {
	// 记录浏览事件
	viewKey := fmt.Sprintf("views:%s:%s", event.ChannelID, time.Now().Format("2006-01-02"))
	s.redis.Incr(context.Background(), viewKey)
	s.redis.Expire(context.Background(), viewKey, 30*24*time.Hour)
}

func (s *AttributionService) handleRedirectEvent(event *AttributionEvent) {
	// 记录重定向事件
	redirectKey := fmt.Sprintf("redirects:%s:%s", event.ChannelID, time.Now().Format("2006-01-02"))
	s.redis.Incr(context.Background(), redirectKey)
	s.redis.Expire(context.Background(), redirectKey, 30*24*time.Hour)
}

func (s *AttributionService) handleSignupEvent(event *AttributionEvent) {
	// 记录注册事件
	signupKey := fmt.Sprintf("signups:%s:%s", event.ChannelID, time.Now().Format("2006-01-02"))
	s.redis.Incr(context.Background(), signupKey)
	s.redis.Expire(context.Background(), signupKey, 30*24*time.Hour)
}

func (s *AttributionService) handleConversionEvent(event *ConversionEvent) {
	// 更新转化统计
	conversionKey := fmt.Sprintf("conversions:%s:%s", event.ChannelID, time.Now().Format("2006-01-02"))
	s.redis.Incr(context.Background(), conversionKey)
	s.redis.Expire(context.Background(), conversionKey, 30*24*time.Hour)

	// 更新收入统计
	revenueKey := fmt.Sprintf("revenue:%s:%s", event.ChannelID, time.Now().Format("2006-01-02"))
	s.redis.IncrByFloat(context.Background(), revenueKey, event.Revenue)
	s.redis.Expire(context.Background(), revenueKey, 30*24*time.Hour)
}

func (s *AttributionService) updateUserAttributionPath(event *AttributionEvent) {
	// 更新用户的归因路径
	pathKey := fmt.Sprintf("attribution_path:%s", event.UserID)
	
	// 获取当前路径
	path, err := s.redis.LRange(context.Background(), pathKey, 0, -1).Result()
	if err != nil {
		path = []string{}
	}

	// 添加新的触点
	touchpoint := fmt.Sprintf("%s:%s:%d", event.ChannelID, event.EventType, event.Timestamp.Unix())
	s.redis.LPush(context.Background(), pathKey, touchpoint)
	
	// 限制路径长度
	s.redis.LTrim(context.Background(), pathKey, 0, 9) // 保留最近10个触点
	
	// 设置过期时间
	s.redis.Expire(context.Background(), pathKey, time.Duration(s.config.AttributionWindow)*time.Second)
}

func (s *AttributionService) getUserAttributionPath(userID string) []string {
	pathKey := fmt.Sprintf("attribution_path:%s", userID)
	path, err := s.redis.LRange(context.Background(), pathKey, 0, -1).Result()
	if err != nil {
		return []string{}
	}
	return path
}

func (s *AttributionService) updateChannelStats(conversion *ConversionEvent) {
	// 更新渠道的实时统计
	statsKey := fmt.Sprintf("channel_stats:%s", conversion.ChannelID)
	
	pipe := s.redis.Pipeline()
	pipe.HIncrBy(context.Background(), statsKey, "total_conversions", 1)
	pipe.HIncrByFloat(context.Background(), statsKey, "total_revenue", conversion.Revenue)
	pipe.HIncrByFloat(context.Background(), statsKey, "total_amount", conversion.Amount)
	pipe.Expire(context.Background(), statsKey, 24*time.Hour)
	
	_, err := pipe.Exec(context.Background())
	if err != nil {
		s.logger.Errorf("Failed to update channel stats: %v", err)
	}
}

func (s *AttributionService) updateAttributionStats() {
	s.logger.Debug("Updating attribution statistics")
	
	// 获取所有活跃渠道
	var channels []models.Channel
	if err := s.db.Where("status = ? AND is_active = ?", "active", true).Find(&channels).Error; err != nil {
		s.logger.Errorf("Failed to fetch channels for stats: %v", err)
		return
	}

	// 计算每个渠道的统计数据
	for _, channel := range channels {
		stats := s.calculateChannelStats(channel.ID)
		s.saveAttributionStats(stats)
	}
}

func (s *AttributionService) calculateChannelStats(channelID string) *AttributionStats {
	today := time.Now().Format("2006-01-02")
	
	// 获取点击数
	clickKey := fmt.Sprintf("clicks:%s:%s", channelID, today)
	clicks, _ := s.redis.Get(context.Background(), clickKey).Int64()
	
	// 获取转化数
	conversionKey := fmt.Sprintf("conversions:%s:%s", channelID, today)
	conversions, _ := s.redis.Get(context.Background(), conversionKey).Int64()
	
	// 获取收入
	revenueKey := fmt.Sprintf("revenue:%s:%s", channelID, today)
	revenue, _ := s.redis.Get(context.Background(), revenueKey).Float64()
	
	// 计算转化率
	var conversionRate float64
	if clicks > 0 {
		conversionRate = float64(conversions) / float64(clicks) * 100
	}
	
	// 计算平均订单价值
	var averageOrderValue float64
	if conversions > 0 {
		averageOrderValue = revenue / float64(conversions)
	}

	return &AttributionStats{
		ChannelID:         channelID,
		TotalClicks:       clicks,
		TotalConversions:  conversions,
		ConversionRate:    conversionRate,
		TotalRevenue:      revenue,
		AverageOrderValue: averageOrderValue,
		Period:            today,
	}
}

func (s *AttributionService) saveAttributionStats(stats *AttributionStats) {
	// 保存统计数据到数据库
	statsRecord := &models.AttributionStats{
		ChannelID:         stats.ChannelID,
		TotalClicks:       stats.TotalClicks,
		TotalConversions:  stats.TotalConversions,
		ConversionRate:    stats.ConversionRate,
		TotalRevenue:      stats.TotalRevenue,
		AverageOrderValue: stats.AverageOrderValue,
		Period:            stats.Period,
		UpdatedAt:         time.Now(),
	}

	// 使用UPSERT操作
	if err := s.db.Save(statsRecord).Error; err != nil {
		s.logger.Errorf("Failed to save attribution stats: %v", err)
	}
}

func (s *AttributionService) publishAttributionEvent(event *AttributionEvent) {
	eventData := map[string]interface{}{
		"type":  "attribution_event",
		"event": event,
	}

	if err := s.kafka.PublishMessage("attribution-events", event.UserID, eventData); err != nil {
		s.logger.Errorf("Failed to publish attribution event: %v", err)
	}
}

func (s *AttributionService) publishConversionEvent(event *ConversionEvent) {
	eventData := map[string]interface{}{
		"type":       "conversion_event",
		"conversion": event,
	}

	if err := s.kafka.PublishMessage("attribution-events", event.UserID, eventData); err != nil {
		s.logger.Errorf("Failed to publish conversion event: %v", err)
	}
}

func (s *AttributionService) GetAttributionStats(channelID string, period string) (*AttributionStats, error) {
	var stats models.AttributionStats
	if err := s.db.Where("channel_id = ? AND period = ?", channelID, period).First(&stats).Error; err != nil {
		return nil, err
	}

	return &AttributionStats{
		ChannelID:         stats.ChannelID,
		TotalClicks:       stats.TotalClicks,
		TotalConversions:  stats.TotalConversions,
		ConversionRate:    stats.ConversionRate,
		TotalRevenue:      stats.TotalRevenue,
		AverageOrderValue: stats.AverageOrderValue,
		Period:            stats.Period,
	}, nil
}

func (s *AttributionService) GetConversions(channelID string, startDate, endDate time.Time) ([]ConversionEvent, error) {
	var conversions []models.ConversionEvent
	query := s.db.Where("timestamp BETWEEN ? AND ?", startDate, endDate)
	
	if channelID != "" {
		query = query.Where("channel_id = ?", channelID)
	}
	
	if err := query.Find(&conversions).Error; err != nil {
		return nil, err
	}

	// 转换为返回类型
	var result []ConversionEvent
	for _, c := range conversions {
		result = append(result, ConversionEvent{
			ID:              c.ID,
			UserID:          c.UserID,
			ChannelID:       c.ChannelID,
			AssetID:         c.AssetID,
			Amount:          c.Amount,
			Fee:             c.Fee,
			ConversionType:  c.ConversionType,
			AttributionPath: c.AttributionPath,
			Revenue:         c.Revenue,
			Timestamp:       c.Timestamp,
		})
	}

	return result, nil
}
