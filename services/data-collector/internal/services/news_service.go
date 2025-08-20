package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rwa-platform/data-collector/internal/config"
	"github.com/rwa-platform/data-collector/internal/kafka"
	"github.com/rwa-platform/data-collector/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type NewsService struct {
	db     *gorm.DB
	redis  *redis.Client
	kafka  *kafka.Producer
	config *config.Config
	client *http.Client
	logger *logrus.Logger
}

type NewsAPIResponse struct {
	Status       string `json:"status"`
	TotalResults int    `json:"totalResults"`
	Articles     []struct {
		Source struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"source"`
		Author      string    `json:"author"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		URL         string    `json:"url"`
		URLToImage  string    `json:"urlToImage"`
		PublishedAt time.Time `json:"publishedAt"`
		Content     string    `json:"content"`
	} `json:"articles"`
}

func NewNewsService(db *gorm.DB, redisClient *redis.Client, kafkaProducer *kafka.Producer, cfg *config.Config) *NewsService {
	return &NewsService{
		db:     db,
		redis:  redisClient,
		kafka:  kafkaProducer,
		config: cfg,
		client: &http.Client{
			Timeout: time.Duration(cfg.RequestTimeout) * time.Second,
		},
		logger: logrus.New(),
	}
}

func (s *NewsService) StartNewsCollection(ctx context.Context) {
	s.logger.Info("Starting news collection service")
	
	ticker := time.NewTicker(time.Duration(s.config.NewsCollectionInterval) * time.Second)
	defer ticker.Stop()

	// 立即执行一次
	s.collectNews(ctx)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("News collection service stopped")
			return
		case <-ticker.C:
			s.collectNews(ctx)
		}
	}
}

func (s *NewsService) collectNews(ctx context.Context) {
	s.logger.Info("Starting news collection cycle")

	// 收集不同类型的新闻
	keywords := []string{
		"stablecoin",
		"treasury",
		"RWA",
		"real world assets",
		"USDT",
		"USDC",
		"DAI",
		"government bonds",
		"money market",
		"defi",
	}

	for _, keyword := range keywords {
		select {
		case <-ctx.Done():
			return
		default:
			s.collectNewsForKeyword(ctx, keyword)
			time.Sleep(2 * time.Second) // 避免API限制
		}
	}

	s.logger.Info("News collection cycle completed")
}

func (s *NewsService) collectNewsForKeyword(ctx context.Context, keyword string) {
	// 使用NewsAPI收集新闻
	if err := s.collectFromNewsAPI(ctx, keyword); err != nil {
		s.logger.Errorf("Failed to collect news from NewsAPI for keyword %s: %v", keyword, err)
	}

	// 可以添加其他新闻源
	// s.collectFromCryptoNews(ctx, keyword)
	// s.collectFromRSSFeeds(ctx, keyword)
}

func (s *NewsService) collectFromNewsAPI(ctx context.Context, keyword string) error {
	// 构建API请求
	baseURL := "https://newsapi.org/v2/everything"
	
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return err
	}

	// 设置查询参数
	q := req.URL.Query()
	q.Add("q", keyword)
	q.Add("language", "en")
	q.Add("sortBy", "publishedAt")
	q.Add("pageSize", "50")
	q.Add("from", time.Now().AddDate(0, 0, -1).Format("2006-01-02")) // 最近1天
	req.URL.RawQuery = q.Encode()

	// 设置API密钥（如果有的话）
	if apiKey := s.config.NewsAPIKey; apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("NewsAPI returned status %d", resp.StatusCode)
	}

	var newsResponse NewsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&newsResponse); err != nil {
		return err
	}

	// 处理新闻文章
	for _, article := range newsResponse.Articles {
		s.processNewsArticle(article, keyword)
	}

	s.logger.Debugf("Collected %d news articles for keyword: %s", len(newsResponse.Articles), keyword)
	return nil
}

