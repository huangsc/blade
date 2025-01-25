package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/huangsc/blade/cache"
)

func main() {
	// 创建Redis缓存实例
	c := cache.NewRedisCache(cache.RedisOptions{
		Options: cache.Options{
			TTL: time.Second * 5, // 默认5秒过期
			OnEvicted: func(key string, value interface{}) {
				log.Printf("缓存项被移除: key=%s, value=%v\n", key, value)
			},
		},
		Addr:         "localhost:6379", // Redis地址
		Password:     "",               // Redis密码
		DB:           0,                // 使用默认数据库
		PoolSize:     10,               // 连接池大小
		MinIdleConns: 5,                // 最小空闲连接
		KeyPrefix:    "cache:",         // 键前缀
	})
	defer c.Close()

	ctx := context.Background()

	// 测试设置和获取
	log.Println("\n测试设置和获取...")
	err := c.Set(ctx, "key1", "value1", 0)
	if err != nil {
		log.Fatal(err)
	}

	value, err := c.Get(ctx, "key1")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("key1 = %v\n", value)

	// 测试不存在的键
	log.Println("\n测试不存在的键...")
	value, err = c.Get(ctx, "not_exist")
	if err != nil {
		log.Printf("预期的错误: %v\n", err)
	}

	// 测试过期
	log.Println("\n测试过期...")
	err = c.Set(ctx, "key2", "value2", time.Second)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Second * 2)
	value, err = c.Get(ctx, "key2")
	if err != nil {
		log.Printf("预期的过期错误: %v\n", err)
	}

	// 测试删除
	log.Println("\n测试删除...")
	err = c.Set(ctx, "key3", "value3", 0)
	if err != nil {
		log.Fatal(err)
	}

	err = c.Delete(ctx, "key3")
	if err != nil {
		log.Fatal(err)
	}

	value, err = c.Get(ctx, "key3")
	if err != nil {
		log.Printf("预期的错误: %v\n", err)
	}

	// 测试清空
	log.Println("\n测试清空...")
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("key%d", i)
		err = c.Set(ctx, key, i, 0)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("当前缓存数量: %d\n", c.Len(ctx))

	err = c.Clear(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("清空后缓存数量: %d\n", c.Len(ctx))

	// 测试结构体
	log.Println("\n测试结构体...")
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	user := User{
		ID:   1,
		Name: "张三",
	}

	err = c.Set(ctx, "user:1", user, time.Minute)
	if err != nil {
		log.Fatal(err)
	}

	value, err = c.Get(ctx, "user:1")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("user:1 = %v\n", value)
}
