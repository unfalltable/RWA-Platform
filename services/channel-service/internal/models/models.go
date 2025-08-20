package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
)

// Channel 渠道模型
type Channel struct {
	ID               string                 `json:"id" gorm:"primaryKey"`
	Name             string                 `json:"name" gorm:"not null"`
	DisplayName      string                 `json:"display_name"`
	Description      string                 `json:"description"`
	Type             string                 `json:"type" gorm:"not null"` // exchange, broker, dex, issuer, bank, platform
	Status           string                 `json:"status" gorm:"default:active"`
	IsActive         bool                   `json:"is_active" gorm:"default:true"`
	Website          string                 `json:"website"`
	Logo             string                 `json:"logo"`
	
	// 合规信息
	Compliance       ChannelCompliance      `json:"compliance" gorm:"embedded"`
	
	// 支持的资产
	SupportedAssets  []ChannelAsset         `json:"supported_assets" gorm:"type:jsonb"`
	
	// 费用信息
	Fees             ChannelFees            `json:"fees" gorm:"embedded"`
	
	// 支付方式
	PaymentMethods   []PaymentMethod        `json:"payment_methods" gorm:"type:jsonb"`
	
	// 客服支持
	Support          ChannelSupport         `json:"support" gorm:"embedded"`
	
	// API信息
	API              *ChannelAPI            `json:"api" gorm:"embedded"`
	
	// 安全信息
	Security         ChannelSecurity        `json:"security" gorm:"embedded"`
	
	// 元数据
	Metadata         datatypes.JSON         `json:"metadata"`
	
	// 时间戳
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	LastSyncedAt     *time.Time             `json:"last_synced_at"`
}

// ChannelCompliance 渠道合规信息
type ChannelCompliance struct {
	Licenses          []License    `json:"licenses" gorm:"type:jsonb"`
	SupportedRegions  pq.StringArray `json:"supported_regions" gorm:"type:text[]"`
	RestrictedRegions pq.StringArray `json:"restricted_regions" gorm:"type:text[]"`
	KYCRequired       bool         `json:"kyc_required" gorm:"default:false"`
	KYCLevels         pq.StringArray `json:"kyc_levels" gorm:"type:text[]"`
	AccreditedOnly    bool         `json:"accredited_only" gorm:"default:false"`
	MinimumNetWorth   float64      `json:"minimum_net_worth"`
}

// License 许可证信息
type License struct {
	Jurisdiction  string    `json:"jurisdiction"`
	LicenseType   string    `json:"license_type"`
	LicenseNumber string    `json:"license_number"`
	IssuedDate    time.Time `json:"issued_date"`
	ExpiryDate    *time.Time `json:"expiry_date"`
}

// ChannelAsset 渠道支持的资产
type ChannelAsset struct {
	AssetID      string         `json:"asset_id"`
	AssetType    string         `json:"asset_type"`
	TradingPairs pq.StringArray `json:"trading_pairs" gorm:"type:text[]"`
	MinimumOrder float64        `json:"minimum_order"`
	MaximumOrder float64        `json:"maximum_order"`
	IsActive     bool           `json:"is_active"`
}

// ChannelFees 渠道费用信息
type ChannelFees struct {
	Trading    TradingFees    `json:"trading" gorm:"embedded;embeddedPrefix:trading_"`
	Deposit    DepositFees    `json:"deposit" gorm:"embedded;embeddedPrefix:deposit_"`
	Withdrawal WithdrawalFees `json:"withdrawal" gorm:"embedded;embeddedPrefix:withdrawal_"`
	Management float64        `json:"management"`
}

// TradingFees 交易费用
type TradingFees struct {
	Maker float64 `json:"maker"`
	Taker float64 `json:"taker"`
	Flat  float64 `json:"flat"`
}

// DepositFees 存款费用
type DepositFees struct {
	Crypto float64 `json:"crypto"`
	Fiat   float64 `json:"fiat"`
	Wire   float64 `json:"wire"`
}

// WithdrawalFees 提款费用
type WithdrawalFees struct {
	Crypto float64 `json:"crypto"`
	Fiat   float64 `json:"fiat"`
	Wire   float64 `json:"wire"`
}

// PaymentMethod 支付方式
type PaymentMethod struct {
	Method       string        `json:"method"`
	Currencies   pq.StringArray `json:"currencies" gorm:"type:text[]"`
	ProcessingTime string      `json:"processing_time"`
	Limits       PaymentLimits `json:"limits"`
}

// PaymentLimits 支付限额
type PaymentLimits struct {
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Daily   float64 `json:"daily"`
	Monthly float64 `json:"monthly"`
}

// ChannelSupport 客服支持信息
type ChannelSupport struct {
	Email        string         `json:"email"`
	Phone        string         `json:"phone"`
	Chat         bool           `json:"chat"`
	Hours        string         `json:"hours"`
	Languages    pq.StringArray `json:"languages" gorm:"type:text[]"`
	ResponseTime string         `json:"response_time"`
}

// ChannelAPI API信息
type ChannelAPI struct {
	HasReadOnlyAPI bool       `json:"has_read_only_api"`
	HasTradingAPI  bool       `json:"has_trading_api"`
	Documentation  string     `json:"documentation"`
	RateLimits     RateLimits `json:"rate_limits"`
}

// RateLimits API限流信息
type RateLimits struct {
	Requests int    `json:"requests"`
	Period   string `json:"period"`
}

// ChannelSecurity 安全信息
type ChannelSecurity struct {
	Insurance *Insurance `json:"insurance"`
	Custody   CustodyInfo `json:"custody" gorm:"embedded;embeddedPrefix:custody_"`
	Audits    []Audit    `json:"audits" gorm:"type:jsonb"`
}

