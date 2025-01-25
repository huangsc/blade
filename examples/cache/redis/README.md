# Redis 缓存示例

本示例演示了如何使用 Blade 框架的 Redis 缓存组件。

## 功能特性

- 支持 Redis 缓存
- 支持 TTL 过期机制
- 支持键前缀
- 支持连接池
- 支持移除回调
- 支持 JSON 序列化
- 线程安全

## 前置条件

1. 安装 Go 1.16 或更高版本
2. 安装并启动 Redis 服务器

## 快速开始

1. 启动 Redis 服务器：
```bash
redis-server
```

2. 运行示例程序：
```bash
go run main.go
```

## 使用方法

### 1. 创建缓存实例

```go
cache := cache.NewRedisCache(cache.RedisOptions{
    Options: cache.Options{
        TTL: time.Second * 5, // 默认5秒过期
        OnEvicted: func(key string, value interface{}) {
            log.Printf("缓存项被移除: key=%s, value=%v\n", key, value)
        },
    },
    Addr:         "localhost:6379", // Redis地址
    Password:     "",               // Redis密码
    DB:           0,                // 使用默认数据库
    PoolSize:     10,               // 连接池大小
    MinIdleConns: 5,                // 最小空闲连接
    KeyPrefix:    "cache:",         // 键前缀
})
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

### 3. 结构体操作

```go
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

user := User{
    ID:   1,
    Name: "张三",
}

// 存储结构体
err := cache.Set(ctx, "user:1", user, time.Minute)

// 获取结构体
value, err := cache.Get(ctx, "user:1")
```

## 最佳实践

1. Redis 配置
   - 根据需求设置合适的连接池大小
   - 使用有意义的键前缀
   - 启用持久化机制

2. 错误处理
   - 处理连接错误
   - 处理序列化错误
   - 实现优雅的降级策略

3. 性能优化
   - 合理设置连接池参数
   - 使用批量操作
   - 避免大键值对

## 常见问题

1. 连接问题？
   - 检查 Redis 服务是否启动
   - 验证连接参数是否正确
   - 确认防火墙设置

2. 性能问题？
   - 调整连接池大小
   - 使用 pipeline 批量操作
   - 监控 Redis 指标

3. 内存问题？
   - 启用 maxmemory 限制
   - 配置合适的淘汰策略
   - 监控内存使用情况

## 参考资料

- [Redis 官方文档](https://redis.io/documentation)
- [Go-Redis 文档](https://redis.uptrace.dev/guide/)
- [Redis 最佳实践](https://redis.io/topics/memory-optimization) 