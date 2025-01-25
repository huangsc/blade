# 缓存示例

本示例演示了如何使用 Knife 框架的缓存组件。

## 功能特性

- 支持内存缓存
- 支持 TTL 过期机制
- 支持最大条目限制
- 支持移除回调
- 支持自动清理过期项
- 线程安全

## 前置条件

1. 安装 Go 1.16 或更高版本

## 快速开始

1. 运行示例程序：
```bash
go run main.go
```

## 使用方法

### 1. 创建缓存实例

```go
cache := cache.NewMemoryCache(
    cache.WithTTL(time.Second*5),           // 默认5秒过期
    cache.WithMaxEntries(100),              // 最多存储100个条目
    cache.WithOnEvicted(func(key string, value interface{}) {
        log.Printf("缓存项被移除: key=%s, value=%v\n", key, value)
    }),
)
defer cache.Close()
```

### 2. 基本操作

```go
// 设置缓存
err := cache.Set(ctx, "key", "value", time.Second*10)

// 获取缓存
value, err := cache.Get(ctx, "key")

// 删除缓存
err := cache.Delete(ctx, "key")

// 清空缓存
err := cache.Clear(ctx)

// 获取缓存数量
count := cache.Len(ctx)
```

## 最佳实践

1. 缓存配置
   - 根据数据特性设置合适的 TTL
   - 设置合理的最大条目数
   - 实现合适的移除回调函数

2. 错误处理
   - 区分不同类型的错误
   - 正确处理缓存未命中的情况
   - 实现优雅的降级策略

3. 性能优化
   - 使用读写锁提高并发性能
   - 定期清理过期项
   - 合理设置缓存容量

## 常见问题

1. 缓存项不过期？
   - 检查 TTL 设置是否正确
   - 确认清理器是否正常运行
   - 验证时间设置是否合理

2. 内存占用过高？
   - 调整最大条目数
   - 减少 TTL 时间
   - 增加清理频率

3. 性能问题？
   - 使用性能分析工具定位瓶颈
   - 优化锁的使用
   - 调整缓存参数

## 参考资料

- [缓存设计模式](https://martinfowler.com/bliki/TwoLevelCache.html)
- [Go 内存管理](https://golang.org/doc/gc-guide)
- [并发编程](https://golang.org/doc/effective_go#concurrency) 