// Insurance 保险信息
type Insurance struct {
	Coverage float64 `json:"coverage"`
	Provider string  `json:"provider"`
}

// CustodyInfo 托管信息
type CustodyInfo struct {
	Type        string `json:"type"`
	Provider    string `json:"provider"`
	Segregation bool   `json:"segregation"`
}

// Audit 审计信息
type Audit struct {
	Auditor    string    `json:"auditor"`
	ReportDate time.Time `json:"report_date"`
	ReportURL  string    `json:"report_url"`
	Scope      string    `json:"scope"`
}

// AttributionEvent 归因事件
type AttributionEvent struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	UserID      string                 `json:"user_id" gorm:"not null;index"`
	SessionID   string                 `json:"session_id" gorm:"index"`
	EventType   string                 `json:"event_type" gorm:"not null"` // click, view, redirect, signup
	ChannelID   string                 `json:"channel_id" gorm:"index"`
	AssetID     string                 `json:"asset_id"`
	Amount      float64                `json:"amount"`
	RedirectID  string                 `json:"redirect_id"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Referrer    string                 `json:"referrer"`
	UTMSource   string                 `json:"utm_source"`
	UTMMedium   string                 `json:"utm_medium"`
	UTMCampaign string                 `json:"utm_campaign"`
	Metadata    datatypes.JSON         `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp" gorm:"index"`
}

// ConversionEvent 转化事件
type ConversionEvent struct {
	ID              string         `json:"id" gorm:"primaryKey"`
	UserID          string         `json:"user_id" gorm:"not null;index"`
	ChannelID       string         `json:"channel_id" gorm:"index"`
	AssetID         string         `json:"asset_id"`
	Amount          float64        `json:"amount"`
	Fee             float64        `json:"fee"`
	ConversionType  string         `json:"conversion_type"` // purchase, deposit, trade
	AttributionPath pq.StringArray `json:"attribution_path" gorm:"type:text[]"`
	Revenue         float64        `json:"revenue"`
	Timestamp       time.Time      `json:"timestamp" gorm:"index"`
}

// AttributionStats 归因统计
type AttributionStats struct {
	ID                string    `json:"id" gorm:"primaryKey"`
	ChannelID         string    `json:"channel_id" gorm:"not null;index"`
	TotalClicks       int64     `json:"total_clicks"`
	TotalConversions  int64     `json:"total_conversions"`
	ConversionRate    float64   `json:"conversion_rate"`
	TotalRevenue      float64   `json:"total_revenue"`
	AverageOrderValue float64   `json:"average_order_value"`
	Period            string    `json:"period" gorm:"index"` // YYYY-MM-DD
	UpdatedAt         time.Time `json:"updated_at"`
}

// ChannelRating 渠道评分
type ChannelRating struct {
	ID           string                `json:"id" gorm:"primaryKey"`
	ChannelID    string                `json:"channel_id" gorm:"not null;index"`
	OverallScore float64               `json:"overall_score"`
	Scores       ChannelRatingScores   `json:"scores" gorm:"embedded"`
	UserReviews  UserReviews           `json:"user_reviews" gorm:"embedded"`
	RiskEvents   []RiskEvent           `json:"risk_events" gorm:"type:jsonb"`
	UpdatedAt    time.Time             `json:"updated_at"`
}

// ChannelRatingScores 渠道评分详情
type ChannelRatingScores struct {
	Security       float64 `json:"security"`
	Compliance     float64 `json:"compliance"`
	Liquidity      float64 `json:"liquidity"`
	Fees           float64 `json:"fees"`
	UserExperience float64 `json:"user_experience"`
	Support        float64 `json:"support"`
	Reputation     float64 `json:"reputation"`
}

// UserReviews 用户评价
type UserReviews struct {
	TotalReviews    int            `json:"total_reviews"`
	AverageRating   float64        `json:"average_rating"`
	Distribution    datatypes.JSON `json:"distribution"`
}

// RiskEvent 风险事件
type RiskEvent struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"` // low, medium, high, critical
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Resolved    bool      `json:"resolved"`
	Impact      string    `json:"impact"`
}

// RedirectLog 重定向日志
type RedirectLog struct {
	ID         string                 `json:"id" gorm:"primaryKey"`
	UserID     string                 `json:"user_id" gorm:"not null;index"`
	ChannelID  string                 `json:"channel_id" gorm:"index"`
	AssetID    string                 `json:"asset_id"`
	Amount     float64                `json:"amount"`
	RedirectID string                 `json:"redirect_id" gorm:"unique;index"`
	Status     string                 `json:"status"` // pending, completed, expired, failed
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	Metadata   datatypes.JSON         `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
	CompletedAt *time.Time            `json:"completed_at"`
	ExpiresAt  time.Time              `json:"expires_at"`
}

// ChannelPerformance 渠道性能指标
type ChannelPerformance struct {
	ID                string    `json:"id" gorm:"primaryKey"`
	ChannelID         string    `json:"channel_id" gorm:"not null;index"`
	Date              time.Time `json:"date" gorm:"index"`
	TotalVolume       float64   `json:"total_volume"`
	TotalTransactions int64     `json:"total_transactions"`
	AverageResponseTime float64 `json:"average_response_time"`
	SuccessRate       float64   `json:"success_rate"`
	ErrorRate         float64   `json:"error_rate"`
	Uptime            float64   `json:"uptime"`
	UpdatedAt         time.Time `json:"updated_at"`
}
