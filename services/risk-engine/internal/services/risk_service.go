package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/rwa-platform/risk-engine/internal/config"
	"github.com/rwa-platform/risk-engine/internal/kafka"
	"github.com/rwa-platform/risk-engine/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type RiskService struct {
	db     *gorm.DB
	redis  *redis.Client
	kafka  *kafka.Producer
	config *config.Config
	logger *logrus.Logger
}

type RiskAssessmentRequest struct {
	UserID      string                 `json:"user_id"`
	AssetID     string                 `json:"asset_id"`
	ChannelID   string                 `json:"channel_id"`
	Amount      float64                `json:"amount"`
	Action      string                 `json:"action"` // invest, trade, withdraw
	Context     map[string]interface{} `json:"context"`
}

type RiskAssessmentResult struct {
	RiskScore     float64                `json:"risk_score"`
	RiskLevel     string                 `json:"risk_level"`
	Factors       []RiskFactor           `json:"factors"`
	Recommendations []string             `json:"recommendations"`
	Warnings      []string               `json:"warnings"`
	Approved      bool                   `json:"approved"`
	Conditions    []string               `json:"conditions"`
	ExpiresAt     time.Time              `json:"expires_at"`
}

type RiskFactor struct {
	Type        string  `json:"type"`
	Score       float64 `json:"score"`
	Weight      float64 `json:"weight"`
	Description string  `json:"description"`
	Impact      string  `json:"impact"`
}

type RiskProfile struct {
	UserID           string                 `json:"user_id"`
	RiskTolerance    string                 `json:"risk_tolerance"` // conservative, moderate, aggressive
	InvestmentGoals  []string               `json:"investment_goals"`
	TimeHorizon      string                 `json:"time_horizon"`
	LiquidityNeeds   string                 `json:"liquidity_needs"`
	ExperienceLevel  string                 `json:"experience_level"`
	FinancialStatus  FinancialStatus        `json:"financial_status"`
	RiskFactors      map[string]interface{} `json:"risk_factors"`
	LastUpdated      time.Time              `json:"last_updated"`
}

type FinancialStatus struct {
	NetWorth        float64 `json:"net_worth"`
	AnnualIncome    float64 `json:"annual_income"`
	LiquidAssets    float64 `json:"liquid_assets"`
	InvestmentRatio float64 `json:"investment_ratio"`
}

func NewRiskService(db *gorm.DB, redisClient *redis.Client, kafkaProducer *kafka.Producer, cfg *config.Config) *RiskService {
	return &RiskService{
		db:     db,
		redis:  redisClient,
		kafka:  kafkaProducer,
		config: cfg,
		logger: logrus.New(),
	}
}

func (s *RiskService) StartRiskMonitoring(ctx context.Context) {
	s.logger.Info("Starting risk monitoring service")
	
	ticker := time.NewTicker(time.Duration(s.config.RiskMonitoringInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Risk monitoring service stopped")
			return
		case <-ticker.C:
			s.performRiskMonitoring(ctx)
		}
	}
}

func (s *RiskService) performRiskMonitoring(ctx context.Context) {
	s.logger.Debug("Performing risk monitoring cycle")

	// 监控用户风险变化
	s.monitorUserRiskChanges(ctx)
	
	// 监控资产风险变化
	s.monitorAssetRiskChanges(ctx)
	
	// 监控市场风险
	s.monitorMarketRisk(ctx)
	
	// 监控系统性风险
	s.monitorSystemicRisk(ctx)
}

