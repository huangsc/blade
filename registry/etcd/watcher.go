package etcd

import (
	"context"
	"encoding/json"

	"github.com/huangsc/blade/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// watcher 实现了 registry.Watcher 接口
type watcher struct {
	key    string
	client *clientv3.Client
	ctx    context.Context
	cancel context.CancelFunc
	wch    clientv3.WatchChan
}

// newWatcher 创建新的 watcher
func newWatcher(ctx context.Context, key string, client *clientv3.Client) registry.Watcher {
	ctx, cancel := context.WithCancel(ctx)
	w := &watcher{
		key:    key,
		client: client,
		ctx:    ctx,
		cancel: cancel,
		wch:    client.Watch(ctx, key, clientv3.WithPrefix()),
	}
	return w
}

// Next 实现 registry.Watcher 接口
func (w *watcher) Next() ([]*registry.ServiceInstance, error) {
	select {
	case <-w.ctx.Done():
		return nil, w.ctx.Err()
	case wresp := <-w.wch:
		if wresp.Err() != nil {
			return nil, wresp.Err()
		}
		var items []*registry.ServiceInstance
		for _, ev := range wresp.Events {
			si := &registry.ServiceInstance{}
			if ev.Type == clientv3.EventTypeDelete {
				continue
			}
			if err := json.Unmarshal(ev.Kv.Value, si); err != nil {
				return nil, err
			}
			items = append(items, si)
		}
		return items, nil
	}
}

// Stop 实现 registry.Watcher 接口
func (w *watcher) Stop() error {
	w.cancel()
	return nil
}
