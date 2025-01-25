# Kafka Tracing 示例

这个示例展示了如何在 Kafka 消息队列中使用 Knife 框架的 Tracing 中间件来实现分布式追踪。

## 功能特点

- 支持 Kafka 生产者和消费者追踪
- 支持消息头中的上下文传播
- 支持多级 Span 创建
- 支持属性和事件记录
- 支持错误处理和状态设置
- 支持消费者组和分区信息记录

## 前置条件

1. 安装 Go 1.21 或更高版本
2. 安装 Docker（用于运行 Kafka 和 OpenTelemetry Collector）
3. 安装 Jaeger（用于查看追踪数据）
4. 安装 librdkafka（Confluent Kafka Go 客户端依赖）：
   ```bash
   # macOS
   brew install librdkafka

   # Ubuntu
   apt-get install librdkafka-dev

   # CentOS
   yum install librdkafka-devel
   ```

## 快速开始

1. 启动 Kafka：

```bash
docker run -d --name kafka \
    -p 9092:9092 \
    -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
    -e KAFKA_LISTENERS=PLAINTEXT://0.0.0.0:9092 \
    -e KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181 \
    confluentinc/cp-kafka
```

2. 启动 OpenTelemetry Collector：

```bash
docker run -d --name otel-collector \
    -p 4317:4317 \
    -p 4318:4318 \
    -v $(pwd)/otel-collector-config.yaml:/etc/otelcol/config.yaml \
    otel/opentelemetry-collector-contrib
```

3. 启动 Jaeger：

```bash
docker run -d --name jaeger \
    -p 16686:16686 \
    -p 14250:14250 \
    jaegertracing/all-in-one
```

4. 启动消费者：

```bash
go run consumer/main.go
```

5. 启动生产者：

```bash
go run producer/main.go
```

6. 查看追踪数据：

访问 Jaeger UI：http://localhost:16686

## 代码结构

- `producer/`: 生产者代码
  - `main.go`: 生产者实现
- `consumer/`: 消费者代码
  - `main.go`: 消费者实现

## 追踪点说明

### 生产者

1. 消息发送
   - 创建消息
   - 注入追踪上下文
   - 发送消息
   - 处理投递报告

### 消费者

1. 消息消费
   - 提取追踪上下文
   - 处理消息
   - 记录消费结果

## 最佳实践

1. 命名规范
   - 使用 `kafka.{operation}` 格式命名 Span
   - 使用有意义的事件名称
   - 保持命名一致性

2. 属性设置
   - 记录消息系统信息
   - 记录主题和分区
   - 记录消息 ID 和内容
   - 记录消费者组信息

3. 错误处理
   - 记录所有错误
   - 提供错误上下文
   - 正确设置状态码

4. 上下文传播
   - 使用消息头传递上下文
   - 正确提取和注入上下文
   - 处理上下文丢失的情况

## 常见问题

1. 看不到追踪数据？
   - 检查 Collector 配置
   - 确认 Kafka 连接正常
   - 检查上下文传播

2. 丢失部分 Span？
   - 检查 Span 结束调用
   - 确认错误处理正确
   - 检查上下文传播

3. 性能问题？
   - 调整采样率
   - 优化属性数量
   - 使用异步导出

## 参考资料

- [OpenTelemetry 文档](https://opentelemetry.io/docs/)
- [Confluent Kafka Go 文档](https://docs.confluent.io/platform/current/clients/confluent-kafka-go/) 