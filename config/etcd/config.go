package etcd

import (
	"context"
	"encoding/json"
	"path"
	"sync"
	"time"

	"github.com/huangsc/blade/config"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Config ETCD配置中心
type Config struct {
	client  *clientv3.Client
	prefix  string
	values  map[string]config.Value
	parser  config.Parser
	mu      sync.RWMutex
	watches map[string][]chan *config.Change
}

// Options 配置选项
type Options struct {
	Prefix string        // 配置前缀
	TTL    time.Duration // 配置TTL
	Parser config.Parser // 配置解析器
}

// Option 定义配置函数类型
type Option func(*Options)

// WithPrefix 设置配置前缀
func WithPrefix(prefix string) Option {
	return func(o *Options) {
		o.Prefix = prefix
	}
}

// WithTTL 设置配置TTL
func WithTTL(ttl time.Duration) Option {
	return func(o *Options) {
		o.TTL = ttl
	}
}

// WithParser 设置配置解析器
func WithParser(parser config.Parser) Option {
	return func(o *Options) {
		o.Parser = parser
	}
}

// New 创建ETCD配置中心
func New(client *clientv3.Client, opts ...Option) (*Config, error) {
	options := &Options{
		Prefix: "/config",
		TTL:    time.Hour,
		Parser: &defaultParser{},
	}
	for _, o := range opts {
		o(options)
	}

	c := &Config{
		client:  client,
		prefix:  options.Prefix,
		values:  make(map[string]config.Value),
		parser:  options.Parser,
		watches: make(map[string][]chan *config.Change),
	}

	return c, nil
}

// Load 加载配置
func (c *Config) Load() error {
	resp, err := c.client.Get(context.Background(), c.prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 清空现有配置
	c.values = make(map[string]config.Value)

	// 解析配置
	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		var value interface{}
		if err := json.Unmarshal(kv.Value, &value); err != nil {
			return err
		}

		values, err := c.parser.Parse(map[string]interface{}{key: value})
		if err != nil {
			return err
		}
		for k, v := range values {
			c.values[k] = v
		}
	}

	return nil
}

// Get 获取配置值
func (c *Config) Get(key string) (config.Value, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if value, ok := c.values[key]; ok {
		return value, nil
	}
	return nil, config.ErrNotFound
}

// Watch 监听配置变更
func (c *Config) Watch(ctx context.Context, key string) (config.Watcher, error) {
	// 创建监听通道
	ch := make(chan *config.Change, 1)

	c.mu.Lock()
	if _, ok := c.watches[key]; !ok {
		c.watches[key] = make([]chan *config.Change, 0)
	}
	c.watches[key] = append(c.watches[key], ch)
	c.mu.Unlock()

	// 启动监听
	watchKey := path.Join(c.prefix, key)
	go c.watch(ctx, watchKey, ch)

	return &watcher{
		key:    key,
		config: c,
		ch:     ch,
	}, nil
}

// watch 监听配置变更
func (c *Config) watch(ctx context.Context, key string, ch chan *config.Change) {
	wch := c.client.Watch(ctx, key)
	for {
		select {
		case <-ctx.Done():
			return
		case wresp := <-wch:
			for _, ev := range wresp.Events {
				var value interface{}
				if err := json.Unmarshal(ev.Kv.Value, &value); err != nil {
					continue
				}

				values, err := c.parser.Parse(map[string]interface{}{key: value})
				if err != nil {
					continue
				}

				for k, v := range values {
					change := &config.Change{
						Key:       k,
						Value:     v,
						PreValue:  c.values[k],
						Timestamp: time.Now(),
					}

					switch ev.Type {
					case clientv3.EventTypePut:
						if _, ok := c.values[k]; ok {
							change.Type = config.Update
						} else {
							change.Type = config.Create
						}
					case clientv3.EventTypeDelete:
						change.Type = config.Delete
					}

					c.mu.Lock()
					c.values[k] = v
					c.mu.Unlock()

					ch <- change
				}
			}
		}
	}
}

// Close 关闭配置中心
func (c *Config) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 关闭所有监听
	for _, chs := range c.watches {
		for _, ch := range chs {
			close(ch)
		}
	}
	c.watches = make(map[string][]chan *config.Change)

	return nil
}

// watcher 实现配置监听器
type watcher struct {
	key    string
	config *Config
	ch     chan *config.Change
}

// Next 获取下一个配置变更
func (w *watcher) Next() (*config.Change, error) {
	change, ok := <-w.ch
	if !ok {
		return nil, config.ErrNotFound
	}
	return change, nil
}

// Stop 停止监听
func (w *watcher) Stop() error {
	w.config.mu.Lock()
	defer w.config.mu.Unlock()

	// 从监听列表中移除
	if chs, ok := w.config.watches[w.key]; ok {
		for i, ch := range chs {
			if ch == w.ch {
				w.config.watches[w.key] = append(chs[:i], chs[i+1:]...)
				close(ch)
				break
			}
		}
	}

	return nil
}
