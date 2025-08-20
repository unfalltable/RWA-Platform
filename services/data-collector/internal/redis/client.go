package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type Client struct {
	client *redis.Client
	logger *logrus.Logger
}

func NewClient(redisURL string) (*Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %v", err)
	}

	client := redis.NewClient(opts)
	
	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &Client{
		client: client,
		logger: logrus.New(),
	}, nil
}

// 基础操作
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *Client) Get(ctx context.Context, key string) *redis.StringCmd {
	return c.client.Get(ctx, key)
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Exists(ctx, keys...).Result()
}

// JSON操作
func (c *Client) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

func (c *Client) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

// Hash操作
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) error {
	return c.client.HSet(ctx, key, values...).Err()
}

func (c *Client) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	return c.client.HGet(ctx, key, field)
}

func (c *Client) HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd {
	return c.client.HGetAll(ctx, key)
}

func (c *Client) HDel(ctx context.Context, key string, fields ...string) error {
	return c.client.HDel(ctx, key, fields...).Err()
}

// List操作
func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) error {
	return c.client.LPush(ctx, key, values...).Err()
}

func (c *Client) RPush(ctx context.Context, key string, values ...interface{}) error {
	return c.client.RPush(ctx, key, values...).Err()
}

func (c *Client) LPop(ctx context.Context, key string) *redis.StringCmd {
	return c.client.LPop(ctx, key)
}

func (c *Client) RPop(ctx context.Context, key string) *redis.StringCmd {
	return c.client.RPop(ctx, key)
}

func (c *Client) LRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return c.client.LRange(ctx, key, start, stop)
}

func (c *Client) LLen(ctx context.Context, key string) *redis.IntCmd {
	return c.client.LLen(ctx, key)
}

// Set操作
func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SAdd(ctx, key, members...).Err()
}

func (c *Client) SRem(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SRem(ctx, key, members...).Err()
}

func (c *Client) SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	return c.client.SMembers(ctx, key)
}

func (c *Client) SIsMember(ctx context.Context, key string, member interface{}) *redis.BoolCmd {
	return c.client.SIsMember(ctx, key, member)
}

// Sorted Set操作
func (c *Client) ZAdd(ctx context.Context, key string, members ...*redis.Z) error {
	return c.client.ZAdd(ctx, key, members...).Err()
}

func (c *Client) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return c.client.ZRem(ctx, key, members...).Err()
}

func (c *Client) ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return c.client.ZRange(ctx, key, start, stop)
}

func (c *Client) ZRangeWithScores(ctx context.Context, key string, start, stop int64) *redis.ZSliceCmd {
	return c.client.ZRangeWithScores(ctx, key, start, stop)
}

func (c *Client) ZRevRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return c.client.ZRevRange(ctx, key, start, stop)
}

func (c *Client) ZScore(ctx context.Context, key, member string) *redis.FloatCmd {
	return c.client.ZScore(ctx, key, member)
}

// 过期时间操作
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

func (c *Client) TTL(ctx context.Context, key string) *redis.DurationCmd {
	return c.client.TTL(ctx, key)
}

// 批量操作
func (c *Client) Pipeline() redis.Pipeliner {
	return c.client.Pipeline()
}

func (c *Client) TxPipeline() redis.Pipeliner {
	return c.client.TxPipeline()
}

// 发布订阅
func (c *Client) Publish(ctx context.Context, channel string, message interface{}) error {
	return c.client.Publish(ctx, channel, message).Err()
}

func (c *Client) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return c.client.Subscribe(ctx, channels...)
}

// 缓存辅助方法
func (c *Client) CacheGet(ctx context.Context, key string, dest interface{}) (bool, error) {
	data, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return false, err
	}

	return true, nil
}

func (c *Client) CacheSet(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

// 分布式锁
func (c *Client) Lock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	result, err := c.client.SetNX(ctx, key, "locked", expiration).Result()
	return result, err
}

func (c *Client) Unlock(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// 限流器
func (c *Client) RateLimit(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	pipe := c.client.TxPipeline()
	
	// 增加计数
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	
	results, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	count := results[0].(*redis.IntCmd).Val()
	return count <= limit, nil
}

// 健康检查
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// 关闭连接
func (c *Client) Close() error {
	return c.client.Close()
}