func (s *RiskService) AssessRisk(request *RiskAssessmentRequest) (*RiskAssessmentResult, error) {
	s.logger.Debugf("Assessing risk for user %s, asset %s, amount %f", 
		request.UserID, request.AssetID, request.Amount)

	// 获取用户风险档案
	userProfile, err := s.GetUserRiskProfile(request.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user risk profile: %v", err)
	}

	// 获取资产风险信息
	assetRisk, err := s.getAssetRisk(request.AssetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset risk: %v", err)
	}

	// 获取渠道风险信息
	channelRisk, err := s.getChannelRisk(request.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel risk: %v", err)
	}

	// 计算风险因子
	factors := s.calculateRiskFactors(request, userProfile, assetRisk, channelRisk)

	// 计算综合风险分数
	riskScore := s.calculateOverallRiskScore(factors)

	// 确定风险等级
	riskLevel := s.determineRiskLevel(riskScore)

	// 生成建议和警告
	recommendations := s.generateRecommendations(riskScore, factors, userProfile)
	warnings := s.generateWarnings(riskScore, factors)

	// 判断是否批准
	approved := s.isApproved(riskScore, riskLevel, userProfile)

	// 生成条件
	conditions := s.generateConditions(riskScore, factors)

	result := &RiskAssessmentResult{
		RiskScore:       riskScore,
		RiskLevel:       riskLevel,
		Factors:         factors,
		Recommendations: recommendations,
		Warnings:        warnings,
		Approved:        approved,
		Conditions:      conditions,
		ExpiresAt:       time.Now().Add(time.Duration(s.config.RiskAssessmentTTL) * time.Second),
	}

	// 记录风险评估
	s.recordRiskAssessment(request, result)

	// 发布风险评估事件
	s.publishRiskAssessmentEvent(request, result)

	return result, nil
}

func (s *RiskService) calculateRiskFactors(
	request *RiskAssessmentRequest,
	userProfile *RiskProfile,
	assetRisk *models.AssetRisk,
	channelRisk *models.ChannelRisk,
) []RiskFactor {
	var factors []RiskFactor

	// 1. 用户风险因子
	userFactor := s.calculateUserRiskFactor(userProfile, request.Amount)
	factors = append(factors, userFactor)

	// 2. 资产风险因子
	assetFactor := s.calculateAssetRiskFactor(assetRisk, request.Amount)
	factors = append(factors, assetFactor)

	// 3. 渠道风险因子
	channelFactor := s.calculateChannelRiskFactor(channelRisk)
	factors = append(factors, channelFactor)

	// 4. 市场风险因子
	marketFactor := s.calculateMarketRiskFactor(request.AssetID)
	factors = append(factors, marketFactor)

	// 5. 流动性风险因子
	liquidityFactor := s.calculateLiquidityRiskFactor(request.AssetID, request.Amount)
	factors = append(factors, liquidityFactor)

	// 6. 集中度风险因子
	concentrationFactor := s.calculateConcentrationRiskFactor(request.UserID, request.AssetID, request.Amount)
	factors = append(factors, concentrationFactor)

	return factors
}

func (s *RiskService) calculateUserRiskFactor(profile *RiskProfile, amount float64) RiskFactor {
	score := 0.0
	
	// 基于风险承受能力
	switch profile.RiskTolerance {
	case "conservative":
		score += 0.3
	case "moderate":
		score += 0.5
	case "aggressive":
		score += 0.8
	}

	// 基于投资比例
	if profile.FinancialStatus.InvestmentRatio > 0.8 {
		score += 0.3
	} else if profile.FinancialStatus.InvestmentRatio > 0.5 {
		score += 0.2
	}

	// 基于投资金额占净资产比例
	if profile.FinancialStatus.NetWorth > 0 {
		ratio := amount / profile.FinancialStatus.NetWorth
		if ratio > 0.1 {
			score += 0.3
		} else if ratio > 0.05 {
			score += 0.2
		}
	}

	return RiskFactor{
		Type:        "user_risk",
		Score:       math.Min(score, 1.0),
		Weight:      0.25,
		Description: "User risk profile and financial status",
		Impact:      s.getImpactLevel(score),
	}
}

func (s *RiskService) calculateAssetRiskFactor(assetRisk *models.AssetRisk, amount float64) RiskFactor {
	score := 0.0

	if assetRisk != nil {
		// 基于资产波动率
		score += assetRisk.Volatility * 0.4

		// 基于信用风险
		score += assetRisk.CreditRisk * 0.3

		// 基于流动性风险
		score += assetRisk.LiquidityRisk * 0.3
	} else {
		// 默认中等风险
		score = 0.5
	}

	return RiskFactor{
		Type:        "asset_risk",
		Score:       math.Min(score, 1.0),
		Weight:      0.3,
		Description: "Asset-specific risk characteristics",
		Impact:      s.getImpactLevel(score),
	}
}

