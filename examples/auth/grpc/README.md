# gRPC 认证示例

本示例演示了如何在 gRPC 服务中使用认证中间件。

## 功能特性

- 基于 JWT 的认证机制
- 支持一元 RPC 认证
- 支持流式 RPC 认证
- 支持角色验证
- 完整的错误处理

## 快速开始

1. 生成 protobuf 代码：
```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/user.proto
```

2. 启动服务器：
```bash
go run server/main.go
```

3. 运行客户端：
```bash
go run client/main.go
```

## 服务接口

### GetUser
获取用户信息。
- 普通用户只能获取自己的信息
- 管理员可以获取任何用户的信息

### ListUsers
获取用户列表（流式）。
- 只有管理员可以访问
- 支持分页

### UpdateUser
更新用户信息。
- 普通用户只能更新自己的信息
- 管理员可以更新任何用户的信息

### DeleteUser
删除用户。
- 只有管理员可以删除用户

## 使用说明

1. 创建认证器：
```go
authenticator := auth.NewJWTAuthenticator(
    "your-secret-key",
    time.Hour*24,
)
```

2. 服务器端添加认证拦截器：
```go
server := grpc.NewServer(
    grpc.UnaryInterceptor(auth.UnaryServerInterceptor(authenticator)),
    grpc.StreamInterceptor(auth.StreamServerInterceptor(authenticator)),
)
```

3. 客户端添加认证拦截器：
```go
conn, err := grpc.Dial("localhost:9000",
    grpc.WithInsecure(),
    grpc.WithUnaryInterceptor(auth.UnaryClientInterceptor(token)),
    grpc.WithStreamInterceptor(auth.StreamClientInterceptor(token)),
)
```

4. 在处理函数中获取用户信息：
```go
claims, ok := auth.FromContext(ctx)
if !ok {
    return nil, status.Error(codes.Internal, "failed to get claims from context")
}
```

## 最佳实践

1. 认证配置
   - 使用安全的密钥
   - 设置合理的过期时间
   - 使用环境变量存储敏感信息

2. 错误处理
   - 使用合适的 gRPC 状态码
   - 提供有意义的错误信息
   - 记录认证失败的日志

3. 性能优化
   - 使用连接池
   - 启用 keepalive
   - 合理设置超时时间

4. 安全建议
   - 使用 TLS 加密
   - 实现令牌刷新
   - 添加请求限制

## 常见问题

1. 认证失败？
   - 检查令牌格式
   - 验证密钥是否正确
   - 确认令牌未过期

2. 权限不足？
   - 检查用户角色
   - 验证权限规则
   - 确认请求路径

3. 连接问题？
   - 检查服务器地址
   - 验证网络连接
   - 确认防火墙设置 