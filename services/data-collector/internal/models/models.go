package models

import (
	"time"

	"gorm.io/gorm"
)

// Asset 资产模型
type Asset struct {
	ID          string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Symbol      string    `gorm:"uniqueIndex;not null" json:"symbol"`
	Name        string    `gorm:"not null" json:"name"`
	Type        string    `gorm:"not null" json:"type"`
	Contracts   []byte    `gorm:"type:jsonb" json:"contracts"`
	Metadata    []byte    `gorm:"type:jsonb" json:"metadata"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// PriceData 价格数据模型
type PriceData struct {
	ID        string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssetID   string    `gorm:"type:uuid;not null;index" json:"asset_id"`
	Symbol    string    `gorm:"not null;index" json:"symbol"`
	Price     float64   `gorm:"type:decimal(20,8);not null" json:"price"`
	Currency  string    `gorm:"default:'USD'" json:"currency"`
	Volume24h *float64  `gorm:"type:decimal(20,2)" json:"volume_24h"`
	Change24h *float64  `gorm:"type:decimal(10,4)" json:"change_24h"`
	Change7d  *float64  `gorm:"type:decimal(10,4)" json:"change_7d"`
	Change30d *float64  `gorm:"type:decimal(10,4)" json:"change_30d"`
	MarketCap *float64  `gorm:"type:decimal(20,2)" json:"market_cap"`
	Source    string    `gorm:"not null" json:"source"`
	Timestamp time.Time `gorm:"not null;index" json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`

	// 关联
	Asset Asset `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
}

// BlockchainTransaction 区块链交易模型
type BlockchainTransaction struct {
	ID              string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Chain           string    `gorm:"not null;index" json:"chain"`
	Hash            string    `gorm:"not null;uniqueIndex" json:"hash"`
	BlockNumber     uint64    `gorm:"not null;index" json:"block_number"`
	BlockHash       string    `gorm:"not null" json:"block_hash"`
	TransactionIndex uint      `gorm:"not null" json:"transaction_index"`
	FromAddress     string    `gorm:"not null;index" json:"from_address"`
	ToAddress       *string   `gorm:"index" json:"to_address"`
	Value           string    `gorm:"type:decimal(78,0)" json:"value"`
	GasUsed         *uint64   `json:"gas_used"`
	GasPrice        *string   `gorm:"type:decimal(78,0)" json:"gas_price"`
	Status          *uint64   `json:"status"`
	ContractAddress *string   `json:"contract_address"`
	Logs            []byte    `gorm:"type:jsonb" json:"logs"`
	Timestamp       time.Time `gorm:"not null;index" json:"timestamp"`
	CreatedAt       time.Time `json:"created_at"`
}

// TokenTransfer 代币转账模型
type TokenTransfer struct {
	ID              string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Chain           string    `gorm:"not null;index" json:"chain"`
	TransactionHash string    `gorm:"not null;index" json:"transaction_hash"`
	LogIndex        uint      `gorm:"not null" json:"log_index"`
	ContractAddress string    `gorm:"not null;index" json:"contract_address"`
	FromAddress     string    `gorm:"not null;index" json:"from_address"`
	ToAddress       string    `gorm:"not null;index" json:"to_address"`
	Value           string    `gorm:"type:decimal(78,0);not null" json:"value"`
	TokenSymbol     *string   `json:"token_symbol"`
	TokenName       *string   `json:"token_name"`
	TokenDecimals   *uint8    `json:"token_decimals"`
	BlockNumber     uint64    `gorm:"not null;index" json:"block_number"`
	Timestamp       time.Time `gorm:"not null;index" json:"timestamp"`
	CreatedAt       time.Time `json:"created_at"`

	// 关联
	Transaction BlockchainTransaction `gorm:"foreignKey:TransactionHash;references:Hash" json:"transaction,omitempty"`
}

// NewsArticle 新闻文章模型
type NewsArticle struct {
	ID          string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title       string    `gorm:"not null" json:"title"`
	Content     *string   `gorm:"type:text" json:"content"`
	Summary     *string   `gorm:"type:text" json:"summary"`
	URL         string    `gorm:"not null;uniqueIndex" json:"url"`
	Source      string    `gorm:"not null;index" json:"source"`
	Author      *string   `json:"author"`
	Category    *string   `gorm:"index" json:"category"`
	Tags        []byte    `gorm:"type:jsonb" json:"tags"`
	Language    string    `gorm:"default:'en'" json:"language"`
	Sentiment   *float64  `gorm:"type:decimal(3,2)" json:"sentiment"` // -1 to 1
	Relevance   *float64  `gorm:"type:decimal(3,2)" json:"relevance"` // 0 to 1
	PublishedAt time.Time `gorm:"not null;index" json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DataSource 数据源模型
type DataSource struct {
	ID          string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"not null;uniqueIndex" json:"name"`
	Type        string    `gorm:"not null" json:"type"` // price, blockchain, news
	URL         string    `gorm:"not null" json:"url"`
	APIKey      *string   `json:"-"` // 不在JSON中暴露
	Config      []byte    `gorm:"type:jsonb" json:"config"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	RateLimit   *int      `json:"rate_limit"` // 每分钟请求数
	LastSyncAt  *time.Time `json:"last_sync_at"`
	NextSyncAt  *time.Time `json:"next_sync_at"`
	ErrorCount  int       `gorm:"default:0" json:"error_count"`
	LastError   *string   `json:"last_error"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SyncJob 同步任务模型
type SyncJob struct {
	ID           string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Type         string    `gorm:"not null;index" json:"type"` // price, blockchain, news
	Status       string    `gorm:"not null;index" json:"status"` // pending, running, completed, failed
	DataSourceID *string   `gorm:"type:uuid" json:"data_source_id"`
	Config       []byte    `gorm:"type:jsonb" json:"config"`
	Progress     int       `gorm:"default:0" json:"progress"` // 0-100
	RecordsTotal *int      `json:"records_total"`
	RecordsProcessed int   `gorm:"default:0" json:"records_processed"`
	RecordsSuccess   int   `gorm:"default:0" json:"records_success"`
	RecordsError     int   `gorm:"default:0" json:"records_error"`
	ErrorMessage     *string `json:"error_message"`
	StartedAt        *time.Time `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// 关联
	DataSource *DataSource `gorm:"foreignKey:DataSourceID" json:"data_source,omitempty"`
}

// MetricData 指标数据模型
type MetricData struct {
	ID        string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssetID   *string   `gorm:"type:uuid;index" json:"asset_id"`
	Chain     *string   `gorm:"index" json:"chain"`
	MetricType string   `gorm:"not null;index" json:"metric_type"` // volume, tvl, holders, etc.
	Value     float64   `gorm:"type:decimal(20,8);not null" json:"value"`
	Unit      string    `gorm:"not null" json:"unit"`
	Source    string    `gorm:"not null" json:"source"`
	Metadata  []byte    `gorm:"type:jsonb" json:"metadata"`
	Timestamp time.Time `gorm:"not null;index" json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`

	// 关联
	Asset *Asset `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
}

// 表名设置
func (Asset) TableName() string {
	return "assets"
}

func (PriceData) TableName() string {
	return "price_data"
}

func (BlockchainTransaction) TableName() string {
	return "blockchain_transactions"
}

func (TokenTransfer) TableName() string {
	return "token_transfers"
}

func (NewsArticle) TableName() string {
	return "news_articles"
}

func (DataSource) TableName() string {
	return "data_sources"
}

func (SyncJob) TableName() string {
	return "sync_jobs"
}

func (MetricData) TableName() string {
	return "metric_data"
}
