package database

import (
	"fmt"
	"time"

	"github.com/rwa-platform/data-collector/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewConnection(databaseURL string) (*gorm.DB, error) {
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	db, err := gorm.Open(postgres.Open(databaseURL), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 自动迁移
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %v", err)
	}

	return db, nil
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Asset{},
		&models.PriceData{},
		&models.BlockchainTransaction{},
		&models.TokenTransfer{},
		&models.NewsArticle{},
		&models.DataSource{},
		&models.SyncJob{},
		&models.MetricData{},
	)
}

// 创建索引
func CreateIndexes(db *gorm.DB) error {
	// 价格数据索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_price_data_symbol_timestamp ON price_data(symbol, timestamp DESC)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_price_data_asset_timestamp ON price_data(asset_id, timestamp DESC)").Error; err != nil {
		return err
	}

	// 区块链交易索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_blockchain_tx_chain_block ON blockchain_transactions(chain, block_number DESC)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_blockchain_tx_from_addr ON blockchain_transactions(from_address)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_blockchain_tx_to_addr ON blockchain_transactions(to_address)").Error; err != nil {
		return err
	}

	// 代币转账索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_token_transfer_contract ON token_transfers(contract_address, block_number DESC)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_token_transfer_from ON token_transfers(from_address, timestamp DESC)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_token_transfer_to ON token_transfers(to_address, timestamp DESC)").Error; err != nil {
		return err
	}

	// 新闻文章索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_news_published_at ON news_articles(published_at DESC)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_news_source_category ON news_articles(source, category)").Error; err != nil {
		return err
	}

	// 指标数据索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_metric_data_asset_type ON metric_data(asset_id, metric_type, timestamp DESC)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_metric_data_chain_type ON metric_data(chain, metric_type, timestamp DESC)").Error; err != nil {
		return err
	}

	return nil
}

// 数据库健康检查
func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// 获取数据库统计信息
func GetStats(db *gorm.DB) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取各表的记录数
	var count int64

	// 资产数量
	if err := db.Model(&models.Asset{}).Count(&count).Error; err == nil {
		stats["assets_count"] = count
	}

	// 价格数据数量
	if err := db.Model(&models.PriceData{}).Count(&count).Error; err == nil {
		stats["price_data_count"] = count
	}

	// 区块链交易数量
	if err := db.Model(&models.BlockchainTransaction{}).Count(&count).Error; err == nil {
		stats["blockchain_transactions_count"] = count
	}

	// 代币转账数量
	if err := db.Model(&models.TokenTransfer{}).Count(&count).Error; err == nil {
		stats["token_transfers_count"] = count
	}

	// 新闻文章数量
	if err := db.Model(&models.NewsArticle{}).Count(&count).Error; err == nil {
		stats["news_articles_count"] = count
	}

	// 获取连接池统计
	sqlDB, err := db.DB()
	if err == nil {
		dbStats := sqlDB.Stats()
		stats["db_stats"] = map[string]interface{}{
			"open_connections":     dbStats.OpenConnections,
			"in_use":              dbStats.InUse,
			"idle":                dbStats.Idle,
			"wait_count":          dbStats.WaitCount,
			"wait_duration":       dbStats.WaitDuration.String(),
			"max_idle_closed":     dbStats.MaxIdleClosed,
			"max_idle_time_closed": dbStats.MaxIdleTimeClosed,
			"max_lifetime_closed": dbStats.MaxLifetimeClosed,
		}
	}

	return stats, nil
}

// 清理旧数据
func CleanupOldData(db *gorm.DB, days int) error {
	cutoffTime := time.Now().AddDate(0, 0, -days)

	// 清理旧的价格数据（保留最新的）
	if err := db.Where("timestamp < ? AND id NOT IN (SELECT DISTINCT ON (symbol) id FROM price_data ORDER BY symbol, timestamp DESC)", cutoffTime).Delete(&models.PriceData{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup old price data: %v", err)
	}

	// 清理旧的新闻文章
	if err := db.Where("published_at < ?", cutoffTime).Delete(&models.NewsArticle{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup old news articles: %v", err)
	}

	// 清理旧的指标数据
	if err := db.Where("timestamp < ?", cutoffTime).Delete(&models.MetricData{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup old metric data: %v", err)
	}

	return nil
}
