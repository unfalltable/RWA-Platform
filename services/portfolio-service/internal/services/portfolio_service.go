package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/rwa-platform/portfolio-service/internal/config"
	"github.com/rwa-platform/portfolio-service/internal/kafka"
	"github.com/rwa-platform/portfolio-service/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PortfolioService struct {
	db     *gorm.DB
	redis  *redis.Client
	kafka  *kafka.Producer
	config *config.Config
	logger *logrus.Logger
}

type Portfolio struct {
	UserID          string                 `json:"user_id"`
	TotalValue      float64                `json:"total_value"`
	TotalCost       float64                `json:"total_cost"`
	TotalReturn     float64                `json:"total_return"`
	TotalReturnPct  float64                `json:"total_return_pct"`
	DayChange       float64                `json:"day_change"`
	DayChangePct    float64                `json:"day_change_pct"`
	Positions       []Position             `json:"positions"`
	Allocation      AssetAllocation        `json:"allocation"`
	Performance     PerformanceMetrics     `json:"performance"`
	RiskMetrics     RiskMetrics            `json:"risk_metrics"`
	LastUpdated     time.Time              `json:"last_updated"`
}

type Position struct {
	ID              string                 `json:"id"`
	AssetID         string                 `json:"asset_id"`
	AssetName       string                 `json:"asset_name"`
	AssetType       string                 `json:"asset_type"`
	Quantity        float64                `json:"quantity"`
	AveragePrice    float64                `json:"average_price"`
	CurrentPrice    float64                `json:"current_price"`
	MarketValue     float64                `json:"market_value"`
	CostBasis       float64                `json:"cost_basis"`
	UnrealizedPnL   float64                `json:"unrealized_pnl"`
	UnrealizedPnLPct float64               `json:"unrealized_pnl_pct"`
	DayChange       float64                `json:"day_change"`
	DayChangePct    float64                `json:"day_change_pct"`
	Weight          float64                `json:"weight"`
	Channels        []PositionChannel      `json:"channels"`
	Transactions    []Transaction          `json:"transactions"`
	LastUpdated     time.Time              `json:"last_updated"`
}

type PositionChannel struct {
	ChannelID    string  `json:"channel_id"`
	ChannelName  string  `json:"channel_name"`
	Quantity     float64 `json:"quantity"`
	MarketValue  float64 `json:"market_value"`
	Weight       float64 `json:"weight"`
}

type Transaction struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // buy, sell, deposit, withdraw, dividend, fee
	AssetID     string    `json:"asset_id"`
	Quantity    float64   `json:"quantity"`
	Price       float64   `json:"price"`
	Amount      float64   `json:"amount"`
	Fee         float64   `json:"fee"`
	ChannelID   string    `json:"channel_id"`
	Timestamp   time.Time `json:"timestamp"`
	Status      string    `json:"status"`
}

type AssetAllocation struct {
	ByAssetType  map[string]AllocationItem `json:"by_asset_type"`
	ByChannel    map[string]AllocationItem `json:"by_channel"`
	ByRegion     map[string]AllocationItem `json:"by_region"`
	BySector     map[string]AllocationItem `json:"by_sector"`
}

type AllocationItem struct {
	Value       float64 `json:"value"`
	Weight      float64 `json:"weight"`
	Count       int     `json:"count"`
	Change24h   float64 `json:"change_24h"`
}

type PerformanceMetrics struct {
	Return1D    float64 `json:"return_1d"`
	Return7D    float64 `json:"return_7d"`
	Return30D   float64 `json:"return_30d"`
	Return90D   float64 `json:"return_90d"`
	Return1Y    float64 `json:"return_1y"`
	ReturnYTD   float64 `json:"return_ytd"`
	ReturnTotal float64 `json:"return_total"`
	CAGR        float64 `json:"cagr"`
	Volatility  float64 `json:"volatility"`
	SharpeRatio float64 `json:"sharpe_ratio"`
	MaxDrawdown float64 `json:"max_drawdown"`
}

