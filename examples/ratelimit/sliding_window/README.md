# 滑动窗口限流器示例

本示例演示了如何使用 Knife 框架的滑动窗口限流器来控制请求速率。

## 功能特性

- 基于滑动窗口算法的限流
- 支持自定义窗口大小
- 支持自定义窗口精度
- 提供多种限流策略：
  - 直接拒绝（Allow）
  - 等待通过（Wait）
  - 预约请求（Reserve）
- 支持自定义限流规则：
  - 速率限制（Rate）
  - 窗口大小（Window）
  - 窗口精度（Precision）

## 前置条件

1. 安装 Go 1.16 或更高版本

## 快速开始

1. 运行示例程序：
```bash
go run main.go
```

## 使用方法

### 1. 创建滑动窗口限流器

```go
limiter := ratelimit.NewSlidingWindow(
    ratelimit.WithRate(10),                        // 每秒允许10个请求
    ratelimit.WithWindow(time.Second),             // 窗口大小为1秒
    ratelimit.WithPrecision(time.Millisecond*100), // 精度为100ms
)
```

### 2. 使用限流器

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
// 使用预约的配额
```

## 最佳实践

1. 选择合适的窗口大小
   - 较小的窗口：更精确的限流，但开销更大
   - 较大的窗口：开销更小，但可能出现突发流量

2. 设置合理的窗口精度
   - 较高的精度：更平滑的限流效果，但内存占用更大
   - 较低的精度：内存占用更小，但限流效果可能不够平滑

3. 错误处理
   - 检查 Allow/Wait/Reserve 的返回值
   - 正确处理上下文取消
   - 设置合适的超时时间

4. 性能优化
   - 定期清理过期的窗口数据
   - 避免过高的窗口精度
   - 及时取消不需要的预约

## 常见问题

1. 限流不够平滑？
   - 增加窗口精度
   - 减小窗口大小
   - 调整速率限制

2. 内存占用过高？
   - 减少窗口精度
   - 增大窗口大小
   - 及时清理过期数据

3. 响应延迟大？
   - 使用预约模式设置最大等待时间
   - 调整窗口参数
   - 优化清理策略

## 参考资料

- [滑动窗口算法](https://en.wikipedia.org/wiki/Sliding_window_protocol)
- [限流算法对比](https://konghq.com/blog/how-to-design-a-scalable-rate-limiting-algorithm)
- [Go 并发编程](https://golang.org/doc/effective_go#concurrency) 