package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

// Client gRPC客户端
type Client struct {
	*grpc.ClientConn
	opts *Options
}

// New 创建gRPC客户端
func New(opts ...Option) (*Client, error) {
	options := &Options{
		Target:    "localhost:9000",
		Timeout:   time.Second * 5,
		KeepAlive: true,
	}
	for _, o := range opts {
		o(options)
	}

	var dialOpts []grpc.DialOption

	// 设置安全选项
	if !options.Secure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// 添加拦截器
	if len(options.UnaryInterceptors) > 0 {
		dialOpts = append(dialOpts, grpc.WithChainUnaryInterceptor(options.UnaryInterceptors...))
	}
	if len(options.StreamInterceptors) > 0 {
		dialOpts = append(dialOpts, grpc.WithChainStreamInterceptor(options.StreamInterceptors...))
	}

	// 设置keepalive参数
	if options.KeepAlive {
		dialOpts = append(dialOpts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second, // 每10秒ping一次
			Timeout:             3 * time.Second,  // 3秒内没有响应则认为连接断开
			PermitWithoutStream: true,             // 允许在没有RPC的情况下发送ping
		}))
	}

	ctx, cancel := context.WithTimeout(context.Background(), options.Timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, options.Target, dialOpts...)
	if err != nil {
		return nil, err
	}

	client := &Client{
		ClientConn: conn,
		opts:       options,
	}

	// 检查健康状态
	if options.EnableHealthCheck {
		if err := client.checkHealth(ctx); err != nil {
			conn.Close()
			return nil, err
		}
	}

	return client, nil
}

// checkHealth 检查服务健康状态
func (c *Client) checkHealth(ctx context.Context) error {
	healthClient := grpc_health_v1.NewHealthClient(c.ClientConn)
	_, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	return err
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	return c.ClientConn.Close()
}
