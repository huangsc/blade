# ETCD 配置中心示例

本示例展示了如何使用 Knife 框架的 ETCD 配置中心组件进行配置管理。

## 前置条件

1. 安装并启动 ETCD 服务
```bash
# MacOS
brew install etcd
brew services start etcd

# Docker
docker run -d \
  --name etcd \
  -p 2379:2379 \
  -p 2380:2380 \
  gcr.io/etcd-development/etcd:v3.5.0
```

2. 安装依赖
```bash
go mod tidy
```

## 运行示例

```bash
go run main.go
```

## 功能演示

1. 初始化配置
```bash
# 设置服务器配置
etcdctl put /myapp/config/server '{"host":"localhost","port":8080}'

# 设置数据库配置
etcdctl put /myapp/config/database '{"host":"localhost","port":3306,"user":"root","password":"123456","name":"myapp"}'

# 设置Redis配置
etcdctl put /myapp/config/redis '{"host":"localhost","port":6379,"password":"","db":0}'
```

2. 更新配置
```bash
# 更新数据库配置
etcdctl put /myapp/config/database '{"host":"localhost","port":3306,"user":"admin","password":"new-password","name":"myapp"}'
```

3. 删除配置
```bash
# 删除Redis配置
etcdctl del /myapp/config/redis
```

## 功能特性

1. 配置管理
   - 支持 JSON 格式配置
   - 支持多种数据类型
   - 支持配置嵌套
   - 支持配置扫描到结构体

2. 配置监听
   - 实时监听配置变更
   - 支持创建、更新、删除事件
   - 支持优雅关闭

3. 类型转换
   - 支持布尔值
   - 支持整数值
   - 支持浮点值
   - 支持字符串
   - 支持时间间隔
   - 支持时间值
   - 支持切片
   - 支持映射

## 注意事项

1. 生产环境使用时需要：
   - 配置 ETCD 集群
   - 启用 TLS 安全连接
   - 配置访问认证
   - 设置合适的 TTL
   - 添加错误重试机制

2. 最佳实践：
   - 使用合适的配置前缀
   - 合理组织配置结构
   - 避免频繁更新配置
   - 注意配置大小限制
   - 做好配置备份 