package logger

import (
	"context"
	"time"
)

// Level 日志级别
type Level int8

const (
	// DebugLevel 调试级别
	DebugLevel Level = iota - 1
	// InfoLevel 信息级别
	InfoLevel
	// WarnLevel 警告级别
	WarnLevel
	// ErrorLevel 错误级别
	ErrorLevel
	// FatalLevel 致命级别
	FatalLevel
)

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// Logger 日志接口
type Logger interface {
	// Debug 输出调试日志
	Debug(msg string, fields ...Field)
	// Info 输出信息日志
	Info(msg string, fields ...Field)
	// Warn 输出警告日志
	Warn(msg string, fields ...Field)
	// Error 输出错误日志
	Error(msg string, fields ...Field)
	// Fatal 输出致命日志
	Fatal(msg string, fields ...Field)

	// WithContext 设置上下文
	WithContext(ctx context.Context) Logger
	// WithFields 设置字段
	WithFields(fields ...Field) Logger
	// WithLevel 设置日志级别
	WithLevel(level Level) Logger
}

// Option 配置选项
type Option func(*Options)

// Options 配置选项
type Options struct {
	// Level 日志级别
	Level Level
	// Fields 全局字段
	Fields []Field
	// TimeFormat 时间格式
	TimeFormat string
	// CallerSkip 调用者跳过层数
	CallerSkip int
	// EnableCaller 是否启用调用者信息
	EnableCaller bool
	// EnableStacktrace 是否启用堆栈跟踪
	EnableStacktrace bool
	// StacktraceLevel 堆栈跟踪级别
	StacktraceLevel Level
	// OutputPaths 输出路径
	OutputPaths []string
	// ErrorOutputPaths 错误输出路径
	ErrorOutputPaths []string
	// Development 是否为开发模式
	Development bool
}

// WithLevel 设置日志级别
func WithLevel(level Level) Option {
	return func(o *Options) {
		o.Level = level
	}
}

// WithFields 设置全局字段
func WithFields(fields ...Field) Option {
	return func(o *Options) {
		o.Fields = append(o.Fields, fields...)
	}
}

// WithTimeFormat 设置时间格式
func WithTimeFormat(format string) Option {
	return func(o *Options) {
		o.TimeFormat = format
	}
}

// WithCaller 设置调用者信息
func WithCaller(skip int) Option {
	return func(o *Options) {
		o.EnableCaller = true
		o.CallerSkip = skip
	}
}

// WithStacktrace 设置堆栈跟踪
func WithStacktrace(level Level) Option {
	return func(o *Options) {
		o.EnableStacktrace = true
		o.StacktraceLevel = level
	}
}

// WithOutputPaths 设置输出路径
func WithOutputPaths(paths ...string) Option {
	return func(o *Options) {
		o.OutputPaths = paths
	}
}

// WithErrorOutputPaths 设置错误输出路径
func WithErrorOutputPaths(paths ...string) Option {
	return func(o *Options) {
		o.ErrorOutputPaths = paths
	}
}

// WithDevelopment 设置开发模式
func WithDevelopment(development bool) Option {
	return func(o *Options) {
		o.Development = development
	}
}

// String 返回日志级别字符串
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Any 创建任意类型字段
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// String 创建字符串字段
func String(key string, value string) Field {
	return Field{Key: key, Value: value}
}

// Int 创建整数字段
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 创建64位整数字段
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 创建64位浮点数字段
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool 创建布尔字段
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Time 创建时间字段
func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value}
}

// Duration 创建时间间隔字段
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

// Error 创建错误字段
func Error(err error) Field {
	return Field{Key: "error", Value: err}
}
