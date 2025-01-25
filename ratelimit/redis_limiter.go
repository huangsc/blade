package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const luaScript = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local requested = tonumber(ARGV[2])
local rate = tonumber(ARGV[3])
local burst = tonumber(ARGV[4])
local ttl = tonumber(ARGV[5])

-- 获取当前桶的信息
local bucket = redis.call('hmget', key, 'last', 'tokens')
local last = tonumber(bucket[1]) or now
local tokens = tonumber(bucket[2]) or burst

-- 计算新增的令牌数
local elapsed = math.max(0, now - last)
local delta = elapsed * rate
tokens = math.min(burst, tokens + delta)

-- 判断令牌是否足够
if tokens < requested then
    -- 令牌不足,返回需要等待的时间
    local wait = (requested - tokens) / rate
    return {0, wait}
end

-- 更新令牌数
tokens = tokens - requested
redis.call('hmset', key, 'last', now, 'tokens', tokens)
redis.call('expire', key, ttl)

return {1, 0}
`

// redisLimiter 基于 Redis 的分布式限流器
type redisLimiter struct {
	// rdb Redis 客户端
	rdb redis.UniversalClient

	// rate 令牌产生速率(每秒)
	rate float64

	// burst 令牌桶大小
	burst int64

	// ttl 键的过期时间
	ttl time.Duration
}

// NewRedisLimiter 创建基于 Redis 的分布式限流器
func NewRedisLimiter(rdb redis.UniversalClient, opts ...Option) Limiter {
	options := &Options{
		Rate:  1,
		Burst: 1,
		TTL:   time.Hour,
	}
	for _, opt := range opts {
		opt(options)
	}

	return &redisLimiter{
		rdb:   rdb,
		rate:  options.Rate,
		burst: options.Burst,
		ttl:   options.TTL,
	}
}

// Allow 判断是否允许请求通过
func (l *redisLimiter) Allow(ctx context.Context, key string, n int64) (bool, error) {
	return l.AllowN(ctx, key, n, time.Now())
}

// AllowN 判断在指定时间内是否允许请求通过
func (l *redisLimiter) AllowN(ctx context.Context, key string, n int64, now time.Time) (bool, error) {
	if n <= 0 {
		return true, nil
	}

	// 执行 Lua 脚本
	results, err := l.rdb.Eval(ctx, luaScript, []string{key}, []string{
		strconv.FormatInt(now.Unix(), 10),
		strconv.FormatInt(n, 10),
		fmt.Sprintf("%.2f", l.rate),
		strconv.FormatInt(l.burst, 10),
		strconv.FormatInt(int64(l.ttl.Seconds()), 10),
	}).Result()
	if err != nil {
		return false, err
	}

	// 解析结果
	res, ok := results.([]interface{})
	if !ok || len(res) != 2 {
		return false, fmt.Errorf("invalid redis response: %v", results)
	}

	// 获取是否允许通过
	allowed, ok := res[0].(int64)
	if !ok {
		return false, fmt.Errorf("invalid allowed value: %v", res[0])
	}

	return allowed == 1, nil
}

// Wait 等待直到获取到足够的令牌
func (l *redisLimiter) Wait(ctx context.Context, key string, n int64) error {
	return l.WaitN(ctx, key, n, time.Now())
}

// WaitN 等待直到指定时间获取到足够的令牌
func (l *redisLimiter) WaitN(ctx context.Context, key string, n int64, now time.Time) error {
	// 先尝试直接获取令牌
	if ok, err := l.AllowN(ctx, key, n, now); err != nil {
		return err
	} else if ok {
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
func (l *redisLimiter) Reserve(ctx context.Context, key string, n int64) (*Reservation, error) {
	return l.ReserveN(ctx, key, n, time.Now())
}

// ReserveN 在指定时间预约未来的令牌
func (l *redisLimiter) ReserveN(ctx context.Context, key string, n int64, now time.Time) (*Reservation, error) {
	if n <= 0 {
		return &Reservation{
			OK:        true,
			Limit:     l.burst,
			Tokens:    0,
			TimeToAct: now,
			DelayFrom: now,
		}, nil
	}

	// 执行 Lua 脚本
	results, err := l.rdb.Eval(ctx, luaScript, []string{key}, []string{
		strconv.FormatInt(now.Unix(), 10),
		strconv.FormatInt(n, 10),
		fmt.Sprintf("%.2f", l.rate),
		strconv.FormatInt(l.burst, 10),
		strconv.FormatInt(int64(l.ttl.Seconds()), 10),
	}).Result()
	if err != nil {
		return nil, err
	}

	// 解析结果
	res, ok := results.([]interface{})
	if !ok || len(res) != 2 {
		return nil, fmt.Errorf("invalid redis response: %v", results)
	}

	// 获取是否允许通过和等待时间
	allowed, ok := res[0].(int64)
	if !ok {
		return nil, fmt.Errorf("invalid allowed value: %v", res[0])
	}

	wait, ok := res[1].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid wait value: %v", res[1])
	}

	// 创建预约信息
	waitDuration := time.Duration(wait * float64(time.Second))
	r := &Reservation{
		OK:        allowed == 1,
		Limit:     l.burst,
		Tokens:    n,
		TimeToAct: now.Add(waitDuration),
		DelayFrom: now,
	}

	return r, nil
}
