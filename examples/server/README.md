# 服务器示例

本目录包含了 HTTP 和 gRPC 服务器的使用示例。

## HTTP 服务器示例

### 功能特性

1. RESTful API 设计
2. 路由分组
3. 中间件支持
4. 全局响应头
5. 优雅关闭

### 运行示例

```bash
cd http
go run main.go
```

### API 接口

1. 健康检查
```bash
curl http://localhost:8080/v1/health
```

2. 创建用户
```bash
curl -X POST http://localhost:8080/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"测试用户","email":"test@example.com"}'
```

3. 获取用户
```bash
curl http://localhost:8080/v1/users/123
```

4. 更新用户
```bash
curl -X PUT http://localhost:8080/v1/users/123 \
  -H "Content-Type: application/json" \
  -d '{"name":"新名字","email":"new@example.com"}'
```

5. 删除用户
```bash
curl -X DELETE http://localhost:8080/v1/users/123
```

## gRPC 服务器示例

### 功能特性

1. Protocol Buffers 接口定义
2. 健康检查服务
3. 反射服务
4. 优雅关闭

### 生成代码

1. 安装工具
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

2. 生成代码
```bash
cd grpc
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/user.proto
```

### 运行示例

```bash
cd grpc
go run main.go
```

### 测试服务

可以使用 [grpcurl](https://github.com/fullstorydev/grpcurl) 工具测试 gRPC 服务：

1. 查看服务列表
```bash
grpcurl -plaintext localhost:9000 list
```

2. 查看服务方法
```bash
grpcurl -plaintext localhost:9000 list user.UserService
```

3. 创建用户
```bash
grpcurl -plaintext -d '{"name":"测试用户","email":"test@example.com"}' \
  localhost:9000 user.UserService/CreateUser
```

4. 获取用户
```bash
grpcurl -plaintext -d '{"id":"123"}' \
  localhost:9000 user.UserService/GetUser
```

5. 更新用户
```bash
grpcurl -plaintext -d '{"id":"123","name":"新名字","email":"new@example.com"}' \
  localhost:9000 user.UserService/UpdateUser
```

6. 删除用户
```bash
grpcurl -plaintext -d '{"id":"123"}' \
  localhost:9000 user.UserService/DeleteUser
```

## 注意事项

1. 示例代码仅供参考，生产环境使用时需要：
   - 添加适当的日志记录
   - 添加错误处理
   - 添加参数验证
   - 添加安全措施
   - 添加监控指标

2. 生产环境配置建议：
   - 使用配置文件或环境变量管理配置
   - 启用 TLS/SSL
   - 设置适当的超时时间
   - 配置访问控制
   - 添加限流措施 