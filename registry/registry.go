package registry

import (
	"context"
	"time"
)

// EventType 定义服务事件类型
type EventType int

const (
	// EventCreate 服务创建事件
	EventCreate EventType = iota
	// EventUpdate 服务更新事件
	EventUpdate
	// EventDelete 服务删除事件
	EventDelete
)

// Registry 定义服务注册与发现接口
type Registry interface {
	// Register 注册服务
	Register(ctx context.Context, service *ServiceInstance) error
	// Deregister 注销服务
	Deregister(ctx context.Context, service *ServiceInstance) error
	// GetService 获取服务实例列表
	GetService(ctx context.Context, serviceName string) ([]*ServiceInstance, error)
	// Watch 监听服务变更
	Watch(ctx context.Context, serviceName string) (Watcher, error)
}

// ServiceInstance 定义服务实例
type ServiceInstance struct {
	ID        string            // 实例ID
	Name      string            // 服务名称
	Version   string            // 服务版本
	Metadata  map[string]string // 服务元数据
	Endpoints []string          // 服务地址列表
	Status    int               // 服务状态
	TTL       time.Duration     // 服务TTL
}

// Watcher 定义服务监听接口
type Watcher interface {
	// Next 返回服务变更事件
	Next() ([]*ServiceInstance, error)
	// Stop 停止监听
	Stop() error
}

// Event 定义服务变更事件类型
type Event struct {
	Type     EventType
	Service  *ServiceInstance
	PreValue *ServiceInstance // 变更前的值
}
