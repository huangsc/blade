package cache

import (
	"context"
	"sync"
	"time"
)

// memory 内存缓存实现
type memory struct {
	// mutex 互斥锁
	mutex sync.RWMutex

	// items 缓存项
	items map[string]*Item

	// options 配置选项
	options Options

	// janitor 清理器
	janitor *janitor
}

// NewMemoryCache 创建内存缓存
func NewMemoryCache(opts ...Option) Cache {
	options := Options{
		TTL:        time.Hour,
		MaxEntries: 10000,
	}

	for _, opt := range opts {
		opt(&options)
	}

	m := &memory{
		items:   make(map[string]*Item),
		options: options,
	}

	// 启动清理器
	m.janitor = newJanitor(m)
	go m.janitor.run()

	return m
}

// Get 获取缓存值
func (m *memory) Get(ctx context.Context, key string) (interface{}, error) {
	m.mutex.RLock()
	item, ok := m.items[key]
	m.mutex.RUnlock()

	if !ok {
		return nil, ErrKeyNotFound
	}

	if item.ExpireAt.Before(time.Now()) {
		m.Delete(ctx, key)
		return nil, ErrKeyExpired
	}

	return item.Value, nil
}

// Set 设置缓存值
func (m *memory) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = m.options.TTL
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查是否超过最大条目数
	if m.options.MaxEntries > 0 && len(m.items) >= m.options.MaxEntries {
		m.evict()
	}

	m.items[key] = &Item{
		Key:       key,
		Value:     value,
		ExpireAt:  time.Now().Add(ttl),
		CreatedAt: time.Now(),
	}

	return nil
}

// Delete 删除缓存值
func (m *memory) Delete(ctx context.Context, key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if item, ok := m.items[key]; ok && m.options.OnEvicted != nil {
		m.options.OnEvicted(key, item.Value)
	}

	delete(m.items, key)
	return nil
}

// Clear 清空缓存
func (m *memory) Clear(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.options.OnEvicted != nil {
		for k, v := range m.items {
			m.options.OnEvicted(k, v.Value)
		}
	}

	m.items = make(map[string]*Item)
	return nil
}

// Keys 获取所有键
func (m *memory) Keys(ctx context.Context) ([]string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	keys := make([]string, 0, len(m.items))
	for k := range m.items {
		keys = append(keys, k)
	}

	return keys, nil
}

// Len 获取缓存项数量
func (m *memory) Len(ctx context.Context) int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return len(m.items)
}

// Close 关闭缓存
func (m *memory) Close() error {
	m.janitor.stop()
	return m.Clear(context.Background())
}

// evict 移除过期或最旧的条目
func (m *memory) evict() {
	now := time.Now()
	oldest := now
	var oldestKey string

	// 先尝试移除过期项
	for k, v := range m.items {
		if v.ExpireAt.Before(now) {
			if m.options.OnEvicted != nil {
				m.options.OnEvicted(k, v.Value)
			}
			delete(m.items, k)
			return
		}
		if v.CreatedAt.Before(oldest) {
			oldest = v.CreatedAt
			oldestKey = k
		}
	}

	// 如果没有过期项，移除最旧的项
	if oldestKey != "" {
		if m.options.OnEvicted != nil {
			m.options.OnEvicted(oldestKey, m.items[oldestKey].Value)
		}
		delete(m.items, oldestKey)
	}
}

// janitor 清理器
type janitor struct {
	// interval 清理间隔
	interval time.Duration

	// stopCh 停止信号
	stopCh chan bool

	// cache 缓存实例
	cache *memory
}

// newJanitor 创建清理器
func newJanitor(cache *memory) *janitor {
	return &janitor{
		interval: time.Minute,
		stopCh:   make(chan bool),
		cache:    cache,
	}
}

// run 运行清理器
func (j *janitor) run() {
	ticker := time.NewTicker(j.interval)
	for {
		select {
		case <-ticker.C:
			j.clean()
		case <-j.stopCh:
			ticker.Stop()
			return
		}
	}
}

// clean 清理过期项
func (j *janitor) clean() {
	now := time.Now()
	j.cache.mutex.Lock()
	for k, v := range j.cache.items {
		if v.ExpireAt.Before(now) {
			if j.cache.options.OnEvicted != nil {
				j.cache.options.OnEvicted(k, v.Value)
			}
			delete(j.cache.items, k)
		}
	}
	j.cache.mutex.Unlock()
}

// stop 停止清理器
func (j *janitor) stop() {
	j.stopCh <- true
}
