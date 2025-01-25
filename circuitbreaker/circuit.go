package circuitbreaker

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// circuit 基本的熔断器实现
type circuit struct {
	// name 熔断器名称
	name string

	// maxRequests 最大请求数
	maxRequests uint64

	// interval 统计周期
	interval time.Duration

	// timeout 超时时间
	timeout time.Duration

	// readyToTrip 确定是否触发熔断的函数
	readyToTrip func(counts Counts) bool

	// onStateChange 状态变更回调函数
	onStateChange func(name string, from State, to State)

	// mutex 互斥锁
	mutex sync.Mutex

	// state 当前状态
	state State

	// generation 当前代
	generation *Generation

	// counts 当前计数
	counts Counts
}

// NewBreaker 创建新的熔断器
func NewBreaker(name string, opts ...Option) Breaker {
	settings := &Settings{
		Name:          name,
		MaxRequests:   1,
		Interval:      time.Minute,
		Timeout:       time.Second * 60,
		ReadyToTrip:   defaultReadyToTrip,
		OnStateChange: defaultOnStateChange,
	}

	for _, opt := range opts {
		opt(settings)
	}

	return &circuit{
		name:          settings.Name,
		maxRequests:   settings.MaxRequests,
		interval:      settings.Interval,
		timeout:       settings.Timeout,
		readyToTrip:   settings.ReadyToTrip,
		onStateChange: settings.OnStateChange,
		state:         StateClosed,
		generation:    newGeneration(),
	}
}

// newGeneration 创建新的一代
func newGeneration() *Generation {
	return &Generation{
		ExpiredAt: time.Now().Add(time.Minute),
	}
}

// Name 返回熔断器名称
func (c *circuit) Name() string {
	return c.name
}

// State 返回当前状态
func (c *circuit) State() State {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.state
}

// Counts 返回当前计数
func (c *circuit) Counts() Counts {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.counts
}

// Execute 执行受保护的函数
func (c *circuit) Execute(ctx context.Context, fn func() error) error {
	done, err := c.Allow()
	if err != nil {
		return err
	}

	err = fn()
	done(err == nil)
	return err
}

// Allow 判断是否允许请求通过
func (c *circuit) Allow() (func(success bool), error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	state := c.state

	// 检查是否需要重置计数
	if c.generation.ExpiredAt.Before(now) {
		c.toNewGeneration(now)
	}

	// 根据状态判断是否允许请求
	switch state {
	case StateClosed:
		// 关闭状态,允许请求
		atomic.AddUint64(&c.counts.Requests, 1)
		return c.done, nil

	case StateOpen:
		// 开启状态,检查是否超时
		if c.generation.ExpiredAt.Before(now) {
			// 超时后进入半开状态
			c.setState(StateHalfOpen, now)
			atomic.AddUint64(&c.counts.Requests, 1)
			return c.done, nil
		}
		return nil, ErrCircuitOpen

	default: // StateHalfOpen
		// 半开状态,检查是否超过最大请求数
		if atomic.LoadUint64(&c.counts.Requests) >= c.maxRequests {
			return nil, ErrTooManyRequests
		}
		atomic.AddUint64(&c.counts.Requests, 1)
		return c.done, nil
	}
}

// Success 报告成功
func (c *circuit) Success() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	atomic.AddUint64(&c.counts.TotalSuccesses, 1)
	atomic.AddUint64(&c.counts.ConsecutiveSuccesses, 1)
	atomic.StoreUint64(&c.counts.ConsecutiveFailures, 0)

	// 在半开状态下,如果连续成功次数达到阈值,则关闭熔断器
	if c.state == StateHalfOpen && atomic.LoadUint64(&c.counts.ConsecutiveSuccesses) >= c.maxRequests {
		c.setState(StateClosed, time.Now())
	}
}

// Failure 报告失败
func (c *circuit) Failure() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	atomic.AddUint64(&c.counts.TotalFailures, 1)
	atomic.AddUint64(&c.counts.ConsecutiveFailures, 1)
	atomic.StoreUint64(&c.counts.ConsecutiveSuccesses, 0)

	// 检查是否需要触发熔断
	if c.readyToTrip(c.counts) {
		c.setState(StateOpen, time.Now())
	}
}

// done 请求完成的回调函数
func (c *circuit) done(success bool) {
	if success {
		c.Success()
	} else {
		c.Failure()
	}
}

// setState 设置熔断器状态
func (c *circuit) setState(state State, now time.Time) {
	if c.state == state {
		return
	}

	prev := c.state
	c.state = state

	// 重置计数
	c.counts = Counts{}

	// 设置过期时间
	var expiry time.Time
	switch state {
	case StateClosed:
		expiry = now.Add(c.interval)
	case StateOpen:
		expiry = now.Add(c.timeout)
	default: // StateHalfOpen
		expiry = now.Add(c.interval)
	}
	c.generation = &Generation{
		ExpiredAt: expiry,
	}

	// 触发状态变更回调
	if c.onStateChange != nil {
		c.onStateChange(c.name, prev, state)
	}
}

// toNewGeneration 创建新的一代
func (c *circuit) toNewGeneration(now time.Time) {
	c.counts = Counts{}
	c.generation = &Generation{
		ExpiredAt: now.Add(c.interval),
	}
}
