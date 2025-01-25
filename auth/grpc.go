package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor 创建一元 RPC 认证拦截器
func UnaryServerInterceptor(auth Authenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 从元数据中获取令牌
		claims, err := authenticateRequest(ctx, auth)
		if err != nil {
			return nil, err
		}

		// 将认证信息添加到上下文
		newCtx := NewContext(ctx, claims)
		return handler(newCtx, req)
	}
}

// StreamServerInterceptor 创建流式 RPC 认证拦截器
func StreamServerInterceptor(auth Authenticator) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// 从元数据中获取令牌
		claims, err := authenticateRequest(ss.Context(), auth)
		if err != nil {
			return err
		}

		// 将认证信息添加到上下文
		newCtx := NewContext(ss.Context(), claims)
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          newCtx,
		}
		return handler(srv, wrappedStream)
	}
}

// authenticateRequest 验证请求中的令牌
func authenticateRequest(ctx context.Context, auth Authenticator) (*Claims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, ErrMissingToken.Error())
	}

	// 从元数据中获取令牌
	values := md.Get("authorization")
	if len(values) == 0 {
		return nil, status.Error(codes.Unauthenticated, ErrMissingToken.Error())
	}

	// 解析令牌
	authHeader := values[0]
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, status.Error(codes.Unauthenticated, ErrInvalidToken.Error())
	}

	// 验证令牌
	claims, err := auth.ValidateToken(parts[1])
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return claims, nil
}

// wrappedServerStream 包装的服务器流
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context 实现 grpc.ServerStream 接口
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// UnaryClientInterceptor 创建一元 RPC 客户端拦截器
func UnaryClientInterceptor(token string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// 添加令牌到元数据
		newCtx := attachToken(ctx, token)
		return invoker(newCtx, method, req, reply, cc, opts...)
	}
}

// StreamClientInterceptor 创建流式 RPC 客户端拦截器
func StreamClientInterceptor(token string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		// 添加令牌到元数据
		newCtx := attachToken(ctx, token)
		return streamer(newCtx, desc, cc, method, opts...)
	}
}

// attachToken 将令牌添加到上下文
func attachToken(ctx context.Context, token string) context.Context {
	if token == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
}
