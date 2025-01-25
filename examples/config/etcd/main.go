package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/huangsc/blade/config/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Config 应用配置
type Config struct {
	Server struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"server"`
	Database struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Name     string `json:"name"`
	} `json:"database"`
	Redis struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	} `json:"redis"`
}

func main() {
	// 创建 ETCD 客户端
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		log.Fatalf("创建 ETCD 客户端失败: %v", err)
	}
	defer client.Close()

	// 创建配置中心
	cfg, err := etcd.New(client,
		etcd.WithPrefix("/myapp/config"),
		etcd.WithTTL(time.Hour),
	)
	if err != nil {
		log.Fatalf("创建配置中心失败: %v", err)
	}

	// 加载配置
	if err := cfg.Load(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 获取配置
	serverValue, err := cfg.Get("server")
	if err != nil {
		log.Printf("获取服务器配置失败: %v", err)
	} else {
		var server struct {
			Host string `json:"host"`
			Port int    `json:"port"`
		}
		if err := serverValue.Scan(&server); err != nil {
			log.Printf("解析服务器配置失败: %v", err)
		} else {
			log.Printf("服务器配置: %+v", server)
		}
	}

	// 监听配置变更
	ctx, cancel := context.WithCancel(context.Background())
	watcher, err := cfg.Watch(ctx, "database")
	if err != nil {
		log.Printf("监听数据库配置失败: %v", err)
	} else {
		go func() {
			for {
				change, err := watcher.Next()
				if err != nil {
					log.Printf("获取配置变更失败: %v", err)
					continue
				}

				var database struct {
					Host     string `json:"host"`
					Port     int    `json:"port"`
					User     string `json:"user"`
					Password string `json:"password"`
					Name     string `json:"name"`
				}
				if err := change.Value.Scan(&database); err != nil {
					log.Printf("解析数据库配置失败: %v", err)
					continue
				}

				log.Printf("数据库配置已更新: %+v", database)
			}
		}()
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭
	log.Println("正在关闭配置中心...")
	cancel()
	if err := watcher.Stop(); err != nil {
		log.Printf("停止配置监听失败: %v", err)
	}
	log.Println("配置中心已关闭")
}
