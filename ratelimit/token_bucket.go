package ratelimit

import (
	"context"
	"math"
	"sync"
	"time"
)

// bucket 令牌桶
type bucket struct {
	// last 上次更新时间
	last time.Time

	// tokens 当前令牌数
	tokens float64

	// rate 令牌产生速率(每秒)
	rate float64

	// burst 令牌桶大小
	burst float64
}

// tokenBucket 基于内存的令牌桶限流器
type tokenBucket struct {
	// mu 互斥锁
	mu sync.Mutex

	// rate 令牌产生速率(每秒)
	rate float64

	// burst 令牌桶大小
	burst int64

	// buckets 令牌桶映射表
	buckets map[string]*bucket
}

// NewTokenBucket 创建基于内存的令牌桶限流器
func NewTokenBucket(opts ...Option) Limiter {
	options := &Options{
		Rate:  1,
		Burst: 1,
	}
	for _, opt := range opts {
		opt(options)
	}

	return &tokenBucket{
		rate:    options.Rate,
		burst:   options.Burst,
		buckets: make(map[string]*bucket),
	}
}

// Allow 判断是否允许请求通过
func (l *tokenBucket) Allow(ctx context.Context, key string, n int64) (bool, error) {
	return l.AllowN(ctx, key, n, time.Now())
}

// AllowN 判断在指定时间内是否允许请求通过
func (l *tokenBucket) AllowN(ctx context.Context, key string, n int64, now time.Time) (bool, error) {
	if n <= 0 {
		return true, nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// 获取或创建令牌桶
	bkt, ok := l.buckets[key]
	if !ok {
		bkt = &bucket{
			last:   now,
			tokens: float64(l.burst),
			rate:   l.rate,
			burst:  float64(l.burst),
		}
		l.buckets[key] = bkt
	}

	// 计算新增的令牌数
	elapsed := now.Sub(bkt.last).Seconds()
	delta := elapsed * bkt.rate

	// 更新令牌数
	bkt.tokens = math.Min(bkt.burst, bkt.tokens+delta)
	bkt.last = now

	// 判断令牌是否足够
	if bkt.tokens < float64(n) {
		return false, nil
	}

	// 消费令牌
	bkt.tokens -= float64(n)
	return true, nil
}

// Wait 等待直到获取到足够的令牌
func (l *tokenBucket) Wait(ctx context.Context, key string, n int64) error {
	return l.WaitN(ctx, key, n, time.Now())
}

// WaitN 等待直到指定时间获取到足够的令牌
func (l *tokenBucket) WaitN(ctx context.Context, key string, n int64, now time.Time) error {
	// 先尝试直接获取令牌
	if ok, _ := l.AllowN(ctx, key, n, now); ok {
		return nil
	}

	// 计算需要等待的时间
	reservation, err := l.ReserveN(ctx, key, n, now)
	if err != nil {
		return err
	}
	if !reservation.OK {
		return ErrLimitExceeded
	}

	// 等待令牌
	delay := reservation.DelayFrom.Sub(now)
	if delay <= 0 {
		return nil
	}

	select {
	case <-ctx.Done():
		// 取消预约
		reservation.Cancel()
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}

// Reserve 预约未来的令牌
func (l *tokenBucket) Reserve(ctx context.Context, key string, n int64) (*Reservation, error) {
	return l.ReserveN(ctx, key, n, time.Now())
}

// ReserveN 在指定时间预约未来的令牌
func (l *tokenBucket) ReserveN(ctx context.Context, key string, n int64, now time.Time) (*Reservation, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 获取或创建令牌桶
	bkt, ok := l.buckets[key]
	if !ok {
		bkt = &bucket{
			last:   now,
			tokens: float64(l.burst),
			rate:   l.rate,
			burst:  float64(l.burst),
		}
		l.buckets[key] = bkt
	}

	// 计算新增的令牌数
	elapsed := now.Sub(bkt.last).Seconds()
	delta := elapsed * bkt.rate

	// 更新令牌数
	bkt.tokens = math.Min(bkt.burst, bkt.tokens+delta)
	bkt.last = now

	// 计算等待时间
	tokens := bkt.tokens - float64(n)
	var waitDuration time.Duration
	if tokens < 0 {
		waitDuration = time.Duration((-tokens/bkt.rate)*1e9) * time.Nanosecond
	}

	// 创建预约信息
	r := &Reservation{
		OK:        true,
		Limit:     l.burst,
		Tokens:    int64(bkt.tokens),
		TimeToAct: now.Add(waitDuration),
		DelayFrom: now,
	}

	// 消费令牌
	bkt.tokens = tokens
	return r, nil
}
