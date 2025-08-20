package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rwa-platform/data-collector/internal/services"
)

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   "data-collector",
	})
}

// GetPrice 获取资产价格
func GetPrice(priceService *services.PriceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		symbol := c.Param("symbol")
		if symbol == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
			return
		}

		price, err := priceService.GetPrice(symbol)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "price not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": price,
		})
	}
}

// GetPriceHistory 获取价格历史
func GetPriceHistory(priceService *services.PriceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		symbol := c.Param("symbol")
		if symbol == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
			return
		}

		// 解析时间参数
		fromStr := c.Query("from")
		toStr := c.Query("to")
		
		var from, to time.Time
		var err error

		if fromStr != "" {
			from, err = time.Parse(time.RFC3339, fromStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from time format"})
				return
			}
		} else {
			from = time.Now().AddDate(0, 0, -7) // 默认7天前
		}

		if toStr != "" {
			to, err = time.Parse(time.RFC3339, toStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to time format"})
				return
			}
		} else {
			to = time.Now()
		}

		history, err := priceService.GetPriceHistory(symbol, from, to)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get price history"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": history,
			"meta": gin.H{
				"symbol": symbol,
				"from":   from,
				"to":     to,
				"count":  len(history),
			},
		})
	}
}

// GetAssetInfo 获取资产信息
func GetAssetInfo(blockchainService *services.BlockchainService) gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Param("address")
		if address == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "address is required"})
			return
		}

		asset, err := blockchainService.GetAssetInfo(address)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": asset,
		})
	}
}

// GetTransaction 获取交易信息
func GetTransaction(blockchainService *services.BlockchainService) gin.HandlerFunc {
	return func(c *gin.Context) {
		hash := c.Param("hash")
		if hash == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "hash is required"})
			return
		}

		transaction, err := blockchainService.GetTransaction(hash)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": transaction,
		})
	}
}

// GetNews 获取新闻列表
func GetNews(newsService *services.NewsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析分页参数
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		
		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 20
		}

		// 解析筛选参数
		category := c.Query("category")
		source := c.Query("source")
		language := c.Query("language")

		news, total, err := newsService.GetNews(page, limit, category, source, language)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get news"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": news,
			"meta": gin.H{
				"page":       page,
				"limit":      limit,
				"total":      total,
				"total_pages": (total + limit - 1) / limit,
			},
		})
	}
}

// GetNewsDetail 获取新闻详情
func GetNewsDetail(newsService *services.NewsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
			return
		}

		news, err := newsService.GetNewsDetail(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "news not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": news,
		})
	}
}

// TriggerPriceSync 触发价格同步
func TriggerPriceSync(priceService *services.PriceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := priceService.TriggerSync(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to trigger sync"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "price sync triggered successfully",
		})
	}
}

// TriggerBlockchainSync 触发区块链同步
func TriggerBlockchainSync(blockchainService *services.BlockchainService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := blockchainService.TriggerSync(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to trigger sync"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "blockchain sync triggered successfully",
		})
	}
}

// GetStats 获取统计信息
func GetStats(priceService *services.PriceService, blockchainService *services.BlockchainService, newsService *services.NewsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := gin.H{
			"timestamp": time.Now().Unix(),
			"service":   "data-collector",
		}

		// 这里可以添加各种统计信息
		// 例如：最近采集的数据量、错误率、性能指标等

		c.JSON(http.StatusOK, gin.H{
			"data": stats,
		})
	}
}

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			switch err.Type {
			case gin.ErrorTypeBind:
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid request format",
					"details": err.Error(),
				})
			case gin.ErrorTypePublic:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
			}
		}
	}
}

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里可以实现基于Redis的限流逻辑
		c.Next()
	}
}
