# MySQL 数据库示例

本示例演示了如何使用 Blade 框架的 MySQL 数据库组件。

## 功能特性

- 基于 GORM 实现，支持所有 GORM 特性
- 支持连接池管理
- 支持事务处理
- 支持慢查询监控
- 支持链路追踪
- 支持指标采集
- 线程安全

## 前置条件

1. 安装 Go 1.16 或更高版本
2. 安装并启动 MySQL 服务器

## 快速开始

1. 启动 MySQL 服务器：
```bash
mysql.server start
```

2. 创建测试数据库：
```sql
CREATE DATABASE test;
```

3. 运行示例程序：
```bash
go run main.go
```

## 使用方法

### 1. 创建数据库实例

```go
db, err := mysql.NewDB(
    database.WithDSN("root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"),
    database.WithMaxOpenConns(10),
    database.WithMaxIdleConns(5),
    database.WithConnMaxLifetime(time.Hour),
    database.WithSlowThreshold(time.Millisecond*100),
    database.WithTracing(true),
    database.WithLogger(os.Stdout),
)
```

### 2. 定义模型

```go
type User struct {
    ID        uint      `gorm:"primarykey"`
    Name      string    `gorm:"size:100;not null"`
    Email     string    `gorm:"size:100;uniqueIndex"`
    Age       int       `gorm:"not null"`
    CreatedAt time.Time `gorm:"not null"`
    UpdatedAt time.Time `gorm:"not null"`
}
```

### 3. 基本操作

```go
// 自动迁移
db.AutoMigrate(&User{})

// 创建
user := &User{Name: "张三", Email: "zhangsan@example.com", Age: 20}
db.Create(ctx, user)

// 查询
var user User
db.First(ctx, &user, 1)
db.Where("name = ?", "张三").First(ctx, &user)

// 更新
db.Update(ctx, &user, "name", "李四")
db.Updates(ctx, &user, map[string]interface{}{
    "name": "李四",
    "age":  21,
})

// 删除
db.Delete(ctx, &user)
```

### 4. 事务处理

```go
tx, err := db.Begin(ctx)
if err != nil {
    return err
}

if err := tx.Create(ctx, &user1); err != nil {
    tx.Rollback()
    return err
}

if err := tx.Create(ctx, &user2); err != nil {
    tx.Rollback()
    return err
}

return tx.Commit()
```

## 最佳实践

1. 连接池配置
   - 根据负载设置合适的连接数
   - 设置合理的连接生命周期
   - 监控连接池状态

2. 事务处理
   - 仅在必要时使用事务
   - 避免长事务
   - 正确处理事务错误

3. 性能优化
   - 合理使用索引
   - 避免大事务
   - 使用批量操作
   - 启用预编译语句缓存

## 常见问题

1. 连接问题？
   - 检查数据库服务是否启动
   - 验证连接字符串是否正确
   - 确认防火墙设置

2. 性能问题？
   - 检查慢查询日志
   - 分析查询计划
   - 优化索引设计

3. 内存问题？
   - 控制连接池大小
   - 及时关闭连接
   - 避免大量查询

## 参考资料

- [GORM 官方文档](https://gorm.io/docs/)
- [MySQL 官方文档](https://dev.mysql.com/doc/) 