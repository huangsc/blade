# 限流中间件示例

本示例演示了如何使用 Blade 框架的限流中间件来控制请求速率。

## 功能特性

- 基于令牌桶算法的限流
- 支持内存和 Redis 两种存储方式
- 支持分布式限流
- 提供多种限流策略：
  - 直接拒绝（Allow）
  - 等待通过（Wait）
  - 预约令牌（Reserve）
- 支持自定义限流规则：
  - 速率限制（Rate）
  - 突发流量（Burst）
  - 过期时间（TTL）

## 前置条件

1. 安装 Go 1.16 或更高版本
2. 安装并运行 Redis 服务器（如果使用分布式限流）

## 快速开始

1. 启动 Redis 服务器（如果使用分布式限流）：
```bash
redis-server
```

2. 运行示例程序：
```bash
go run main.go
```

## 使用方法

### 1. 创建内存限流器

```go
limiter := ratelimit.NewTokenBucket(
    ratelimit.WithRate(10),   // 每秒生成10个令牌
    ratelimit.WithBurst(100), // 最多存储100个令牌
)
```

### 2. 创建分布式限流器

```go
rdb := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
})

limiter := ratelimit.NewRedisLimiter(rdb,
    ratelimit.WithRate(10),       // 每秒生成10个令牌
    ratelimit.WithBurst(100),     // 最多存储100个令牌
    ratelimit.WithTTL(time.Hour), // 键的过期时间为1小时
)
```

### 3. 使用限流器

#### 直接拒绝模式
```go
allowed, err := limiter.Allow(ctx, "key", 1)
if err != nil {
    // 处理错误
}
if allowed {
    // 允许请求通过
} else {
    // 请求被拒绝
}
```

#### 等待模式
```go
err := limiter.Wait(ctx, "key", 1)
if err != nil {
    // 处理错误
}
// 请求通过
```

#### 预约模式
```go
reservation, err := limiter.Reserve(ctx, "key", 1)
if err != nil {
    // 处理错误
}
if !reservation.OK {
    // 预约失败
    return
}
delay := reservation.Delay()
if delay > maxWait {
    // 等待时间太长，取消预约
    reservation.Cancel()
    return
}
time.Sleep(delay)
// 使用预约的令牌
```

## 最佳实践

1. 选择合适的限流器
   - 单机应用：使用内存限流器
   - 分布式应用：使用 Redis 限流器

2. 设置合理的参数
   - Rate：根据系统处理能力设置
   - Burst：考虑突发流量的大小
   - TTL：避免 Redis 中的键过期

3. 错误处理
   - 检查 Allow/Wait/Reserve 的返回值
   - 正确处理上下文取消
   - 处理 Redis 连接错误

4. 性能优化
   - 合理设置 Redis 连接池
   - 避免过长的等待时间
   - 及时取消不需要的预约

## 常见问题

1. 限流不生效？
   - 检查 Rate 和 Burst 参数是否合理
   - 确认 Redis 连接是否正常
   - 验证键是否正确设置

2. 等待时间过长？
   - 调整 Rate 参数
   - 增加 Burst 值
   - 使用预约模式并设置最大等待时间

3. Redis 内存占用过高？
   - 设置合理的 TTL
   - 及时清理无用的键
   - 监控 Redis 内存使用情况

## 参考资料

- [令牌桶算法](https://en.wikipedia.org/wiki/Token_bucket)
- [Redis 文档](https://redis.io/documentation)
- [Go Rate Limiting 模式](https://golang.org/x/time/rate) 