package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/huangsc/blade/ratelimit"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 创建内存限流器
	memLimiter := ratelimit.NewTokenBucket(
		ratelimit.WithRate(10),   // 每秒生成10个令牌
		ratelimit.WithBurst(100), // 最多存储100个令牌
	)

	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer rdb.Close()

	// 创建分布式限流器
	redisLimiter := ratelimit.NewRedisLimiter(rdb,
		ratelimit.WithRate(10),       // 每秒生成10个令牌
		ratelimit.WithBurst(100),     // 最多存储100个令牌
		ratelimit.WithTTL(time.Hour), // 键的过期时间为1小时
	)

	// 测试内存限流器
	log.Println("Testing memory limiter...")
	testLimiter(memLimiter)

	// 测试分布式限流器
	log.Println("\nTesting Redis limiter...")
	testLimiter(redisLimiter)
}

func testLimiter(limiter ratelimit.Limiter) {
	var wg sync.WaitGroup
	ctx := context.Background()

	// 测试 Allow 方法
	log.Println("Testing Allow method...")
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			allowed, err := limiter.Allow(ctx, "test-key", 1)
			if err != nil {
				log.Printf("Request %d failed: %v", i, err)
				return
			}
			if allowed {
				log.Printf("Request %d allowed", i)
			} else {
				log.Printf("Request %d rejected", i)
			}
		}(i)
	}
	wg.Wait()

	// 测试 Wait 方法
	log.Println("\nTesting Wait method...")
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			start := time.Now()
			err := limiter.Wait(ctx, "test-key", 10) // 请求10个令牌
			if err != nil {
				log.Printf("Request %d failed: %v", i, err)
				return
			}
			elapsed := time.Since(start)
			log.Printf("Request %d completed after %v", i, elapsed)
		}(i)
	}
	wg.Wait()

	// 测试 Reserve 方法
	log.Println("\nTesting Reserve method...")
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			reservation, err := limiter.Reserve(ctx, "test-key", 20) // 预约20个令牌
			if err != nil {
				log.Printf("Request %d failed: %v", i, err)
				return
			}
			if !reservation.OK {
				log.Printf("Request %d rejected", i)
				return
			}
			delay := reservation.Delay()
			log.Printf("Request %d will be allowed after %v", i, delay)
		}(i)
	}
	wg.Wait()

	// 测试取消上下文
	log.Println("\nTesting context cancellation...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := limiter.Wait(ctx, "test-key", 1000) // 请求大量令牌
	if err != nil {
		log.Printf("Wait cancelled: %v", err)
	}
}
