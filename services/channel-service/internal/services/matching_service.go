package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/rwa-platform/channel-service/internal/config"
	"github.com/rwa-platform/channel-service/internal/kafka"
	"github.com/rwa-platform/channel-service/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type MatchingService struct {
	db     *gorm.DB
	redis  *redis.Client
	kafka  *kafka.Producer
	config *config.Config
	logger *logrus.Logger
}

type MatchingRequest struct {
	AssetID       string  `json:"asset_id"`
	Amount        float64 `json:"amount"`
	UserRegion    string  `json:"user_region"`
	KYCLevel      string  `json:"kyc_level"`
	PaymentMethod string  `json:"payment_method"`
	UserID        string  `json:"user_id"`
	Preferences   map[string]interface{} `json:"preferences"`
}

type MatchingResult struct {
	ChannelID       string                 `json:"channel_id"`
	Channel         *models.Channel        `json:"channel"`
	MatchScore      float64                `json:"match_score"`
	EstimatedFees   *FeeEstimate          `json:"estimated_fees"`
	Availability    *ChannelAvailability   `json:"availability"`
	RedirectInfo    *RedirectInfo         `json:"redirect_info"`
	ProcessingTime  *ProcessingTime       `json:"processing_time"`
}

type FeeEstimate struct {
	TradingFee    float64 `json:"trading_fee"`
	WithdrawalFee float64 `json:"withdrawal_fee"`
	TotalFee      float64 `json:"total_fee"`
	Currency      string  `json:"currency"`
}

type ChannelAvailability struct {
	Available bool     `json:"available"`
	Reasons   []string `json:"reasons"`
}

type RedirectInfo struct {
	URL        string                 `json:"url"`
	Method     string                 `json:"method"`
	Parameters map[string]interface{} `json:"parameters"`
	ExpiresAt  time.Time             `json:"expires_at"`
}

type ProcessingTime struct {
	KYC        string `json:"kyc"`
	Deposit    string `json:"deposit"`
	Trade      string `json:"trade"`
	Withdrawal string `json:"withdrawal"`
}

func NewMatchingService(db *gorm.DB, redisClient *redis.Client, kafkaProducer *kafka.Producer, cfg *config.Config) *MatchingService {
	return &MatchingService{
		db:     db,
		redis:  redisClient,
		kafka:  kafkaProducer,
		config: cfg,
		logger: logrus.New(),
	}
}

func (s *MatchingService) StartMatchingEngine(ctx context.Context) {
	s.logger.Info("Starting matching engine")
	
	ticker := time.NewTicker(time.Duration(s.config.MatchingInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Matching engine stopped")
			return
		case <-ticker.C:
			s.processMatchingQueue(ctx)
		}
	}
}

func (s *MatchingService) processMatchingQueue(ctx context.Context) {
	// 处理待撮合的请求队列
	queueKey := "matching:queue"
	
	for {
		// 从队列中获取请求
		result, err := s.redis.BLPop(ctx, time.Second, queueKey).Result()
		if err != nil {
			if err == redis.Nil {
				break // 队列为空
			}
			s.logger.Errorf("Failed to pop from matching queue: %v", err)
			break
		}

		if len(result) < 2 {
			continue
		}

		// 解析请求
		var request MatchingRequest
		if err := json.Unmarshal([]byte(result[1]), &request); err != nil {
			s.logger.Errorf("Failed to unmarshal matching request: %v", err)
			continue
		}

		// 执行撮合
		matches, err := s.MatchChannels(&request)
		if err != nil {
			s.logger.Errorf("Failed to match channels for request: %v", err)
			continue
		}

		// 发布撮合结果
		s.publishMatchingResult(&request, matches)
	}
}

func (s *MatchingService) MatchChannels(request *MatchingRequest) ([]*MatchingResult, error) {
	s.logger.Debugf("Matching channels for asset %s, amount %f", request.AssetID, request.Amount)

	// 获取支持该资产的渠道
	channels, err := s.getEligibleChannels(request.AssetID, request.UserRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to get eligible channels: %v", err)
	}

	if len(channels) == 0 {
		return nil, fmt.Errorf("no eligible channels found for asset %s in region %s", request.AssetID, request.UserRegion)
	}

	// 计算每个渠道的匹配分数
	var results []*MatchingResult
	for _, channel := range channels {
		result := s.calculateChannelMatch(channel, request)
		if result.MatchScore >= s.config.MinMatchingScore {
			results = append(results, result)
		}
	}

	// 按匹配分数排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].MatchScore > results[j].MatchScore
	})

	// 限制返回结果数量
	if len(results) > s.config.MaxMatchingResults {
		results = results[:s.config.MaxMatchingResults]
	}

	return results, nil
}

