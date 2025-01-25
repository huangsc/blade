package grpc

import (
	"time"

	"google.golang.org/grpc"
)

// Options gRPC客户端配置选项
type Options struct {
	Target             string                         // 目标地址
	Timeout            time.Duration                  // 连接超时时间
	KeepAlive          bool                           // 是否启用心跳
	EnableHealthCheck  bool                           // 是否启用健康检查
	Secure             bool                           // 是否启用安全连接
	UnaryInterceptors  []grpc.UnaryClientInterceptor  // 一元拦截器
	StreamInterceptors []grpc.StreamClientInterceptor // 流式拦截器
}

// Option 定义配置函数类型
type Option func(*Options)

// WithTarget 设置目标地址
func WithTarget(target string) Option {
	return func(o *Options) {
		o.Target = target
	}
}

// WithTimeout 设置连接超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

// WithKeepAlive 设置是否启用心跳
func WithKeepAlive(enable bool) Option {
	return func(o *Options) {
		o.KeepAlive = enable
	}
}

// WithHealthCheck 设置是否启用健康检查
func WithHealthCheck(enable bool) Option {
	return func(o *Options) {
		o.EnableHealthCheck = enable
	}
}

// WithSecure 设置是否启用安全连接
func WithSecure(enable bool) Option {
	return func(o *Options) {
		o.Secure = enable
	}
}

// WithUnaryInterceptors 添加一元拦截器
func WithUnaryInterceptors(interceptors ...grpc.UnaryClientInterceptor) Option {
	return func(o *Options) {
		o.UnaryInterceptors = append(o.UnaryInterceptors, interceptors...)
	}
}

// WithStreamInterceptors 添加流式拦截器
func WithStreamInterceptors(interceptors ...grpc.StreamClientInterceptor) Option {
	return func(o *Options) {
		o.StreamInterceptors = append(o.StreamInterceptors, interceptors...)
	}
}