type RiskMetrics struct {
	VaR95       float64 `json:"var_95"`
	VaR99       float64 `json:"var_99"`
	Beta        float64 `json:"beta"`
	Alpha       float64 `json:"alpha"`
	Correlation float64 `json:"correlation"`
	Volatility  float64 `json:"volatility"`
	Skewness    float64 `json:"skewness"`
	Kurtosis    float64 `json:"kurtosis"`
}

func NewPortfolioService(db *gorm.DB, redisClient *redis.Client, kafkaProducer *kafka.Producer, cfg *config.Config) *PortfolioService {
	return &PortfolioService{
		db:     db,
		redis:  redisClient,
		kafka:  kafkaProducer,
		config: cfg,
		logger: logrus.New(),
	}
}

func (s *PortfolioService) GetPortfolio(userID string) (*Portfolio, error) {
	s.logger.Debugf("Getting portfolio for user: %s", userID)

	// 先从缓存获取
	cacheKey := fmt.Sprintf("portfolio:%s", userID)
	cached, err := s.redis.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var portfolio Portfolio
		if err := json.Unmarshal([]byte(cached), &portfolio); err == nil {
			return &portfolio, nil
		}
	}

	// 从数据库构建投资组合
	portfolio, err := s.buildPortfolio(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to build portfolio: %v", err)
	}

	// 缓存结果
	data, _ := json.Marshal(portfolio)
	s.redis.Set(context.Background(), cacheKey, data, time.Duration(s.config.PortfolioCacheTTL)*time.Second)

	return portfolio, nil
}

func (s *PortfolioService) buildPortfolio(userID string) (*Portfolio, error) {
	// 获取用户所有持仓
	positions, err := s.getUserPositions(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user positions: %v", err)
	}

	// 计算投资组合总值
	totalValue := 0.0
	totalCost := 0.0
	dayChange := 0.0

	for _, position := range positions {
		totalValue += position.MarketValue
		totalCost += position.CostBasis
		dayChange += position.DayChange
	}

	totalReturn := totalValue - totalCost
	totalReturnPct := 0.0
	if totalCost > 0 {
		totalReturnPct = (totalReturn / totalCost) * 100
	}

	dayChangePct := 0.0
	if totalValue > 0 {
		dayChangePct = (dayChange / (totalValue - dayChange)) * 100
	}

	// 计算资产配置
	allocation := s.calculateAllocation(positions, totalValue)

	// 计算性能指标
	performance := s.calculatePerformance(userID, positions)

	// 计算风险指标
	riskMetrics := s.calculateRiskMetrics(userID, positions)

	portfolio := &Portfolio{
		UserID:         userID,
		TotalValue:     totalValue,
		TotalCost:      totalCost,
		TotalReturn:    totalReturn,
		TotalReturnPct: totalReturnPct,
		DayChange:      dayChange,
		DayChangePct:   dayChangePct,
		Positions:      positions,
		Allocation:     allocation,
		Performance:    performance,
		RiskMetrics:    riskMetrics,
		LastUpdated:    time.Now(),
	}

	return portfolio, nil
}

