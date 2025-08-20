package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rwa-platform/channel-service/internal/services"
)

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   "channel-service",
	})
}

// GetChannels 获取渠道列表
func GetChannels(channelService *services.ChannelService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析查询参数
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		
		filters := make(map[string]interface{})
		if channelType := c.Query("type"); channelType != "" {
			filters["type"] = channelType
		}
		if status := c.Query("status"); status != "" {
			filters["status"] = status
		}
		if region := c.Query("region"); region != "" {
			filters["region"] = region
		}

		channels, total, err := channelService.GetChannels(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch channels",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": channels,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
				"pages": (total + limit - 1) / limit,
			},
		})
	}
}

// GetChannel 获取单个渠道
func GetChannel(channelService *services.ChannelService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		
		channel, err := channelService.GetChannelByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Channel not found",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": channel,
		})
	}
}

// CreateChannel 创建渠道
func CreateChannel(channelService *services.ChannelService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 添加认证和授权检查
		
		var channel struct {
			Name        string `json:"name" binding:"required"`
			Type        string `json:"type" binding:"required"`
			Description string `json:"description"`
			Website     string `json:"website"`
		}

		if err := c.ShouldBindJSON(&channel); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		// TODO: 实现创建逻辑
		c.JSON(http.StatusCreated, gin.H{
			"message": "Channel created successfully",
		})
	}
}

// UpdateChannel 更新渠道
func UpdateChannel(channelService *services.ChannelService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		
		var updates map[string]interface{}
		if err := c.ShouldBindJSON(&updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		if err := channelService.UpdateChannel(id, updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update channel",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Channel updated successfully",
		})
	}
}

// DeleteChannel 删除渠道
func DeleteChannel(channelService *services.ChannelService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现删除逻辑
		c.JSON(http.StatusOK, gin.H{
			"message": "Channel deleted successfully",
		})
	}
}

// GetChannelAssets 获取渠道支持的资产
func GetChannelAssets(channelService *services.ChannelService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		
		channel, err := channelService.GetChannelByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Channel not found",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": channel.SupportedAssets,
		})
	}
}

// SyncChannel 同步单个渠道
func SyncChannel(channelService *services.ChannelService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现单个渠道同步
		c.JSON(http.StatusOK, gin.H{
			"message": "Channel sync initiated",
		})
	}
}

// MatchChannels 渠道撮合
func MatchChannels(matchingService *services.MatchingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request services.MatchingRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		matches, err := matchingService.MatchChannels(&request)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to match channels",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": matches,
			"count": len(matches),
		})
	}
}

// GetQuote 获取报价
func GetQuote(matchingService *services.MatchingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现报价逻辑
		c.JSON(http.StatusOK, gin.H{
			"message": "Quote functionality not implemented yet",
		})
	}
}

// CreateRedirect 创建重定向
func CreateRedirect(matchingService *services.MatchingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现重定向创建逻辑
		c.JSON(http.StatusCreated, gin.H{
			"message": "Redirect created successfully",
		})
	}
}

// GetRedirect 获取重定向信息
func GetRedirect(matchingService *services.MatchingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		redirectID := c.Param("id")
		
		redirectInfo, err := matchingService.GetRedirectInfo(redirectID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Redirect not found or expired",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": redirectInfo,
		})
	}
}

// TrackAttribution 跟踪归因
func TrackAttribution(attributionService *services.AttributionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var event services.AttributionEvent
		if err := c.ShouldBindJSON(&event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		// 从请求头获取额外信息
		event.IPAddress = c.ClientIP()
		event.UserAgent = c.GetHeader("User-Agent")
		event.Referrer = c.GetHeader("Referer")

		if err := attributionService.TrackEvent(&event); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to track attribution event",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Attribution event tracked successfully",
		})
	}
}

// GetAttributionStats 获取归因统计
func GetAttributionStats(attributionService *services.AttributionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		channelID := c.Query("channel_id")
		period := c.DefaultQuery("period", time.Now().Format("2006-01-02"))

		if channelID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "channel_id is required",
			})
			return
		}

		stats, err := attributionService.GetAttributionStats(channelID, period)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Attribution stats not found",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": stats,
		})
	}
}

// GetConversions 获取转化数据
func GetConversions(attributionService *services.AttributionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		channelID := c.Query("channel_id")
		startDate := c.Query("start_date")
		endDate := c.Query("end_date")

		// 解析日期
		var start, end time.Time
		var err error
		
		if startDate != "" {
			start, err = time.Parse("2006-01-02", startDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid start_date format, use YYYY-MM-DD",
				})
				return
			}
		} else {
			start = time.Now().AddDate(0, 0, -30) // 默认30天前
		}

		if endDate != "" {
			end, err = time.Parse("2006-01-02", endDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid end_date format, use YYYY-MM-DD",
				})
				return
			}
		} else {
			end = time.Now()
		}

		conversions, err := attributionService.GetConversions(channelID, start, end)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch conversions",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": conversions,
			"count": len(conversions),
			"period": gin.H{
				"start": start.Format("2006-01-02"),
				"end":   end.Format("2006-01-02"),
			},
		})
	}
}

// GetSystemStats 获取系统统计
func GetSystemStats(
	channelService *services.ChannelService,
	matchingService *services.MatchingService,
	attributionService *services.AttributionService,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现系统统计逻辑
		c.JSON(http.StatusOK, gin.H{
			"message": "System stats functionality not implemented yet",
		})
	}
}

// SyncAllChannels 同步所有渠道
func SyncAllChannels(channelService *services.ChannelService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现所有渠道同步
		c.JSON(http.StatusOK, gin.H{
			"message": "All channels sync initiated",
		})
	}
}

// DetailedHealthCheck 详细健康检查
func DetailedHealthCheck(channelService *services.ChannelService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现详细健康检查
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"timestamp": time.Now().Unix(),
			"service": "channel-service",
			"version": "1.0.0",
			"checks": gin.H{
				"database": "healthy",
				"redis":    "healthy",
				"kafka":    "healthy",
			},
		})
	}
}
