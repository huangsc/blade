package etcd

import (
	"context"
	"encoding/json"
	"path"
	"sync"
	"time"

	"github.com/huangsc/blade/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	// defaultTTL 默认服务TTL时间
	defaultTTL = time.Second * 15
	// defaultRegisterTimeout 默认注册超时时间
	defaultRegisterTimeout = time.Second * 5
	// defaultWatchTimeout 默认监听超时时间
	defaultWatchTimeout = time.Second * 5
)

// Registry etcd注册中心
type Registry struct {
	client      *clientv3.Client
	prefix      string          // 服务前缀
	ttl         time.Duration   // 服务TTL
	ctx         context.Context // 根上下文
	cancel      context.CancelFunc
	leaseID     clientv3.LeaseID // 租约ID
	keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse
	mutex       sync.Mutex
}

// Options 配置选项
type Options struct {
	Prefix string        // 服务前缀
	TTL    time.Duration // 服务TTL
}

// Option 定义配置函数类型
type Option func(*Options)

// WithPrefix 设置服务前缀
func WithPrefix(prefix string) Option {
	return func(o *Options) {
		o.Prefix = prefix
	}
}

// WithTTL 设置服务TTL
func WithTTL(ttl time.Duration) Option {
	return func(o *Options) {
		o.TTL = ttl
	}
}

// New 创建etcd注册中心
func New(client *clientv3.Client, opts ...Option) (*Registry, error) {
	options := &Options{
		Prefix: "/services",
		TTL:    defaultTTL,
	}
	for _, o := range opts {
		o(options)
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Registry{
		client: client,
		prefix: options.Prefix,
		ttl:    options.TTL,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Register 注册服务
func (r *Registry) Register(ctx context.Context, service *registry.ServiceInstance) error {
	key := r.serviceKey(service)
	value, err := json.Marshal(service)
	if err != nil {
		return err
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 创建租约
	grant, err := r.client.Grant(ctx, int64(r.ttl.Seconds()))
	if err != nil {
		return err
	}

	// 注册服务
	_, err = r.client.Put(ctx, key, string(value), clientv3.WithLease(grant.ID))
	if err != nil {
		return err
	}

	// 保持租约
	keepAliveCh, err := r.client.KeepAlive(r.ctx, grant.ID)
	if err != nil {
		return err
	}

	r.leaseID = grant.ID
	r.keepAliveCh = keepAliveCh

	// 处理续约响应
	go func() {
		for {
			select {
			case _, ok := <-keepAliveCh:
				if !ok {
					return
				}
			case <-r.ctx.Done():
				return
			}
		}
	}()

	return nil
}

// Deregister 注销服务
func (r *Registry) Deregister(ctx context.Context, service *registry.ServiceInstance) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 撤销租约
	if r.leaseID != 0 {
		_, err := r.client.Revoke(ctx, r.leaseID)
		if err != nil {
			return err
		}
	}

	r.cancel()
	return nil
}

// GetService 获取服务实例列表
func (r *Registry) GetService(ctx context.Context, serviceName string) ([]*registry.ServiceInstance, error) {
	key := path.Join(r.prefix, serviceName)
	resp, err := r.client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	var items []*registry.ServiceInstance
	for _, kv := range resp.Kvs {
		si := &registry.ServiceInstance{}
		if err := json.Unmarshal(kv.Value, si); err != nil {
			return nil, err
		}
		items = append(items, si)
	}
	return items, nil
}

// Watch 监听服务变更
func (r *Registry) Watch(ctx context.Context, serviceName string) (registry.Watcher, error) {
	key := path.Join(r.prefix, serviceName)
	return newWatcher(ctx, key, r.client), nil
}

// serviceKey 生成服务键
func (r *Registry) serviceKey(service *registry.ServiceInstance) string {
	return path.Join(r.prefix, service.Name, service.ID)
}