func (s *PortfolioService) getUserPositions(userID string) ([]Position, error) {
	var dbPositions []models.Position
	if err := s.db.Where("user_id = ? AND quantity > 0", userID).
		Preload("Asset").
		Preload("Channels").
		Find(&dbPositions).Error; err != nil {
		return nil, err
	}

	var positions []Position
	for _, dbPos := range dbPositions {
		// 获取当前价格
		currentPrice := s.getCurrentPrice(dbPos.AssetID)
		
		// 计算市场价值
		marketValue := dbPos.Quantity * currentPrice
		
		// 计算未实现盈亏
		unrealizedPnL := marketValue - dbPos.CostBasis
		unrealizedPnLPct := 0.0
		if dbPos.CostBasis > 0 {
			unrealizedPnLPct = (unrealizedPnL / dbPos.CostBasis) * 100
		}

		// 计算日变化
		dayChange := s.calculateDayChange(dbPos.AssetID, dbPos.Quantity)
		dayChangePct := 0.0
		if marketValue > 0 {
			dayChangePct = (dayChange / (marketValue - dayChange)) * 100
		}

		// 构建渠道信息
		var channels []PositionChannel
		for _, ch := range dbPos.Channels {
			channels = append(channels, PositionChannel{
				ChannelID:   ch.ChannelID,
				ChannelName: ch.ChannelName,
				Quantity:    ch.Quantity,
				MarketValue: ch.Quantity * currentPrice,
				Weight:      (ch.Quantity / dbPos.Quantity) * 100,
			})
		}

		// 获取交易历史
		transactions := s.getPositionTransactions(dbPos.ID)

		position := Position{
			ID:               dbPos.ID,
			AssetID:          dbPos.AssetID,
			AssetName:        dbPos.Asset.Name,
			AssetType:        dbPos.Asset.Type,
			Quantity:         dbPos.Quantity,
			AveragePrice:     dbPos.AveragePrice,
			CurrentPrice:     currentPrice,
			MarketValue:      marketValue,
			CostBasis:        dbPos.CostBasis,
			UnrealizedPnL:    unrealizedPnL,
			UnrealizedPnLPct: unrealizedPnLPct,
			DayChange:        dayChange,
			DayChangePct:     dayChangePct,
			Channels:         channels,
			Transactions:     transactions,
			LastUpdated:      time.Now(),
		}

		positions = append(positions, position)
	}

	return positions, nil
}

func (s *PortfolioService) calculateAllocation(positions []Position, totalValue float64) AssetAllocation {
	byAssetType := make(map[string]AllocationItem)
	byChannel := make(map[string]AllocationItem)
	byRegion := make(map[string]AllocationItem)
	bySector := make(map[string]AllocationItem)

	for _, position := range positions {
		weight := 0.0
		if totalValue > 0 {
			weight = (position.MarketValue / totalValue) * 100
		}

		// 按资产类型分配
		if item, exists := byAssetType[position.AssetType]; exists {
			item.Value += position.MarketValue
			item.Weight += weight
			item.Count++
			item.Change24h += position.DayChange
			byAssetType[position.AssetType] = item
		} else {
			byAssetType[position.AssetType] = AllocationItem{
				Value:     position.MarketValue,
				Weight:    weight,
				Count:     1,
				Change24h: position.DayChange,
			}
		}

		// 按渠道分配
		for _, channel := range position.Channels {
			channelWeight := 0.0
			if totalValue > 0 {
				channelWeight = (channel.MarketValue / totalValue) * 100
			}

			if item, exists := byChannel[channel.ChannelName]; exists {
				item.Value += channel.MarketValue
				item.Weight += channelWeight
				item.Count++
				byChannel[channel.ChannelName] = item
			} else {
				byChannel[channel.ChannelName] = AllocationItem{
					Value:  channel.MarketValue,
					Weight: channelWeight,
					Count:  1,
				}
			}
		}

		// TODO: 实现按地区和行业的分配逻辑
	}

	return AssetAllocation{
		ByAssetType: byAssetType,
		ByChannel:   byChannel,
		ByRegion:    byRegion,
		BySector:    bySector,
	}
}

func (s *PortfolioService) calculatePerformance(userID string, positions []Position) PerformanceMetrics {
	// 获取历史价值数据
	historicalValues := s.getHistoricalPortfolioValues(userID)

	// 计算各期间收益率
	return1D := s.calculatePeriodReturn(historicalValues, 1)
	return7D := s.calculatePeriodReturn(historicalValues, 7)
	return30D := s.calculatePeriodReturn(historicalValues, 30)
	return90D := s.calculatePeriodReturn(historicalValues, 90)
	return1Y := s.calculatePeriodReturn(historicalValues, 365)
	returnYTD := s.calculateYTDReturn(historicalValues)
	returnTotal := s.calculateTotalReturn(historicalValues)

	// 计算年化收益率
	cagr := s.calculateCAGR(historicalValues)

	// 计算波动率
	volatility := s.calculateVolatility(historicalValues)

	// 计算夏普比率
	sharpeRatio := s.calculateSharpeRatio(historicalValues)

	// 计算最大回撤
	maxDrawdown := s.calculateMaxDrawdown(historicalValues)

	return PerformanceMetrics{
		Return1D:    return1D,
		Return7D:    return7D,
		Return30D:   return30D,
		Return90D:   return90D,
		Return1Y:    return1Y,
		ReturnYTD:   returnYTD,
		ReturnTotal: returnTotal,
		CAGR:        cagr,
		Volatility:  volatility,
		SharpeRatio: sharpeRatio,
		MaxDrawdown: maxDrawdown,
	}
}

