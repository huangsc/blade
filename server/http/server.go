package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Server HTTP服务器
type Server struct {
	*gin.Engine
	server *http.Server
	opts   *Options
}

// Options HTTP服务器配置选项
type Options struct {
	Address    string            // 服务地址
	Port       int               // 服务端口
	Timeout    time.Duration     // 超时时间
	Mode       string            // 运行模式
	Middleware []gin.HandlerFunc // 中间件列表
	Headers    map[string]string // 全局响应头
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

// WithMode 设置运行模式
func WithMode(mode string) Option {
	return func(o *Options) {
		o.Mode = mode
	}
}

// WithMiddleware 添加中间件
func WithMiddleware(middleware ...gin.HandlerFunc) Option {
	return func(o *Options) {
		o.Middleware = append(o.Middleware, middleware...)
	}
}

// WithHeaders 设置全局响应头
func WithHeaders(headers map[string]string) Option {
	return func(o *Options) {
		o.Headers = headers
	}
}

// New 创建HTTP服务器
func New(opts ...Option) *Server {
	options := &Options{
		Address: "0.0.0.0",
		Port:    8080,
		Timeout: time.Second * 30,
		Mode:    gin.ReleaseMode,
	}
	for _, o := range opts {
		o(options)
	}

	gin.SetMode(options.Mode)
	engine := gin.New()

	// 添加基础中间件
	engine.Use(gin.Recovery())

	// 添加自定义中间件
	if len(options.Middleware) > 0 {
		engine.Use(options.Middleware...)
	}

	// 添加全局响应头
	if len(options.Headers) > 0 {
		engine.Use(func(c *gin.Context) {
			for k, v := range options.Headers {
				c.Header(k, v)
			}
			c.Next()
		})
	}

	return &Server{
		Engine: engine,
		opts:   options,
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.opts.Address, s.opts.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.Engine,
		ReadTimeout:  s.opts.Timeout,
		WriteTimeout: s.opts.Timeout,
		IdleTimeout:  s.opts.Timeout * 2,
	}

	return s.server.ListenAndServe()
}

// Stop 停止服务器
func (s *Server) Stop(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}