func (s *RiskService) calculateChannelRiskFactor(channelRisk *models.ChannelRisk) RiskFactor {
	score := 0.0

	if channelRisk != nil {
		// 基于渠道安全评分
		score += (1.0 - channelRisk.SecurityScore) * 0.4

		// 基于合规评分
		score += (1.0 - channelRisk.ComplianceScore) * 0.3

		// 基于运营风险
		score += channelRisk.OperationalRisk * 0.3
	} else {
		// 默认中等风险
		score = 0.5
	}

	return RiskFactor{
		Type:        "channel_risk",
		Score:       math.Min(score, 1.0),
		Weight:      0.2,
		Description: "Channel security and operational risk",
		Impact:      s.getImpactLevel(score),
	}
}

func (s *RiskService) calculateMarketRiskFactor(assetID string) RiskFactor {
	// 获取市场风险指标
	marketVolatility := s.getMarketVolatility(assetID)
	marketTrend := s.getMarketTrend(assetID)

	score := marketVolatility * 0.6 + marketTrend * 0.4

	return RiskFactor{
		Type:        "market_risk",
		Score:       math.Min(score, 1.0),
		Weight:      0.15,
		Description: "Market volatility and trend risk",
		Impact:      s.getImpactLevel(score),
	}
}

func (s *RiskService) calculateLiquidityRiskFactor(assetID string, amount float64) RiskFactor {
	// 获取流动性指标
	liquidityScore := s.getLiquidityScore(assetID, amount)
	
	// 流动性越低，风险越高
	score := 1.0 - liquidityScore

	return RiskFactor{
		Type:        "liquidity_risk",
		Score:       math.Min(score, 1.0),
		Weight:      0.1,
		Description: "Asset liquidity and market depth risk",
		Impact:      s.getImpactLevel(score),
	}
}

func (s *RiskService) calculateConcentrationRiskFactor(userID, assetID string, amount float64) RiskFactor {
	// 获取用户当前持仓
	currentHoldings := s.getUserHoldings(userID)
	
	// 计算集中度
	concentrationRatio := s.calculateConcentrationRatio(currentHoldings, assetID, amount)
	
	score := 0.0
	if concentrationRatio > 0.5 {
		score = 0.8
	} else if concentrationRatio > 0.3 {
		score = 0.5
	} else if concentrationRatio > 0.2 {
		score = 0.3
	}

	return RiskFactor{
		Type:        "concentration_risk",
		Score:       score,
		Weight:      0.1,
		Description: "Portfolio concentration risk",
		Impact:      s.getImpactLevel(score),
	}
}

func (s *RiskService) calculateOverallRiskScore(factors []RiskFactor) float64 {
	totalScore := 0.0
	totalWeight := 0.0

	for _, factor := range factors {
		totalScore += factor.Score * factor.Weight
		totalWeight += factor.Weight
	}

	if totalWeight == 0 {
		return 0.5 // 默认中等风险
	}

	return totalScore / totalWeight
}

func (s *RiskService) determineRiskLevel(score float64) string {
	if score < 0.2 {
		return "very_low"
	} else if score < 0.4 {
		return "low"
	} else if score < 0.6 {
		return "medium"
	} else if score < 0.8 {
		return "high"
	} else {
		return "very_high"
	}
}

func (s *RiskService) generateRecommendations(score float64, factors []RiskFactor, profile *RiskProfile) []string {
	var recommendations []string

	if score > 0.7 {
		recommendations = append(recommendations, "Consider reducing investment amount")
		recommendations = append(recommendations, "Diversify across multiple assets")
	}

	if score > 0.5 && profile.RiskTolerance == "conservative" {
		recommendations = append(recommendations, "This investment may not align with your risk tolerance")
	}

	// 基于具体风险因子的建议
	for _, factor := range factors {
		if factor.Score > 0.7 {
			switch factor.Type {
			case "concentration_risk":
				recommendations = append(recommendations, "Consider diversifying your portfolio")
			case "liquidity_risk":
				recommendations = append(recommendations, "Ensure you have sufficient liquidity for your needs")
			case "market_risk":
				recommendations = append(recommendations, "Monitor market conditions closely")
			}
		}
	}

	return recommendations
}