func (s *PortfolioService) calculateRiskMetrics(userID string, positions []Position) RiskMetrics {
	// 获取历史收益率数据
	returns := s.getHistoricalReturns(userID)

	// 计算VaR
	var95 := s.calculateVaR(returns, 0.95)
	var99 := s.calculateVaR(returns, 0.99)

	// 计算Beta和Alpha（相对于基准）
	benchmarkReturns := s.getBenchmarkReturns()
	beta := s.calculateBeta(returns, benchmarkReturns)
	alpha := s.calculateAlpha(returns, benchmarkReturns, beta)

	// 计算相关性
	correlation := s.calculateCorrelation(returns, benchmarkReturns)

	// 计算波动率
	volatility := s.calculateVolatility(returns)

	// 计算偏度和峰度
	skewness := s.calculateSkewness(returns)
	kurtosis := s.calculateKurtosis(returns)

	return RiskMetrics{
		VaR95:       var95,
		VaR99:       var99,
		Beta:        beta,
		Alpha:       alpha,
		Correlation: correlation,
		Volatility:  volatility,
		Skewness:    skewness,
		Kurtosis:    kurtosis,
	}
}

// 辅助方法
func (s *PortfolioService) getCurrentPrice(assetID string) float64 {
	// 从缓存或价格服务获取当前价格
	cacheKey := fmt.Sprintf("price:%s", assetID)
	price, err := s.redis.Get(context.Background(), cacheKey).Float64()
	if err == nil {
		return price
	}

	// 从数据库获取最新价格
	var priceData models.AssetPrice
	if err := s.db.Where("asset_id = ?", assetID).
		Order("timestamp DESC").
		First(&priceData).Error; err == nil {
		return priceData.Price
	}

	// 默认价格
	return 1.0
}

func (s *PortfolioService) calculateDayChange(assetID string, quantity float64) float64 {
	// 获取24小时前的价格
	yesterday := time.Now().AddDate(0, 0, -1)
	var priceData models.AssetPrice
	if err := s.db.Where("asset_id = ? AND timestamp <= ?", assetID, yesterday).
		Order("timestamp DESC").
		First(&priceData).Error; err != nil {
		return 0.0
	}

	currentPrice := s.getCurrentPrice(assetID)
	return (currentPrice - priceData.Price) * quantity
}

func (s *PortfolioService) getPositionTransactions(positionID string) []Transaction {
	var dbTransactions []models.Transaction
	if err := s.db.Where("position_id = ?", positionID).
		Order("timestamp DESC").
		Limit(10).
		Find(&dbTransactions).Error; err != nil {
		return []Transaction{}
	}

	var transactions []Transaction
	for _, tx := range dbTransactions {
		transactions = append(transactions, Transaction{
			ID:        tx.ID,
			Type:      tx.Type,
			AssetID:   tx.AssetID,
			Quantity:  tx.Quantity,
			Price:     tx.Price,
			Amount:    tx.Amount,
			Fee:       tx.Fee,
			ChannelID: tx.ChannelID,
			Timestamp: tx.Timestamp,
			Status:    tx.Status,
		})
	}

	return transactions
}

func (s *PortfolioService) getHistoricalPortfolioValues(userID string) []float64 {
	// 获取历史投资组合价值数据
	var values []models.PortfolioValue
	if err := s.db.Where("user_id = ?", userID).
		Order("date DESC").
		Limit(365).
		Find(&values).Error; err != nil {
		return []float64{}
	}

	var result []float64
	for _, v := range values {
		result = append(result, v.TotalValue)
	}

	return result
}

