package database

import (
	"context"
	"io"
	"time"

	"gorm.io/gorm"
)

// DB 数据库接口
type DB interface {
	// AutoMigrate 自动迁移
	AutoMigrate(dst ...interface{}) error

	// Create 创建记录
	Create(ctx context.Context, value interface{}) error

	// First 查询第一条记录
	First(ctx context.Context, dst interface{}, conds ...interface{}) error

	// Find 查询多条记录
	Find(ctx context.Context, dst interface{}, conds ...interface{}) error

	// Update 更新记录
	Update(ctx context.Context, model interface{}, column string, value interface{}) error

	// Updates 批量更新记录
	Updates(ctx context.Context, model interface{}, values interface{}) error

	// Delete 删除记录
	Delete(ctx context.Context, model interface{}, conds ...interface{}) error

	// Where 条件查询
	Where(query interface{}, args ...interface{}) DB

	// Begin 开启事务
	Begin(ctx context.Context, opts ...*TxOptions) (Tx, error)

	// Stats 获取连接池统计信息
	Stats() Stats

	// Close 关闭数据库连接
	Close() error

	// DB 获取原生GORM实例
	DB() *gorm.DB
}

// Tx 事务接口
type Tx interface {
	// Create 创建记录
	Create(ctx context.Context, value interface{}) error

	// First 查询第一条记录
	First(ctx context.Context, dst interface{}, conds ...interface{}) error

	// Find 查询多条记录
	Find(ctx context.Context, dst interface{}, conds ...interface{}) error

	// Update 更新记录
	Update(ctx context.Context, model interface{}, column string, value interface{}) error

	// Updates 批量更新记录
	Updates(ctx context.Context, model interface{}, values interface{}) error

	// Delete 删除记录
	Delete(ctx context.Context, model interface{}, conds ...interface{}) error

	// Where 条件查询
	Where(query interface{}, args ...interface{}) Tx

	// Commit 提交事务
	Commit() error

	// Rollback 回滚事务
	Rollback() error
}

// Config 数据库配置
type Config struct {
	// DSN 数据源名称
	DSN string

	// MaxOpenConns 最大连接数
	MaxOpenConns int

	// MaxIdleConns 最大空闲连接数
	MaxIdleConns int

	// ConnMaxLifetime 连接最大生命周期
	ConnMaxLifetime time.Duration

	// ConnMaxIdleTime 连接最大空闲时间
	ConnMaxIdleTime time.Duration

	// SlowThreshold 慢查询阈值
	SlowThreshold time.Duration

	// EnableTracing 是否启用链路追踪
	EnableTracing bool

	// Logger 日志输出
	Logger io.Writer
}

// TxOptions 事务选项
type TxOptions struct {
	// Isolation 事务隔离级别
	Isolation IsolationLevel

	// ReadOnly 是否只读事务
	ReadOnly bool
}

// IsolationLevel 事务隔离级别
type IsolationLevel int

const (
	// LevelDefault 默认隔离级别
	LevelDefault IsolationLevel = iota

	// LevelReadUncommitted 读未提交
	LevelReadUncommitted

	// LevelReadCommitted 读已提交
	LevelReadCommitted

	// LevelWriteCommitted 写已提交
	LevelWriteCommitted

	// LevelRepeatableRead 可重复读
	LevelRepeatableRead

	// LevelSnapshot 快照
	LevelSnapshot

	// LevelSerializable 串行化
	LevelSerializable

	// LevelLinearizable 线性化
	LevelLinearizable
)

// Stats 连接池统计信息
type Stats struct {
	// MaxOpenConnections 最大连接数
	MaxOpenConnections int

	// OpenConnections 打开的连接数
	OpenConnections int

	// InUse 使用中的连接数
	InUse int

	// Idle 空闲的连接数
	Idle int

	// WaitCount 等待连接的次数
	WaitCount int64

	// WaitDuration 等待连接的总时间
	WaitDuration time.Duration

	// MaxIdleClosed 因为超过最大空闲连接数而关闭的连接数
	MaxIdleClosed int64

	// MaxLifetimeClosed 因为超过最大生命周期而关闭的连接数
	MaxLifetimeClosed int64
}

// Option 配置选项函数
type Option func(*Config)

// WithDSN 设置数据源名称
func WithDSN(dsn string) Option {
	return func(c *Config) {
		c.DSN = dsn
	}
}

// WithMaxOpenConns 设置最大连接数
func WithMaxOpenConns(n int) Option {
	return func(c *Config) {
		c.MaxOpenConns = n
	}
}

// WithMaxIdleConns 设置最大空闲连接数
func WithMaxIdleConns(n int) Option {
	return func(c *Config) {
		c.MaxIdleConns = n
	}
}

// WithConnMaxLifetime 设置连接最大生命周期
func WithConnMaxLifetime(d time.Duration) Option {
	return func(c *Config) {
		c.ConnMaxLifetime = d
	}
}

// WithConnMaxIdleTime 设置连接最大空闲时间
func WithConnMaxIdleTime(d time.Duration) Option {
	return func(c *Config) {
		c.ConnMaxIdleTime = d
	}
}

// WithSlowThreshold 设置慢查询阈值
func WithSlowThreshold(d time.Duration) Option {
	return func(c *Config) {
		c.SlowThreshold = d
	}
}

// WithTracing 设置是否启用链路追踪
func WithTracing(enable bool) Option {
	return func(c *Config) {
		c.EnableTracing = enable
	}
}

// WithLogger 设置日志输出
func WithLogger(w io.Writer) Option {
	return func(c *Config) {
		c.Logger = w
	}
}
