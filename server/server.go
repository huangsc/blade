package server

import (
	"context"
)

// Server 接口定义了服务器的基本行为
type Server interface {
	// Start 启动服务器
	Start() error
	// Stop 优雅关闭服务器
	Stop(context.Context) error
}

// Options 定义服务器配置选项
type Options struct {
	Address     string            // 服务地址
	Port        int               // 服务端口
	Middleware  []interface{}     // 中间件列表
	MetricsPath string            // 监控指标路径
	Headers     map[string]string // 全局响应头
}

// Option 定义配置函数类型
type Option func(*Options)

// WithAddress 设置服务地址
func WithAddress(addr string) Option {
	return func(o *Options) {
		o.Address = addr
	}
}

// WithPort 设置服务端口
func WithPort(port int) Option {
	return func(o *Options) {
		o.Port = port
	}
}

// WithMiddleware 添加中间件
func WithMiddleware(middleware ...interface{}) Option {
	return func(o *Options) {
		o.Middleware = append(o.Middleware, middleware...)
	}
}

// WithMetricsPath 设置监控指标路径
func WithMetricsPath(path string) Option {
	return func(o *Options) {
		o.MetricsPath = path
	}
}
