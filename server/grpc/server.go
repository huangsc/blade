package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

// Server gRPC服务器
type Server struct {
	*grpc.Server
	opts   *Options
	health *health.Server
	lis    net.Listener
}

// Options gRPC服务器配置选项
type Options struct {
	Address            string                         // 服务地址
	Port               int                            // 服务端口
	Timeout            time.Duration                  // 超时时间
	EnableHealth       bool                           // 是否启用健康检查
	EnableReflect      bool                           // 是否启用反射服务
	UnaryInterceptors  []grpc.UnaryServerInterceptor  // 一元拦截器
	StreamInterceptors []grpc.StreamServerInterceptor // 流式拦截器
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

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

// WithHealth 设置是否启用健康检查
func WithHealth(enable bool) Option {
	return func(o *Options) {
		o.EnableHealth = enable
	}
}

// WithReflection 设置是否启用反射服务
func WithReflection(enable bool) Option {
	return func(o *Options) {
		o.EnableReflect = enable
	}
}

// WithUnaryInterceptors 添加一元拦截器
func WithUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) Option {
	return func(o *Options) {
		o.UnaryInterceptors = append(o.UnaryInterceptors, interceptors...)
	}
}

// WithStreamInterceptors 添加流式拦截器
func WithStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) Option {
	return func(o *Options) {
		o.StreamInterceptors = append(o.StreamInterceptors, interceptors...)
	}
}

// New 创建gRPC服务器
func New(opts ...Option) *Server {
	options := &Options{
		Address:      "0.0.0.0",
		Port:         9000,
		Timeout:      time.Second * 30,
		EnableHealth: true,
	}
	for _, o := range opts {
		o(options)
	}

	var serverOpts []grpc.ServerOption

	// 添加拦截器
	if len(options.UnaryInterceptors) > 0 {
		serverOpts = append(serverOpts, grpc.ChainUnaryInterceptor(options.UnaryInterceptors...))
	}
	if len(options.StreamInterceptors) > 0 {
		serverOpts = append(serverOpts, grpc.ChainStreamInterceptor(options.StreamInterceptors...))
	}

	// 添加keepalive策略
	serverOpts = append(serverOpts, grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     time.Minute * 5,
		MaxConnectionAge:      time.Hour * 4,
		MaxConnectionAgeGrace: time.Second * 30,
		Time:                  time.Second * 60,
		Timeout:               time.Second * 20,
	}))

	srv := grpc.NewServer(serverOpts...)
	s := &Server{
		Server: srv,
		opts:   options,
	}

	// 注册健康检查服务
	if options.EnableHealth {
		s.health = health.NewServer()
		healthpb.RegisterHealthServer(srv, s.health)
	}

	// 注册反射服务
	if options.EnableReflect {
		reflection.Register(srv)
	}

	return s
}

// Start 启动服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.opts.Address, s.opts.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.lis = lis

	// 设置所有服务健康状态为SERVING
	if s.health != nil {
		s.health.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	}

	return s.Serve(lis)
}

// Stop 停止服务器
func (s *Server) Stop(ctx context.Context) error {
	// 设置所有服务健康状态为NOT_SERVING
	if s.health != nil {
		s.health.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	}

	s.GracefulStop()
	return nil
}
