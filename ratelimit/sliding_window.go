package ratelimit

import (
	"context"
	"sync"
	"time"
)

// window 时间窗口
type window struct {
	// timestamp 窗口的起始时间戳(秒)
	timestamp int64

	// count 当前窗口的请求数
	count int64
}

// slidingWindow 基于内存的滑动窗口限流器
type slidingWindow struct {
	// mu 互斥锁
	mu sync.Mutex

	// rate 请求速率(每秒)
	rate float64

	// window 窗口大小(秒)
	window time.Duration

	// precision 窗口精度(秒)
	precision time.Duration

	// windows 窗口映射表 key -> []window
	windows map[string][]window
}

// NewSlidingWindow 创建基于内存的滑动窗口限流器
func NewSlidingWindow(opts ...Option) Limiter {
	options := &Options{
		Rate:      1,
		Burst:     1,
		Window:    time.Second,
		Precision: time.Millisecond * 100,
	}
	for _, opt := range opts {
		opt(options)
	}

	return &slidingWindow{
		rate:      options.Rate,
		window:    options.Window,
		precision: options.Precision,
		windows:   make(map[string][]window),
	}
}

// Allow 判断是否允许请求通过
func (l *slidingWindow) Allow(ctx context.Context, key string, n int64) (bool, error) {
	return l.AllowN(ctx, key, n, time.Now())
}

// AllowN 判断在指定时间内是否允许请求通过
func (l *slidingWindow) AllowN(ctx context.Context, key string, n int64, now time.Time) (bool, error) {
	if n <= 0 {
		return true, nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// 获取当前时间戳
	timestamp := now.Unix()

	// 获取或创建窗口列表
	windows, ok := l.windows[key]
	if !ok {
		windows = make([]window, 0)
		l.windows[key] = windows
	}

	// 清理过期的窗口
	windowSize := int64(l.window.Seconds())
	expireTime := timestamp - windowSize
	validWindows := make([]window, 0)
	for _, w := range windows {
		if w.timestamp > expireTime {
			validWindows = append(validWindows, w)
		}
	}
	l.windows[key] = validWindows

	// 计算当前请求数
	var count int64
	for _, w := range validWindows {
		count += w.count
	}

	// 判断是否超过限制
	maxRequests := int64(l.rate * l.window.Seconds())
	if count+n > maxRequests {
		return false, nil
	}

	// 更新当前窗口
	precisionSeconds := int64(l.precision.Seconds())
	currentTimestamp := timestamp - (timestamp % precisionSeconds)

	found := false
	for i := range validWindows {
		if validWindows[i].timestamp == currentTimestamp {
			validWindows[i].count += n
			found = true
			break
		}
	}

	if !found {
		validWindows = append(validWindows, window{
			timestamp: currentTimestamp,
			count:     n,
		})
		l.windows[key] = validWindows
	}

	return true, nil
}

// Wait 等待直到获取到足够的配额
func (l *slidingWindow) Wait(ctx context.Context, key string, n int64) error {
	return l.WaitN(ctx, key, n, time.Now())
}

// WaitN 等待直到指定时间获取到足够的配额
func (l *slidingWindow) WaitN(ctx context.Context, key string, n int64, now time.Time) error {
	// 先尝试直接获取配额
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

	// 等待配额
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

// Reserve 预约未来的配额
func (l *slidingWindow) Reserve(ctx context.Context, key string, n int64) (*Reservation, error) {
	return l.ReserveN(ctx, key, n, time.Now())
}

// ReserveN 在指定时间预约未来的配额
func (l *slidingWindow) ReserveN(ctx context.Context, key string, n int64, now time.Time) (*Reservation, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 获取当前时间戳
	timestamp := now.Unix()

	// 获取或创建窗口列表
	windows, ok := l.windows[key]
	if !ok {
		windows = make([]window, 0)
		l.windows[key] = windows
	}

	// 清理过期的窗口
	windowSize := int64(l.window.Seconds())
	expireTime := timestamp - windowSize
	validWindows := make([]window, 0)
	for _, w := range windows {
		if w.timestamp > expireTime {
			validWindows = append(validWindows, w)
		}
	}
	l.windows[key] = validWindows

	// 计算当前请求数
	var count int64
	for _, w := range validWindows {
		count += w.count
	}

	// 计算等待时间
	maxRequests := int64(l.rate * l.window.Seconds())
	var waitDuration time.Duration
	if count+n > maxRequests {
		// 计算需要等待的时间
		waitSeconds := float64(count+n-maxRequests) / l.rate
		waitDuration = time.Duration(waitSeconds * float64(time.Second))
	}

	// 创建预约信息
	r := &Reservation{
		OK:        true,
		Limit:     maxRequests,
		Tokens:    n,
		TimeToAct: now.Add(waitDuration),
		DelayFrom: now,
	}

	return r, nil
}
