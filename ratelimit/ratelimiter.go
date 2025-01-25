package ratelimit

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrLimitExceeded 表示超过限流阈值
	ErrLimitExceeded = errors.New("rate limit exceeded")
)

// Limiter 限流器接口
type Limiter interface {
	// Allow 判断是否允许请求通过
	// key: 限流的键,用于区分不同的限流目标
	// n: 本次请求消耗的令牌数
	Allow(ctx context.Context, key string, n int64) (bool, error)

	// AllowN 判断在指定时间内是否允许请求通过
	// key: 限流的键
	// n: 本次请求消耗的令牌数
	// now: 当前时间
	AllowN(ctx context.Context, key string, n int64, now time.Time) (bool, error)

	// Wait 等待直到获取到足够的令牌
	// key: 限流的键
	// n: 需要的令牌数
	Wait(ctx context.Context, key string, n int64) error

	// WaitN 等待直到指定时间获取到足够的令牌
	// key: 限流的键
	// n: 需要的令牌数
	// now: 当前时间
	WaitN(ctx context.Context, key string, n int64, now time.Time) error

	// Reserve 预约未来的令牌
	// key: 限流的键
	// n: 需要预约的令牌数
	Reserve(ctx context.Context, key string, n int64) (*Reservation, error)

	// ReserveN 在指定时间预约未来的令牌
	// key: 限流的键
	// n: 需要预约的令牌数
	// now: 当前时间
	ReserveN(ctx context.Context, key string, n int64, now time.Time) (*Reservation, error)
}

// Reservation 令牌预约信息
type Reservation struct {
	// OK 表示是否预约成功
	OK bool

	// Limit 令牌桶大小
	Limit int64

	// Tokens 当前可用的令牌数
	Tokens int64

	// TimeToAct 可以获取令牌的时间
	TimeToAct time.Time

	// DelayFrom 从指定时间开始的延迟
	DelayFrom time.Time
}

// Cancel 取消预约
func (r *Reservation) Cancel() {
	r.OK = false
}

// Delay 获取从指定时间到可以获取令牌的延迟时间
func (r *Reservation) Delay() time.Duration {
	return r.DelayFrom.Sub(r.TimeToAct)
}

// Options 限流器配置选项
type Options struct {
	// Rate 令牌产生速率(每秒)
	Rate float64

	// Burst 令牌桶大小
	Burst int64

	// TTL 键的过期时间
	TTL time.Duration

	// Window 窗口大小(秒)
	Window time.Duration

	// Precision 窗口精度(秒)
	Precision time.Duration

	// Limiter 限流器实例
	Limiter Limiter
}

// Option 配置选项函数
type Option func(*Options)

// WithRate 设置令牌产生速率
func WithRate(rate float64) Option {
	return func(o *Options) {
		o.Rate = rate
	}
}

// WithBurst 设置令牌桶大小
func WithBurst(burst int64) Option {
	return func(o *Options) {
		o.Burst = burst
	}
}

// WithTTL 设置键的过期时间
func WithTTL(ttl time.Duration) Option {
	return func(o *Options) {
		o.TTL = ttl
	}
}

// WithWindow 设置窗口大小
func WithWindow(window time.Duration) Option {
	return func(o *Options) {
		o.Window = window
	}
}

// WithPrecision 设置窗口精度
func WithPrecision(precision time.Duration) Option {
	return func(o *Options) {
		o.Precision = precision
	}
}
