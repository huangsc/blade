# Tracing 示例

这个示例展示了如何使用 Blade 框架的 Tracing 中间件来实现分布式追踪。

## 功能特点

- 基于 OpenTelemetry 实现
- 支持 HTTP 请求追踪
- 支持上下文传播
- 支持多级 Span 创建
- 支持属性和事件记录
- 支持错误处理和状态设置

## 前置条件

1. 安装 Go 1.21 或更高版本
2. 安装 Docker（用于运行 OpenTelemetry Collector）
3. 安装 Jaeger（用于查看追踪数据）

## 快速开始

1. 启动 OpenTelemetry Collector：

```bash
docker run -d --name otel-collector \
    -p 4317:4317 \
    -p 4318:4318 \
    -v $(pwd)/otel-collector-config.yaml:/etc/otelcol/config.yaml \
    otel/opentelemetry-collector-contrib
```

2. 启动 Jaeger：

```bash
docker run -d --name jaeger \
    -p 16686:16686 \
    -p 14250:14250 \
    jaegertracing/all-in-one
```

3. 运行示例程序：

```bash
go run main.go
```

4. 发送测试请求：

```bash
curl http://localhost:8080/hello
```

5. 查看追踪数据：

访问 Jaeger UI：http://localhost:16686

## 代码结构

- `main.go`: 主程序，包含 HTTP 服务器和请求处理逻辑
- `otel-collector-config.yaml`: OpenTelemetry Collector 配置文件

## 配置说明

### 追踪器配置

```go
tracer, err := tracing.NewOTelTracer(
    tracing.WithServiceName("example"),
    tracing.WithServiceVersion("v1.0.0"),
    tracing.WithEnvironment("development"),
    tracing.WithEndpoint("localhost:4317"),
    tracing.WithSampler(1.0),
)
```

### Collector 配置

创建 `otel-collector-config.yaml` 文件：

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 1s
    send_batch_size: 1024

exporters:
  jaeger:
    endpoint: jaeger:14250
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [jaeger]
```

## 示例功能

1. HTTP 请求追踪
   - 记录请求方法和 URL
   - 传播追踪上下文
   - 记录请求处理时间

2. 业务处理追踪
   - 记录处理开始和结束事件
   - 记录处理参数和结果
   - 支持自定义属性

3. 数据库操作追踪
   - 记录数据库操作类型
   - 记录数据库连接信息
   - 处理和记录错误

## 最佳实践

1. Span 命名
   - 使用有意义的名称
   - 遵循 `{operation_name}` 格式
   - 保持命名一致性

2. 属性设置
   - 使用标准属性名
   - 只记录有价值的信息
   - 避免敏感信息

3. 错误处理
   - 始终记录错误
   - 提供错误上下文
   - 设置适当的状态

4. 上下文传播
   - 正确提取和注入上下文
   - 使用标准传播器
   - 处理跨服务调用

## 常见问题

1. 看不到追踪数据？
   - 检查 Collector 是否正常运行
   - 确认端口配置正确
   - 查看采样率设置

2. 数据延迟显示？
   - 检查批处理配置
   - 调整发送间隔
   - 确认网络连接

3. 内存占用过高？
   - 调整采样率
   - 减少属性数量
   - 优化批处理大小

## 参考资料

- [OpenTelemetry 文档](https://opentelemetry.io/docs/)
- [Jaeger 文档](https://www.jaegertracing.io/docs/)