func (s *PortfolioService) getHistoricalReturns(userID string) []float64 {
	// 计算历史收益率
	values := s.getHistoricalPortfolioValues(userID)
	if len(values) < 2 {
		return []float64{}
	}

	var returns []float64
	for i := 1; i < len(values); i++ {
		if values[i] > 0 {
			ret := (values[i-1] - values[i]) / values[i]
			returns = append(returns, ret)
		}
	}

	return returns
}

func (s *PortfolioService) getBenchmarkReturns() []float64 {
	// 获取基准收益率（如市场指数）
	// TODO: 实现基准收益率获取逻辑
	return []float64{}
}

// 计算方法（简化实现）
func (s *PortfolioService) calculatePeriodReturn(values []float64, days int) float64 {
	if len(values) <= days {
		return 0.0
	}
	
	current := values[0]
	past := values[days]
	
	if past > 0 {
		return ((current - past) / past) * 100
	}
	return 0.0
}

func (s *PortfolioService) calculateYTDReturn(values []float64) float64 {
	// 计算年初至今收益率
	// TODO: 实现YTD收益率计算
	return 0.0
}

func (s *PortfolioService) calculateTotalReturn(values []float64) float64 {
	if len(values) < 2 {
		return 0.0
	}
	
	current := values[0]
	initial := values[len(values)-1]
	
	if initial > 0 {
		return ((current - initial) / initial) * 100
	}
	return 0.0
}

func (s *PortfolioService) calculateCAGR(values []float64) float64 {
	// 计算复合年增长率
	// TODO: 实现CAGR计算
	return 0.0
}

func (s *PortfolioService) calculateVolatility(data []float64) float64 {
	// 计算波动率
	// TODO: 实现波动率计算
	return 0.0
}

func (s *PortfolioService) calculateSharpeRatio(returns []float64) float64 {
	// 计算夏普比率
	// TODO: 实现夏普比率计算
	return 0.0
}

func (s *PortfolioService) calculateMaxDrawdown(values []float64) float64 {
	// 计算最大回撤
	// TODO: 实现最大回撤计算
	return 0.0
}

func (s *PortfolioService) calculateVaR(returns []float64, confidence float64) float64 {
	// 计算风险价值
	// TODO: 实现VaR计算
	return 0.0
}

func (s *PortfolioService) calculateBeta(returns, benchmarkReturns []float64) float64 {
	// 计算Beta系数
	// TODO: 实现Beta计算
	return 1.0
}

func (s *PortfolioService) calculateAlpha(returns, benchmarkReturns []float64, beta float64) float64 {
	// 计算Alpha系数
	// TODO: 实现Alpha计算
	return 0.0
}

func (s *PortfolioService) calculateCorrelation(returns1, returns2 []float64) float64 {
	// 计算相关性
	// TODO: 实现相关性计算
	return 0.0
}

func (s *PortfolioService) calculateSkewness(returns []float64) float64 {
	// 计算偏度
	// TODO: 实现偏度计算
	return 0.0
}

func (s *PortfolioService) calculateKurtosis(returns []float64) float64 {
	// 计算峰度
	// TODO: 实现峰度计算
	return 0.0
}

// Kafka事件处理
func (s *PortfolioService) HandleTransactionEvent(message []byte) error {
	// 处理交易事件，更新持仓
	var event map[string]interface{}
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	userID, ok := event["user_id"].(string)
	if !ok {
		return fmt.Errorf("invalid user_id in transaction event")
	}

	// 清除用户投资组合缓存
	cacheKey := fmt.Sprintf("portfolio:%s", userID)
	s.redis.Del(context.Background(), cacheKey)

	s.logger.Debugf("Handled transaction event for user: %s", userID)
	return nil
}

func (s *PortfolioService) HandleUserEvent(message []byte) error {
	// 处理用户事件
	return nil
}
