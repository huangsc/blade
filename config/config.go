package config

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrNotFound 配置不存在错误
	ErrNotFound = errors.New("config: key not found")
	// ErrTypeAssert 类型断言错误
	ErrTypeAssert = errors.New("config: type assertion failed")
)

// Config 定义配置管理接口
type Config interface {
	// Load 加载配置
	Load() error
	// Get 获取配置值
	Get(key string) (Value, error)
	// Watch 监听配置变更
	Watch(ctx context.Context, key string) (Watcher, error)
}

// Value 定义配置值接口
type Value interface {
	// Bool 获取布尔值
	Bool() (bool, error)
	// Int 获取整数值
	Int() (int64, error)
	// Float 获取浮点值
	Float() (float64, error)
	// String 获取字符串值
	String() (string, error)
	// Duration 获取时间间隔
	Duration() (time.Duration, error)
	// Time 获取时间值
	Time() (time.Time, error)
	// Slice 获取切片值
	Slice() ([]Value, error)
	// Map 获取映射值
	Map() (map[string]Value, error)
	// Scan 将值扫描到结构体
	Scan(interface{}) error
}

// Watcher 定义配置监听接口
type Watcher interface {
	// Next 返回下一个配置变更
	Next() (*Change, error)
	// Stop 停止监听
	Stop() error
}

// Change 定义配置变更
type Change struct {
	// Key 配置键
	Key string
	// Value 新值
	Value Value
	// PreValue 旧值
	PreValue Value
	// Timestamp 变更时间
	Timestamp time.Time
	// Type 变更类型
	Type ChangeType
}

// ChangeType 定义配置变更类型
type ChangeType int

const (
	// Create 创建配置
	Create ChangeType = iota
	// Update 更新配置
	Update
	// Delete 删除配置
	Delete
)

// Source 定义配置源接口
type Source interface {
	// Load 加载配置
	Load() (map[string]interface{}, error)
	// Watch 监听配置变更
	Watch(ctx context.Context) (<-chan map[string]interface{}, error)
}

// Parser 定义配置解析器接口
type Parser interface {
	// Parse 解析配置
	Parse(map[string]interface{}) (map[string]Value, error)
}
