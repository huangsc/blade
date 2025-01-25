package circuitbreaker

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrCircuitOpen 表示熔断器已打开
	ErrCircuitOpen = errors.New("circuit breaker is open")

	// ErrTooManyRequests 表示请求过多
	ErrTooManyRequests = errors.New("too many requests")
)

// State 熔断器状态
type State int

const (
	// StateClosed 关闭状态(正常工作)
	StateClosed State = iota

	// StateHalfOpen 半开状态(尝试恢复)
	StateHalfOpen

	// StateOpen 打开状态(熔断服务)
	StateOpen
)

// Counts 计数器
type Counts struct {
	// Requests 请求总数
	Requests uint64

	// TotalSuccesses 总成功数
	TotalSuccesses uint64

	// TotalFailures 总失败数
	TotalFailures uint64

	// ConsecutiveSuccesses 连续成功数
	ConsecutiveSuccesses uint64

	// ConsecutiveFailures 连续失败数
	ConsecutiveFailures uint64
}

// Generation 表示熔断器的一代
type Generation struct {
	// Counts 计数器
	Counts Counts

	// ExpiredAt 过期时间
	ExpiredAt time.Time
}

// Settings 熔断器配置
type Settings struct {
	// Name 熔断器名称
	Name string

	// MaxRequests 最大请求数
	MaxRequests uint64

	// Interval 统计周期
	Interval time.Duration

	// Timeout 超时时间
	Timeout time.Duration

	// ReadyToTrip 确定是否触发熔断的函数
	ReadyToTrip func(counts Counts) bool

	// OnStateChange 状态变更回调函数
	OnStateChange func(name string, from State, to State)
}

// Breaker 熔断器接口
type Breaker interface {
	// Name 返回熔断器名称
	Name() string

	// State 返回当前状态
	State() State

	// Counts 返回当前计数
	Counts() Counts

	// Execute 执行受保护的函数
	Execute(ctx context.Context, fn func() error) error

	// Allow 判断是否允许请求通过
	Allow() (func(success bool), error)

	// Success 报告成功
	Success()

	// Failure 报告失败
	Failure()
}

// TwoStepBreaker 支持两阶段执行的熔断器接口
type TwoStepBreaker interface {
	Breaker

	// Allow 判断是否允许请求通过
	Allow() (func(success bool), error)
}

// Option 配置选项函数
type Option func(*Settings)

// WithMaxRequests 设置最大请求数
func WithMaxRequests(n uint64) Option {
	return func(s *Settings) {
		s.MaxRequests = n
	}
}

// WithInterval 设置统计周期
func WithInterval(d time.Duration) Option {
	return func(s *Settings) {
		s.Interval = d
	}
}

// WithTimeout 设置超时时间
func WithTimeout(d time.Duration) Option {
	return func(s *Settings) {
		s.Timeout = d
	}
}

// WithReadyToTrip 设置触发熔断的判断函数
func WithReadyToTrip(fn func(counts Counts) bool) Option {
	return func(s *Settings) {
		s.ReadyToTrip = fn
	}
}

// WithOnStateChange 设置状态变更回调函数
func WithOnStateChange(fn func(name string, from State, to State)) Option {
	return func(s *Settings) {
		s.OnStateChange = fn
	}
}

// defaultReadyToTrip 默认的触发熔断判断函数
func defaultReadyToTrip(counts Counts) bool {
	return counts.ConsecutiveFailures > 5
}

// defaultOnStateChange 默认的状态变更回调函数
func defaultOnStateChange(name string, from State, to State) {}