func (s *NewsService) processNewsArticle(article struct {
	Source struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"source"`
	Author      string    `json:"author"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	URLToImage  string    `json:"urlToImage"`
	PublishedAt time.Time `json:"publishedAt"`
	Content     string    `json:"content"`
}, keyword string) {
	
	// 检查文章是否已存在
	var existingArticle models.NewsArticle
	if err := s.db.Where("url = ?", article.URL).First(&existingArticle).Error; err == nil {
		return // 文章已存在
	}

	// 创建新闻文章记录
	newsArticle := &models.NewsArticle{
		Title:       article.Title,
		URL:         article.URL,
		Source:      article.Source.Name,
		Language:    "en",
		PublishedAt: article.PublishedAt,
	}

	if article.Author != "" {
		newsArticle.Author = &article.Author
	}

	if article.Description != "" {
		newsArticle.Summary = &article.Description
	}

	if article.Content != "" {
		newsArticle.Content = &article.Content
	}

	// 设置分类
	category := s.categorizeNews(article.Title, article.Description, keyword)
	if category != "" {
		newsArticle.Category = &category
	}

	// 设置标签
	tags := s.extractTags(article.Title, article.Description, keyword)
	if len(tags) > 0 {
		tagsJSON, _ := json.Marshal(tags)
		newsArticle.Tags = tagsJSON
	}

	// 计算相关性分数
	relevance := s.calculateRelevance(article.Title, article.Description, keyword)
	newsArticle.Relevance = &relevance

	// 保存到数据库
	if err := s.db.Create(newsArticle).Error; err != nil {
		s.logger.Errorf("Failed to save news article: %v", err)
		return
	}

	// 发布到Kafka
	s.publishNewsUpdate(newsArticle)

	s.logger.Debugf("Saved news article: %s", article.Title)
}

func (s *NewsService) categorizeNews(title, description, keyword string) string {
	content := strings.ToLower(title + " " + description)
	
	if strings.Contains(content, "stablecoin") || strings.Contains(content, "usdt") || strings.Contains(content, "usdc") {
		return "stablecoin"
	}
	if strings.Contains(content, "treasury") || strings.Contains(content, "bond") {
		return "treasury"
	}
	if strings.Contains(content, "rwa") || strings.Contains(content, "real world asset") {
		return "rwa"
	}
	if strings.Contains(content, "defi") || strings.Contains(content, "decentralized finance") {
		return "defi"
	}
	if strings.Contains(content, "regulation") || strings.Contains(content, "regulatory") {
		return "regulation"
	}
	
	return "general"
}

func (s *NewsService) extractTags(title, description, keyword string) []string {
	content := strings.ToLower(title + " " + description)
	tags := []string{keyword}
	
	// 常见标签
	commonTags := []string{
		"bitcoin", "ethereum", "blockchain", "crypto", "cryptocurrency",
		"stablecoin", "defi", "rwa", "treasury", "bond", "yield",
		"regulation", "sec", "fed", "central bank", "cbdc",
	}
	
	for _, tag := range commonTags {
		if strings.Contains(content, tag) && !contains(tags, tag) {
			tags = append(tags, tag)
		}
	}
	
	return tags
}

func (s *NewsService) calculateRelevance(title, description, keyword string) float64 {
	content := strings.ToLower(title + " " + description)
	keyword = strings.ToLower(keyword)
	
	score := 0.0
	
	// 标题中包含关键词
	if strings.Contains(strings.ToLower(title), keyword) {
		score += 0.5
	}
	
	// 描述中包含关键词
	if strings.Contains(strings.ToLower(description), keyword) {
		score += 0.3
	}
	
	// 包含相关术语
	relevantTerms := []string{"rwa", "real world assets", "stablecoin", "treasury", "defi"}
	for _, term := range relevantTerms {
		if strings.Contains(content, term) {
			score += 0.1
		}
	}
	
	if score > 1.0 {
		score = 1.0
	}
	
	return score
}

func (s *NewsService) publishNewsUpdate(article *models.NewsArticle) {
	message := map[string]interface{}{
		"type":         "news_update",
		"id":           article.ID,
		"title":        article.Title,
		"url":          article.URL,
		"source":       article.Source,
		"category":     article.Category,
		"relevance":    article.Relevance,
		"published_at": article.PublishedAt.Unix(),
		"created_at":   article.CreatedAt.Unix(),
	}

	if err := s.kafka.PublishMessage("news-updates", article.ID, message); err != nil {
		s.logger.Errorf("Failed to publish news update: %v", err)
	}
}

func (s *NewsService) GetNews(page, limit int, category, source, language string) ([]models.NewsArticle, int, error) {
	var news []models.NewsArticle
	var total int64

	query := s.db.Model(&models.NewsArticle{})

	// 应用筛选条件
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if source != "" {
		query = query.Where("source = ?", source)
	}
	if language != "" {
		query = query.Where("language = ?", language)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Order("published_at DESC").Offset(offset).Limit(limit).Find(&news).Error; err != nil {
		return nil, 0, err
	}

	return news, int(total), nil
}

func (s *NewsService) GetNewsDetail(id string) (*models.NewsArticle, error) {
	var news models.NewsArticle
	if err := s.db.Where("id = ?", id).First(&news).Error; err != nil {
		return nil, err
	}
	return &news, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
