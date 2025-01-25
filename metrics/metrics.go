package metrics

import (
	"time"
)

// MetricType 指标类型
type MetricType int

const (
	// CounterType 计数器类型
	CounterType MetricType = iota
	// GaugeType 仪表盘类型
	GaugeType
	// HistogramType 直方图类型
	HistogramType
	// SummaryType 摘要类型
	SummaryType
)

// Labels 标签集合
type Labels map[string]string

// Metrics 指标接口
type Metrics interface {
	// Counter 计数器
	Counter(name string, labels Labels) CounterMetric
	// Gauge 仪表盘
	Gauge(name string, labels Labels) GaugeMetric
	// Histogram 直方图
	Histogram(name string, labels Labels) HistogramMetric
	// Summary 摘要
	Summary(name string, labels Labels) SummaryMetric
}

// CounterMetric 计数器接口
type CounterMetric interface {
	// Inc 增加计数，增量为1
	Inc()
	// Add 增加计数，增量为指定值
	Add(value float64)
	// WithLabels 设置标签
	WithLabels(labels Labels) CounterMetric
}

// GaugeMetric 仪表盘接口
type GaugeMetric interface {
	// Set 设置值
	Set(value float64)
	// Inc 增加值，增量为1
	Inc()
	// Dec 减少值，减量为1
	Dec()
	// Add 增加值，增量为指定值
	Add(value float64)
	// Sub 减少值，减量为指定值
	Sub(value float64)
	// WithLabels 设置标签
	WithLabels(labels Labels) GaugeMetric
}

// HistogramMetric 直方图接口
type HistogramMetric interface {
	// Observe 观察值
	Observe(value float64)
	// WithLabels 设置标签
	WithLabels(labels Labels) HistogramMetric
}

// SummaryMetric 摘要接口
type SummaryMetric interface {
	// Observe 观察值
	Observe(value float64)
	// WithLabels 设置标签
	WithLabels(labels Labels) SummaryMetric
}

// Option 配置选项
type Option func(*Options)

// Options 配置选项
type Options struct {
	// Namespace 命名空间
	Namespace string
	// Subsystem 子系统
	Subsystem string
	// ConstLabels 固定标签
	ConstLabels Labels
	// Buckets 直方图桶
	Buckets []float64
	// Objectives 摘要目标
	Objectives map[float64]float64
	// MaxAge 最大存活时间
	MaxAge time.Duration
	// AgeBuckets 存活时间桶数量
	AgeBuckets uint32
	// BufCap 缓冲区容量
	BufCap uint32
}

// WithNamespace 设置命名空间
func WithNamespace(namespace string) Option {
	return func(o *Options) {
		o.Namespace = namespace
	}
}

// WithSubsystem 设置子系统
func WithSubsystem(subsystem string) Option {
	return func(o *Options) {
		o.Subsystem = subsystem
	}
}

// WithConstLabels 设置固定标签
func WithConstLabels(labels Labels) Option {
	return func(o *Options) {
		o.ConstLabels = labels
	}
}

// WithBuckets 设置直方图桶
func WithBuckets(buckets []float64) Option {
	return func(o *Options) {
		o.Buckets = buckets
	}
}

// WithObjectives 设置摘要目标
func WithObjectives(objectives map[float64]float64) Option {
	return func(o *Options) {
		o.Objectives = objectives
	}
}

// WithMaxAge 设置最大存活时间
func WithMaxAge(maxAge time.Duration) Option {
	return func(o *Options) {
		o.MaxAge = maxAge
	}
}

// WithAgeBuckets 设置存活时间桶数量
func WithAgeBuckets(ageBuckets uint32) Option {
	return func(o *Options) {
		o.AgeBuckets = ageBuckets
	}
}

// WithBufCap 设置缓冲区容量
func WithBufCap(bufCap uint32) Option {
	return func(o *Options) {
		o.BufCap = bufCap
	}
}