func (s *MatchingService) getEligibleChannels(assetID, userRegion string) ([]*models.Channel, error) {
	var channels []*models.Channel

	// 从缓存获取
	cacheKey := fmt.Sprintf("eligible_channels:%s:%s", assetID, userRegion)
	cached, err := s.redis.Get(context.Background(), cacheKey).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(cached), &channels); err == nil {
			return channels, nil
		}
	}

	// 从数据库查询
	query := s.db.Where("status = ? AND is_active = ?", "active", true)
	
	// 检查支持的资产
	query = query.Where("supported_assets @> ?", fmt.Sprintf(`[{"asset_id": "%s"}]`, assetID))
	
	// 检查支持的地区
	query = query.Where("supported_regions @> ?", fmt.Sprintf(`["%s"]`, userRegion))

	if err := query.Find(&channels).Error; err != nil {
		return nil, err
	}

	// 缓存结果
	data, _ := json.Marshal(channels)
	s.redis.Set(context.Background(), cacheKey, data, 5*time.Minute)

	return channels, nil
}

func (s *MatchingService) calculateChannelMatch(channel *models.Channel, request *MatchingRequest) *MatchingResult {
	result := &MatchingResult{
		ChannelID: channel.ID,
		Channel:   channel,
	}

	// 计算匹配分数
	score := 0.0
	maxScore := 0.0

	// 1. 费用评分 (权重: 30%)
	feeScore := s.calculateFeeScore(channel, request.Amount)
	score += feeScore * 0.3
	maxScore += 0.3

	// 2. 可用性评分 (权重: 25%)
	availabilityScore := s.calculateAvailabilityScore(channel, request)
	score += availabilityScore * 0.25
	maxScore += 0.25

	// 3. 用户体验评分 (权重: 20%)
	uxScore := s.calculateUXScore(channel, request)
	score += uxScore * 0.2
	maxScore += 0.2

	// 4. 安全性评分 (权重: 15%)
	securityScore := s.calculateSecurityScore(channel)
	score += securityScore * 0.15
	maxScore += 0.15

	// 5. 流动性评分 (权重: 10%)
	liquidityScore := s.calculateLiquidityScore(channel, request.AssetID, request.Amount)
	score += liquidityScore * 0.1
	maxScore += 0.1

	// 标准化分数
	result.MatchScore = score / maxScore

	// 计算费用估算
	result.EstimatedFees = s.calculateFeeEstimate(channel, request.Amount)

	// 检查可用性
	result.Availability = s.checkChannelAvailability(channel, request)

	// 生成重定向信息
	result.RedirectInfo = s.generateRedirectInfo(channel, request)

	// 估算处理时间
	result.ProcessingTime = s.estimateProcessingTime(channel, request)

	return result
}

func (s *MatchingService) calculateFeeScore(channel *models.Channel, amount float64) float64 {
	// 计算总费用
	tradingFee := amount * channel.Fees.Trading.Taker
	withdrawalFee := channel.Fees.Withdrawal.Crypto
	totalFee := tradingFee + withdrawalFee

	// 费用越低分数越高
	// 假设最高费用为1%，最低费用为0.01%
	maxFee := amount * 0.01
	minFee := amount * 0.0001

	if totalFee <= minFee {
		return 1.0
	}
	if totalFee >= maxFee {
		return 0.0
	}

	return 1.0 - (totalFee-minFee)/(maxFee-minFee)
}

func (s *MatchingService) calculateAvailabilityScore(channel *models.Channel, request *MatchingRequest) float64 {
	score := 1.0

	// 检查KYC要求
	if channel.Compliance.KYCRequired && request.KYCLevel == "" {
		score -= 0.3
	}

	// 检查最小投资额
	if channel.Compliance.MinimumNetWorth > 0 {
		// 这里需要用户的净资产信息，暂时假设满足
		score -= 0.1
	}

	// 检查支付方式
	supportedPayment := false
	for _, method := range channel.PaymentMethods {
		if method.Method == request.PaymentMethod {
			supportedPayment = true
			break
		}
	}
	if !supportedPayment {
		score -= 0.4
	}

	return math.Max(0, score)
}

func (s *MatchingService) calculateUXScore(channel *models.Channel, request *MatchingRequest) float64 {
	score := 0.0

	// API可用性
	if channel.API != nil && channel.API.HasTradingAPI {
		score += 0.3
	}

	// 客服支持
	if channel.Support.Chat {
		score += 0.2
	}
	if channel.Support.Phone != "" {
		score += 0.2
	}

	// 响应时间
	if channel.Support.ResponseTime == "instant" {
		score += 0.3
	} else if channel.Support.ResponseTime == "1hour" {
		score += 0.2
	} else {
		score += 0.1
	}

	return score
}

func (s *MatchingService) calculateSecurityScore(channel *models.Channel) float64 {
	score := 0.0

	// 保险覆盖
	if channel.Security.Insurance != nil && channel.Security.Insurance.Coverage > 0 {
		score += 0.4
	}

	// 托管方式
	if channel.Security.Custody.Segregation {
		score += 0.3
	}

	// 审计报告
	if len(channel.Security.Audits) > 0 {
		score += 0.3
	}

	return score
}

