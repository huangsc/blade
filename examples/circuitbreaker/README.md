# 熔断器示例

本示例演示了如何使用 Blade 框架的熔断器来保护服务免受故障的影响。

## 功能特性

- 基于状态机的熔断器实现
- 支持三种状态：
  - 关闭状态（正常工作）
  - 开启状态（熔断服务）
  - 半开状态（尝试恢复）
- 提供多种执行模式：
  - 直接执行（Execute）
  - 两阶段执行（Allow/Done）
- 支持自定义配置：
  - 最大请求数
  - 统计周期
  - 超时时间
  - 熔断条件
  - 状态变更回调

## 前置条件

1. 安装 Go 1.16 或更高版本

## 快速开始

1. 运行示例程序：
```bash
go run main.go
```

## 使用方法

### 1. 创建熔断器

```go
breaker := circuitbreaker.NewBreaker("example",
    circuitbreaker.WithMaxRequests(3),         // 半开状态下最多允许3个请求
    circuitbreaker.WithInterval(time.Minute),  // 统计周期为1分钟
    circuitbreaker.WithTimeout(time.Second*5), // 熔断超时时间为5秒
    circuitbreaker.WithReadyToTrip(func(counts circuitbreaker.Counts) bool {
        // 连续5次失败触发熔断
        return counts.ConsecutiveFailures >= 5
    }),
    circuitbreaker.WithOnStateChange(func(name string, from circuitbreaker.State, to circuitbreaker.State) {
        log.Printf("Circuit breaker %s changed from %v to %v\n", name, from, to)
    }),
)
```

### 2. 使用熔断器

#### 直接执行模式
```go
err := breaker.Execute(ctx, func() error {
    // 执行受保护的操作
    return callService()
})
if err != nil {
    // 处理错误
}
```

#### 两阶段执行模式
```go
// 第一阶段: 检查是否允许请求
done, err := breaker.Allow()
if err != nil {
    // 请求被拒绝
    return err
}

// 执行请求
err = callService()

// 第二阶段: 报告结果
done(err == nil)
```

## 最佳实践

1. 熔断器配置
   - 根据服务特性设置合适的阈值
   - 设置合理的超时时间和统计周期
   - 实现合适的熔断条件判断函数

2. 错误处理
   - 区分不同类型的错误
   - 正确处理熔断状态下的请求
   - 实现优雅的服务降级

3. 监控和告警
   - 监控熔断器状态变化
   - 记录关键指标和统计信息
   - 设置合适的告警阈值

4. 性能优化
   - 使用原子操作进行计数
   - 避免频繁的状态检查
   - 合理设置统计周期

## 常见问题

1. 熔断器不触发？
   - 检查熔断条件是否合适
   - 确认错误是否正确报告
   - 验证统计周期是否合理

2. 服务无法恢复？
   - 调整半开状态的最大请求数
   - 检查超时时间是否过长
   - 确认服务是否真正恢复

3. 误触发问题？
   - 调整熔断条件
   - 增加统计周期
   - 区分临时错误和永久错误

## 参考资料

- [断路器模式](https://martinfowler.com/bliki/CircuitBreaker.html)
- [微服务弹性设计](https://docs.microsoft.com/en-us/azure/architecture/patterns/circuit-breaker)
- [Go 并发编程](https://golang.org/doc/effective_go#concurrency) 