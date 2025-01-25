package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisOptions Redis配置选项
type RedisOptions struct {
	// 基础配置选项
	Options

	// Addr Redis地址
	Addr string

	// Password Redis密码
	Password string

	// DB Redis数据库
	DB int

	// PoolSize 连接池大小
	PoolSize int

	// MinIdleConns 最小空闲连接数
	MinIdleConns int

	// KeyPrefix 键前缀
	KeyPrefix string
}

// redis Redis缓存实现
type redisCache struct {
	// client Redis客户端
	client *redis.Client

	// options 配置选项
	options RedisOptions
}

// NewRedisCache 创建Redis缓存
func NewRedisCache(opts RedisOptions) Cache {
	client := redis.NewClient(&redis.Options{
		Addr:         opts.Addr,
		Password:     opts.Password,
		DB:           opts.DB,
		PoolSize:     opts.PoolSize,
		MinIdleConns: opts.MinIdleConns,
	})

	return &redisCache{
		client:  client,
		options: opts,
	}
}

// Get 获取缓存值
func (r *redisCache) Get(ctx context.Context, key string) (interface{}, error) {
	key = r.options.KeyPrefix + key
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, ErrKeyNotFound
	}
	if err != nil {
		return nil, err
	}

	var item Item
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, err
	}

	if item.ExpireAt.Before(time.Now()) {
		r.Delete(ctx, key)
		return nil, ErrKeyExpired
	}

	return item.Value, nil
}

// Set 设置缓存值
func (r *redisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = r.options.TTL
	}

	item := Item{
		Key:       key,
		Value:     value,
		ExpireAt:  time.Now().Add(ttl),
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	key = r.options.KeyPrefix + key
	return r.client.Set(ctx, key, data, ttl).Err()
}

// Delete 删除缓存值
func (r *redisCache) Delete(ctx context.Context, key string) error {
	if r.options.OnEvicted != nil {
		// 获取旧值
		if value, err := r.Get(ctx, key); err == nil {
			r.options.OnEvicted(key, value)
		}
	}

	key = r.options.KeyPrefix + key
	return r.client.Del(ctx, key).Err()
}

// Clear 清空缓存
func (r *redisCache) Clear(ctx context.Context) error {
	if r.options.OnEvicted != nil {
		// 获取所有键
		keys, err := r.Keys(ctx)
		if err != nil {
			return err
		}

		// 触发回调
		for _, key := range keys {
			if value, err := r.Get(ctx, key); err == nil {
				r.options.OnEvicted(key, value)
			}
		}
	}

	pattern := r.options.KeyPrefix + "*"
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return r.client.Del(ctx, keys...).Err()
	}

	return nil
}

// Keys 获取所有键
func (r *redisCache) Keys(ctx context.Context) ([]string, error) {
	pattern := r.options.KeyPrefix + "*"
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	// 移除前缀
	prefixLen := len(r.options.KeyPrefix)
	result := make([]string, len(keys))
	for i, key := range keys {
		result[i] = key[prefixLen:]
	}

	return result, nil
}

// Len 获取缓存项数量
func (r *redisCache) Len(ctx context.Context) int {
	pattern := r.options.KeyPrefix + "*"
	count, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return 0
	}
	return len(count)
}

// Close 关闭缓存
func (r *redisCache) Close() error {
	return r.client.Close()
}