func (s *MatchingService) calculateLiquidityScore(channel *models.Channel, assetID string, amount float64) float64 {
	// 这里需要实时的流动性数据
	// 暂时返回基于渠道类型的固定分数
	switch channel.Type {
	case "exchange":
		return 0.9
	case "broker":
		return 0.7
	case "dex":
		return 0.6
	default:
		return 0.5
	}
}

func (s *MatchingService) calculateFeeEstimate(channel *models.Channel, amount float64) *FeeEstimate {
	tradingFee := amount * channel.Fees.Trading.Taker
	withdrawalFee := channel.Fees.Withdrawal.Crypto
	
	return &FeeEstimate{
		TradingFee:    tradingFee,
		WithdrawalFee: withdrawalFee,
		TotalFee:      tradingFee + withdrawalFee,
		Currency:      "USD",
	}
}

func (s *MatchingService) checkChannelAvailability(channel *models.Channel, request *MatchingRequest) *ChannelAvailability {
	availability := &ChannelAvailability{
		Available: true,
		Reasons:   []string{},
	}

	// 检查各种可用性条件
	if channel.Compliance.KYCRequired && request.KYCLevel == "" {
		availability.Available = false
		availability.Reasons = append(availability.Reasons, "KYC verification required")
	}

	// 检查支付方式
	supportedPayment := false
	for _, method := range channel.PaymentMethods {
		if method.Method == request.PaymentMethod {
			supportedPayment = true
			break
		}
	}
	if !supportedPayment {
		availability.Available = false
		availability.Reasons = append(availability.Reasons, "Payment method not supported")
	}

	return availability
}

func (s *MatchingService) generateRedirectInfo(channel *models.Channel, request *MatchingRequest) *RedirectInfo {
	redirectID := uuid.New().String()
	
	// 构建重定向URL
	baseURL := channel.Website
	if channel.API != nil && channel.API.HasTradingAPI {
		baseURL = fmt.Sprintf("%s/api/redirect", baseURL)
	}

	// 构建参数
	params := map[string]interface{}{
		"asset_id":    request.AssetID,
		"amount":      request.Amount,
		"redirect_id": redirectID,
		"user_id":     request.UserID,
		"timestamp":   time.Now().Unix(),
	}

	redirectInfo := &RedirectInfo{
		URL:        fmt.Sprintf("%s?redirect_id=%s", baseURL, redirectID),
		Method:     "GET",
		Parameters: params,
		ExpiresAt:  time.Now().Add(time.Duration(s.config.RedirectExpiration) * time.Second),
	}

	// 缓存重定向信息
	s.cacheRedirectInfo(redirectID, redirectInfo, request)

	return redirectInfo
}

func (s *MatchingService) estimateProcessingTime(channel *models.Channel, request *MatchingRequest) *ProcessingTime {
	// 基于渠道类型和KYC要求估算处理时间
	processingTime := &ProcessingTime{
		Trade:      "instant",
		Withdrawal: "1-24 hours",
	}

	if channel.Compliance.KYCRequired {
		processingTime.KYC = "1-3 days"
		processingTime.Deposit = "1-2 hours"
	} else {
		processingTime.KYC = "not required"
		processingTime.Deposit = "instant"
	}

	return processingTime
}

func (s *MatchingService) cacheRedirectInfo(redirectID string, redirectInfo *RedirectInfo, request *MatchingRequest) {
	cacheKey := fmt.Sprintf("redirect:%s", redirectID)
	
	data := map[string]interface{}{
		"redirect_info": redirectInfo,
		"request":       request,
		"created_at":    time.Now(),
	}

	jsonData, _ := json.Marshal(data)
	s.redis.Set(context.Background(), cacheKey, jsonData, time.Duration(s.config.RedirectExpiration)*time.Second)
}

func (s *MatchingService) publishMatchingResult(request *MatchingRequest, results []*MatchingResult) {
	event := map[string]interface{}{
		"type":         "matching_completed",
		"request":      request,
		"results":      results,
		"result_count": len(results),
		"timestamp":    time.Now().Unix(),
	}

	if err := s.kafka.PublishMessage("matching-events", request.UserID, event); err != nil {
		s.logger.Errorf("Failed to publish matching result: %v", err)
	}
}

func (s *MatchingService) GetRedirectInfo(redirectID string) (map[string]interface{}, error) {
	cacheKey := fmt.Sprintf("redirect:%s", redirectID)
	
	data, err := s.redis.Get(context.Background(), cacheKey).Result()
	if err != nil {
		return nil, fmt.Errorf("redirect not found or expired")
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, fmt.Errorf("failed to parse redirect data")
	}

	return result, nil
}
