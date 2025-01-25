package cache

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrKeyNotFound 键不存在
	ErrKeyNotFound = errors.New("key not found")

	// ErrKeyExpired 键已过期
	ErrKeyExpired = errors.New("key expired")
)

// Item 缓存项
type Item struct {
	// Key 缓存键
	Key string

	// Value 缓存值
	Value interface{}

	// ExpireAt 过期时间
	ExpireAt time.Time

	// CreatedAt 创建时间
	CreatedAt time.Time
}

// Options 缓存选项
type Options struct {
	// TTL 过期时间
	TTL time.Duration

	// MaxEntries 最大条目数
	MaxEntries int

	// OnEvicted 条目被移除时的回调函数
	OnEvicted func(key string, value interface{})
}

// Cache 缓存接口
type Cache interface {
	// Get 获取缓存值
	Get(ctx context.Context, key string) (interface{}, error)

	// Set 设置缓存值
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete 删除缓存值
	Delete(ctx context.Context, key string) error

	// Clear 清空缓存
	Clear(ctx context.Context) error

	// Keys 获取所有键
	Keys(ctx context.Context) ([]string, error)

	// Len 获取缓存项数量
	Len(ctx context.Context) int

	// Close 关闭缓存
	Close() error
}

// Option 配置选项函数
type Option func(*Options)

// WithTTL 设置默认过期时间
func WithTTL(ttl time.Duration) Option {
	return func(o *Options) {
		o.TTL = ttl
	}
}

// WithMaxEntries 设置最大条目数
func WithMaxEntries(max int) Option {
	return func(o *Options) {
		o.MaxEntries = max
	}
}

// WithOnEvicted 设置移除回调函数
func WithOnEvicted(fn func(key string, value interface{})) Option {
	return func(o *Options) {
		o.OnEvicted = fn
	}
}
