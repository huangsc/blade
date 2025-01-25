package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/huangsc/blade/cache"
)

func main() {
	// 创建内存缓存实例
	c := cache.NewMemoryCache(
		cache.WithTTL(time.Second*5), // 默认5秒过期
		cache.WithMaxEntries(100),    // 最多存储100个条目
		cache.WithOnEvicted(func(key string, value interface{}) {
			log.Printf("缓存项被移除: key=%s, value=%v\n", key, value)
		}),
	)
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

	// 测试最大条目限制
	log.Println("\n测试最大条目限制...")
	for i := 0; i < 200; i++ {
		key := fmt.Sprintf("key%d", i)
		err = c.Set(ctx, key, i, 0)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("最终缓存数量: %d\n", c.Len(ctx))
}
