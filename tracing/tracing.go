package tracing

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// SpanKind 表示 Span 的类型
type SpanKind int

const (
	// SpanKindUnspecified 未指定类型
	SpanKindUnspecified SpanKind = iota
	// SpanKindInternal 内部调用
	SpanKindInternal
	// SpanKindServer 服务端
	SpanKindServer
	// SpanKindClient 客户端
	SpanKindClient
	// SpanKindProducer 生产者
	SpanKindProducer
	// SpanKindConsumer 消费者
	SpanKindConsumer
)

// Tracer 追踪器接口
type Tracer interface {
	// Start 开始一个新的 Span
	Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)
	// Extract 从载体中提取上下文
	Extract(ctx context.Context, carrier interface{}) context.Context
	// Inject 注入上下文到载体
	Inject(ctx context.Context, carrier interface{}) error
}

// Span 表示一个追踪片段
type Span interface {
	// End 结束 Span
	End()
	// SetName 设置名称
	SetName(name string)
	// SetStatus 设置状态
	SetStatus(code int, description string)
	// SetError 设置错误
	SetError(err error)
	// SetAttributes 设置属性
	SetAttributes(kv ...attribute.KeyValue)
	// AddEvent 添加事件
	AddEvent(name string, attributes ...attribute.KeyValue)
	// RecordError 记录错误
	RecordError(err error, opts ...trace.EventOption)
	// SpanContext 获取 Span 上下文
	SpanContext() trace.SpanContext
	// IsRecording 是否正在记录
	IsRecording() bool
	// TracerProvider 获取追踪器提供者
	TracerProvider() trace.TracerProvider
}

// SpanOption 配置 Span 的选项
type SpanOption func(*SpanOptions)

// SpanOptions Span 的配置选项
type SpanOptions struct {
	// Attributes Span 的属性
	Attributes []attribute.KeyValue
	// StartTime 开始时间
	StartTime time.Time
	// Kind Span 的类型
	Kind SpanKind
}

// WithSpanAttributes 设置 Span 属性
func WithSpanAttributes(kv ...attribute.KeyValue) SpanOption {
	return func(o *SpanOptions) {
		o.Attributes = append(o.Attributes, kv...)
	}
}

// WithStartTime 设置开始时间
func WithStartTime(t time.Time) SpanOption {
	return func(o *SpanOptions) {
		o.StartTime = t
	}
}

// WithSpanKind 设置 Span 类型
func WithSpanKind(kind SpanKind) SpanOption {
	return func(o *SpanOptions) {
		o.Kind = kind
	}
}

// Option 配置选项
type Option func(*Options)

// Options 配置选项
type Options struct {
	// ServiceName 服务名称
	ServiceName string
	// ServiceVersion 服务版本
	ServiceVersion string
	// Environment 环境
	Environment string
	// Endpoint 端点
	Endpoint string
	// Sampler 采样器
	Sampler float64
	// Attributes 全局属性
	Attributes []attribute.KeyValue
	// Timeout 超时时间
	Timeout time.Duration
	// RetryCount 重试次数
	RetryCount int
	// RetryDelay 重试延迟
	RetryDelay time.Duration
}

// WithServiceName 设置服务名称
func WithServiceName(name string) Option {
	return func(o *Options) {
		o.ServiceName = name
	}
}

// WithServiceVersion 设置服务版本
func WithServiceVersion(version string) Option {
	return func(o *Options) {
		o.ServiceVersion = version
	}
}

// WithEnvironment 设置环境
func WithEnvironment(env string) Option {
	return func(o *Options) {
		o.Environment = env
	}
}

// WithEndpoint 设置端点
func WithEndpoint(endpoint string) Option {
	return func(o *Options) {
		o.Endpoint = endpoint
	}
}

// WithSampler 设置采样器
func WithSampler(sampler float64) Option {
	return func(o *Options) {
		o.Sampler = sampler
	}
}

// WithGlobalAttributes 设置全局属性
func WithGlobalAttributes(kv ...attribute.KeyValue) Option {
	return func(o *Options) {
		o.Attributes = append(o.Attributes, kv...)
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

// WithRetryCount 设置重试次数
func WithRetryCount(count int) Option {
	return func(o *Options) {
		o.RetryCount = count
	}
}

// WithRetryDelay 设置重试延迟
func WithRetryDelay(delay time.Duration) Option {
	return func(o *Options) {
		o.RetryDelay = delay
	}
}