func (s *RiskService) generateWarnings(score float64, factors []RiskFactor) []string {
	var warnings []string

	if score > 0.8 {
		warnings = append(warnings, "High risk investment - potential for significant losses")
	}

	for _, factor := range factors {
		if factor.Score > 0.8 {
			switch factor.Type {
			case "asset_risk":
				warnings = append(warnings, "Asset has high volatility and credit risk")
			case "channel_risk":
				warnings = append(warnings, "Channel has security or operational concerns")
			case "market_risk":
				warnings = append(warnings, "Market conditions are highly volatile")
			}
		}
	}

	return warnings
}

func (s *RiskService) isApproved(score float64, level string, profile *RiskProfile) bool {
	// 基于风险分数和用户风险承受能力判断
	switch profile.RiskTolerance {
	case "conservative":
		return score < 0.4
	case "moderate":
		return score < 0.7
	case "aggressive":
		return score < 0.9
	default:
		return score < 0.5
	}
}

func (s *RiskService) generateConditions(score float64, factors []RiskFactor) []string {
	var conditions []string

	if score > 0.6 {
		conditions = append(conditions, "Additional risk disclosure required")
		conditions = append(conditions, "Enhanced monitoring for 30 days")
	}

	if score > 0.8 {
		conditions = append(conditions, "Mandatory cooling-off period of 24 hours")
		conditions = append(conditions, "Risk manager approval required")
	}

	return conditions
}

// 辅助方法
func (s *RiskService) getImpactLevel(score float64) string {
	if score < 0.3 {
		return "low"
	} else if score < 0.7 {
		return "medium"
	} else {
		return "high"
	}
}

func (s *RiskService) getAssetRisk(assetID string) (*models.AssetRisk, error) {
	var assetRisk models.AssetRisk
	if err := s.db.Where("asset_id = ?", assetID).First(&assetRisk).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 没有风险数据，使用默认值
		}
		return nil, err
	}
	return &assetRisk, nil
}

func (s *RiskService) getChannelRisk(channelID string) (*models.ChannelRisk, error) {
	var channelRisk models.ChannelRisk
	if err := s.db.Where("channel_id = ?", channelID).First(&channelRisk).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &channelRisk, nil
}

func (s *RiskService) getMarketVolatility(assetID string) float64 {
	// 从缓存或数据库获取市场波动率
	cacheKey := fmt.Sprintf("market_volatility:%s", assetID)
	val, err := s.redis.Get(context.Background(), cacheKey).Float64()
	if err == nil {
		return val
	}
	
	// 默认值
	return 0.3
}

func (s *RiskService) getMarketTrend(assetID string) float64 {
	// 获取市场趋势指标
	return 0.2 // 默认值
}

func (s *RiskService) getLiquidityScore(assetID string, amount float64) float64 {
	// 计算流动性评分
	return 0.7 // 默认值
}

func (s *RiskService) getUserHoldings(userID string) map[string]float64 {
	// 获取用户持仓
	holdings := make(map[string]float64)
	// TODO: 从数据库获取实际持仓数据
	return holdings
}

func (s *RiskService) calculateConcentrationRatio(holdings map[string]float64, assetID string, amount float64) float64 {
	totalValue := amount
	for _, value := range holdings {
		totalValue += value
	}
	
	if totalValue == 0 {
		return 0
	}
	
	assetValue := holdings[assetID] + amount
	return assetValue / totalValue
}

func (s *RiskService) recordRiskAssessment(request *RiskAssessmentRequest, result *RiskAssessmentResult) {
	assessment := &models.RiskAssessment{
		ID:        uuid.New().String(),
		UserID:    request.UserID,
		AssetID:   request.AssetID,
		ChannelID: request.ChannelID,
		Amount:    request.Amount,
		Action:    request.Action,
		RiskScore: result.RiskScore,
		RiskLevel: result.RiskLevel,
		Approved:  result.Approved,
		Factors:   result.Factors,
		Context:   request.Context,
		CreatedAt: time.Now(),
		ExpiresAt: result.ExpiresAt,
	}

	if err := s.db.Create(assessment).Error; err != nil {
		s.logger.Errorf("Failed to record risk assessment: %v", err)
	}
}

