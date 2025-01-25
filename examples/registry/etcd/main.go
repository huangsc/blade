package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/huangsc/blade/registry"
	"github.com/huangsc/blade/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	// 创建 etcd 客户端
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		log.Fatalf("创建 etcd 客户端失败: %v", err)
	}
	defer client.Close()

	// 创建注册中心
	r, err := etcd.New(client,
		etcd.WithPrefix("/my-services"),
		etcd.WithTTL(time.Second*30),
	)
	if err != nil {
		log.Fatalf("创建注册中心失败: %v", err)
	}

	// 创建服务实例
	service := &registry.ServiceInstance{
		ID:      "user-service-1",
		Name:    "user-service",
		Version: "v1.0.0",
		Metadata: map[string]string{
			"region": "cn-shanghai",
			"zone":   "cn-shanghai-a",
		},
		Endpoints: []string{
			"grpc://localhost:9000",
			"http://localhost:8080",
		},
	}

	// 注册服务
	ctx := context.Background()
	if err := r.Register(ctx, service); err != nil {
		log.Fatalf("注册服务失败: %v", err)
	}
	log.Printf("服务注册成功: %s", service.ID)

	// 查询服务
	services, err := r.GetService(ctx, service.Name)
	if err != nil {
		log.Printf("查询服务失败: %v", err)
	} else {
		log.Printf("发现服务实例数量: %d", len(services))
		for _, svc := range services {
			log.Printf("服务实例: ID=%s, 端点=%v", svc.ID, svc.Endpoints)
		}
	}

	// 监听服务变更
	watch, err := r.Watch(ctx, service.Name)
	if err != nil {
		log.Printf("监听服务失败: %v", err)
	} else {
		go func() {
			for {
				services, err := watch.Next()
				if err != nil {
					log.Printf("监听服务变更失败: %v", err)
					continue
				}
				log.Printf("服务列表更新: %d 个实例", len(services))
				for _, svc := range services {
					log.Printf("服务实例: ID=%s, 端点=%v", svc.ID, svc.Endpoints)
				}
			}
		}()
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 注销服务
	log.Println("正在注销服务...")
	if err := r.Deregister(ctx, service); err != nil {
		log.Printf("注销服务失败: %v", err)
	} else {
		log.Println("服务注销成功")
	}
}
