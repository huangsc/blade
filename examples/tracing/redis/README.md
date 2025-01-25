# Redis 追踪示例

本示例演示了如何使用 Knife 框架的 Tracing 中间件对 Redis 操作进行分布式追踪。

## 功能特性

- 基本操作追踪：SET、GET、DEL 等命令
- 管道操作追踪：Pipeline 批量命令执行
- 事务操作追踪：MULTI/EXEC 事务处理
- 上下文传播：跨服务调用的上下文传递
- 错误处理：异常情况的追踪记录
- 属性记录：缓存大小、TTL、命中率等指标

## 前置条件

1. 安装 Go 1.16 或更高版本
2. 安装并运行 Redis 服务器
3. 安装并运行 OpenTelemetry Collector
4. 安装并运行 Jaeger（用于查看追踪数据）

## 快速开始

1. 启动 Redis 服务器：
```bash
redis-server
```

2. 启动 OpenTelemetry Collector：
```bash
otelcol --config otel-collector-config.yaml
```

3. 启动 Jaeger：
```bash
docker run -d --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest
```

4. 运行示例程序：
```bash
go run main.go
```

5. 访问 Jaeger UI 查看追踪数据：
```
http://localhost:16686
```

## 代码结构

- `main.go`：主程序，包含 Redis 操作和追踪逻辑
- `otel-collector-config.yaml`：OpenTelemetry Collector 配置文件

## 追踪点说明

1. SET 操作
   - 操作类型：写入
   - 追踪属性：键名、值大小、TTL
   - 错误处理：序列化错误、Redis 错误

2. GET 操作
   - 操作类型：读取
   - 追踪属性：键名、值大小、缓存命中
   - 错误处理：Redis 错误、反序列化错误

3. DEL 操作
   - 操作类型：删除
   - 追踪属性：键名
   - 错误处理：Redis 错误

4. Pipeline 操作
   - 操作类型：批量
   - 追踪属性：命令数量、计数器值
   - 错误处理：管道执行错误

5. Transaction 操作
   - 操作类型：事务
   - 追踪属性：事务类型、键名
   - 错误处理：事务执行错误、乐观锁错误

## 最佳实践

1. 命名规范
   - Span 名称：使用 "redis.操作" 格式
   - 属性名称：使用 "db.system"、"db.operation" 等标准属性

2. 属性设置
   - 记录关键操作信息：键名、值大小等
   - 避免记录敏感信息：密码、个人数据等

3. 错误处理
   - 使用 SetError 记录错误信息
   - 区分不同类型的错误：网络、超时等

4. 上下文传播
   - 正确传递和使用 context
   - 设置合适的超时时间

## 常见问题

1. 看不到追踪数据？
   - 检查 Collector 是否正常运行
   - 确认采样率设置是否正确
   - 验证 Jaeger 端口是否可访问

2. 追踪数据不完整？
   - 检查 context 是否正确传递
   - 确认所有操作都已添加追踪

3. 性能问题？
   - 调整采样率
   - 优化属性记录数量
   - 使用异步导出器

## 参考资料

- [OpenTelemetry 文档](https://opentelemetry.io/docs/)
- [Redis 文档](https://redis.io/documentation)
- [go-redis 文档](https://redis.uptrace.dev/) 