func (s *RiskService) publishRiskAssessmentEvent(request *RiskAssessmentRequest, result *RiskAssessmentResult) {
	event := map[string]interface{}{
		"type":       "risk_assessment_completed",
		"request":    request,
		"result":     result,
		"timestamp":  time.Now().Unix(),
	}

	if err := s.kafka.PublishMessage("risk-events", request.UserID, event); err != nil {
		s.logger.Errorf("Failed to publish risk assessment event: %v", err)
	}
}

func (s *RiskService) GetUserRiskProfile(userID string) (*RiskProfile, error) {
	// 先从缓存获取
	cacheKey := fmt.Sprintf("risk_profile:%s", userID)
	cached, err := s.redis.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var profile RiskProfile
		if err := json.Unmarshal([]byte(cached), &profile); err == nil {
			return &profile, nil
		}
	}

	// 从数据库获取
	var profile models.UserRiskProfile
	if err := s.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建默认风险档案
			return s.createDefaultRiskProfile(userID)
		}
		return nil, err
	}

	result := &RiskProfile{
		UserID:          profile.UserID,
		RiskTolerance:   profile.RiskTolerance,
		InvestmentGoals: profile.InvestmentGoals,
		TimeHorizon:     profile.TimeHorizon,
		LiquidityNeeds:  profile.LiquidityNeeds,
		ExperienceLevel: profile.ExperienceLevel,
		FinancialStatus: FinancialStatus{
			NetWorth:        profile.NetWorth,
			AnnualIncome:    profile.AnnualIncome,
			LiquidAssets:    profile.LiquidAssets,
			InvestmentRatio: profile.InvestmentRatio,
		},
		RiskFactors: profile.RiskFactors,
		LastUpdated: profile.UpdatedAt,
	}

	// 缓存结果
	data, _ := json.Marshal(result)
	s.redis.Set(context.Background(), cacheKey, data, 30*time.Minute)

	return result, nil
}

func (s *RiskService) createDefaultRiskProfile(userID string) (*RiskProfile, error) {
	profile := &RiskProfile{
		UserID:          userID,
		RiskTolerance:   "moderate",
		InvestmentGoals: []string{"growth"},
		TimeHorizon:     "medium",
		LiquidityNeeds:  "medium",
		ExperienceLevel: "beginner",
		FinancialStatus: FinancialStatus{
			NetWorth:        0,
			AnnualIncome:    0,
			LiquidAssets:    0,
			InvestmentRatio: 0,
		},
		RiskFactors: make(map[string]interface{}),
		LastUpdated: time.Now(),
	}

	// 保存到数据库
	dbProfile := &models.UserRiskProfile{
		UserID:          profile.UserID,
		RiskTolerance:   profile.RiskTolerance,
		InvestmentGoals: profile.InvestmentGoals,
		TimeHorizon:     profile.TimeHorizon,
		LiquidityNeeds:  profile.LiquidityNeeds,
		ExperienceLevel: profile.ExperienceLevel,
		NetWorth:        profile.FinancialStatus.NetWorth,
		AnnualIncome:    profile.FinancialStatus.AnnualIncome,
		LiquidAssets:    profile.FinancialStatus.LiquidAssets,
		InvestmentRatio: profile.FinancialStatus.InvestmentRatio,
		RiskFactors:     profile.RiskFactors,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.db.Create(dbProfile).Error; err != nil {
		return nil, err
	}

	return profile, nil
}

// 监控方法
func (s *RiskService) monitorUserRiskChanges(ctx context.Context) {
	// 监控用户风险档案变化
	s.logger.Debug("Monitoring user risk changes")
}

func (s *RiskService) monitorAssetRiskChanges(ctx context.Context) {
	// 监控资产风险变化
	s.logger.Debug("Monitoring asset risk changes")
}

func (s *RiskService) monitorMarketRisk(ctx context.Context) {
	// 监控市场风险
	s.logger.Debug("Monitoring market risk")
}

func (s *RiskService) monitorSystemicRisk(ctx context.Context) {
	// 监控系统性风险
	s.logger.Debug("Monitoring systemic risk")
}

// Kafka事件处理
func (s *RiskService) HandleUserEvent(message []byte) error {
	// 处理用户事件
	return nil
}

func (s *RiskService) HandleTransactionEvent(message []byte) error {
	// 处理交易事件
	return nil
}

func (s *RiskService) HandleMarketEvent(message []byte) error {
	// 处理市场事件
	return nil
}
