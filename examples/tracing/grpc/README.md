# gRPC Tracing 示例

这个示例展示了如何在 gRPC 服务中使用 Blade 框架的 Tracing 中间件来实现分布式追踪。

## 功能特点

- 支持 gRPC 一元调用和流式调用的追踪
- 支持服务端和客户端追踪
- 支持上下文传播
- 支持多级 Span 创建
- 支持属性和事件记录
- 支持错误处理和状态设置

## 前置条件

1. 安装 Go 1.21 或更高版本
2. 安装 Protocol Buffers 编译器
3. 安装 Go Protocol Buffers 插件：
   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```
4. 安装 Docker（用于运行 OpenTelemetry Collector）
5. 安装 Jaeger（用于查看追踪数据）

## 快速开始

1. 生成 gRPC 代码：

```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/hello.proto
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

4. 启动 gRPC 服务器：

```bash
go run server/main.go
```

5. 运行 gRPC 客户端：

```bash
go run client/main.go
```

6. 查看追踪数据：

访问 Jaeger UI：http://localhost:16686

## 代码结构

- `proto/`: Protocol Buffers 定义
  - `hello.proto`: 服务定义文件
- `server/`: 服务器端代码
  - `main.go`: 服务器实现
- `client/`: 客户端代码
  - `main.go`: 客户端实现

## 追踪点说明

### 服务器端

1. 一元调用 (SayHello)
   - 服务器接收请求
   - 处理请求
   - 返回响应

2. 流式调用 (SayHelloStream)
   - 服务器接收请求
   - 多次处理请求
   - 多次发送响应

### 客户端

1. 一元调用 (SayHello)
   - 创建请求
   - 发送请求
   - 接收响应

2. 流式调用 (SayHelloStream)
   - 创建请求
   - 发送请求
   - 多次接收响应

## 最佳实践

1. 命名规范
   - 使用 `grpc.{service}.{method}` 格式命名 Span
   - 使用有意义的事件名称
   - 保持命名一致性

2. 属性设置
   - 记录 RPC 系统信息
   - 记录服务和方法名
   - 记录请求参数（注意敏感信息）

3. 错误处理
   - 记录所有错误
   - 提供错误上下文
   - 正确设置状态码

4. 上下文传播
   - 使用 gRPC 元数据传递上下文
   - 正确提取和注入上下文
   - 处理上下文丢失的情况

## 常见问题

1. 看不到追踪数据？
   - 检查 Collector 配置
   - 确认服务连接正常
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

- [OpenTelemetry gRPC 文档](https://opentelemetry.io/docs/instrumentation/go/manual/#grpc)
- [gRPC 文档](https://grpc.io/docs/) 