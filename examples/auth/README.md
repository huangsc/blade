# 认证中间件示例

本示例演示了如何使用 Blade 框架的认证中间件来保护 API 接口。

## 功能特性

- 基于 JWT 的认证机制
- 支持令牌生成和验证
- 支持角色验证
- 支持中间件链式调用
- 完整的错误处理

## 快速开始

1. 运行示例程序：
```bash
go run main.go
```

2. 登录获取令牌：
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123456"}'
```

3. 使用令牌访问受保护的API：
```bash
# 获取用户信息
curl http://localhost:8080/api/user \
  -H "Authorization: Bearer YOUR_TOKEN"

# 创建用户(需要管理员权限)
curl -X POST http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer YOUR_TOKEN"

# 删除用户(需要管理员权限)
curl -X DELETE http://localhost:8080/api/admin/users/1 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## API 接口

### POST /login
登录接口，用于获取访问令牌。

请求体：
```json
{
  "username": "admin",
  "password": "123456"
}
```

响应：
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### GET /api/user
获取当前用户信息。

响应：
```json
{
  "user_id": "1",
  "username": "admin",
  "role": "admin"
}
```

### POST /api/admin/users
创建用户(需要管理员权限)。

响应：
```json
{
  "message": "创建用户成功"
}
```

### DELETE /api/admin/users/:id
删除用户(需要管理员权限)。

响应：
```json
{
  "message": "删除用户成功"
}
```

## 使用说明

1. 创建认证器：
```go
authenticator := auth.NewJWTAuthenticator(
    "your-secret-key",
    time.Hour*24,
)
```

2. 添加认证中间件：
```go
api := r.Group("/api")
api.Use(auth.AuthMiddleware(authenticator))
```

3. 添加角色验证中间件：
```go
admin := api.Group("/admin")
admin.Use(auth.RequireRole("admin"))
```

4. 在处理函数中获取用户信息：
```go
claims, _ := c.Get(string(auth.ClaimsKey))
userClaims := claims.(*auth.Claims)
```

## 最佳实践

1. 密钥管理
   - 使用足够长的随机密钥
   - 定期轮换密钥
   - 使用环境变量存储密钥

2. 令牌设置
   - 设置合理的过期时间
   - 包含必要的用户信息
   - 避免存储敏感数据

3. 错误处理
   - 返回合适的HTTP状态码
   - 提供有意义的错误信息
   - 记录认证失败的日志

4. 安全建议
   - 使用HTTPS传输
   - 实现令牌刷新机制
   - 添加请求频率限制

## 常见问题

1. 令牌无效？
   - 检查令牌格式
   - 验证密钥是否正确
   - 确认令牌未过期

2. 权限不足？
   - 检查用户角色
   - 验证中间件顺序
   - 确认路由配置

3. 性能问题？
   - 使用令牌黑名单
   - 实现缓存机制
   - 优化验证逻辑 