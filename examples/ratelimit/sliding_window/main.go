package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/huangsc/blade/ratelimit"
)

func main() {
	// 创建滑动窗口限流器
	limiter := ratelimit.NewSlidingWindow(
		ratelimit.WithRate(10),                        // 每秒允许10个请求
		ratelimit.WithWindow(time.Second),             // 窗口大小为1秒
		ratelimit.WithPrecision(time.Millisecond*100), // 精度为100ms
	)

	// 测试限流器
	testSlidingWindow(limiter)
}

func testSlidingWindow(limiter ratelimit.Limiter) {
	var wg sync.WaitGroup
	ctx := context.Background()

	// 测试正常流量
	log.Println("Testing normal traffic...")
	for i := 0; i < 5; i++ {
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

	// 测试突发流量
	log.Println("\nTesting burst traffic...")
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

	// 测试等待模式
	log.Println("\nTesting wait mode...")
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			start := time.Now()
			err := limiter.Wait(ctx, "test-key", 1)
			if err != nil {
				log.Printf("Request %d failed: %v", i, err)
				return
			}
			elapsed := time.Since(start)
			log.Printf("Request %d completed after %v", i, elapsed)
		}(i)
	}
	wg.Wait()

	// 测试预约模式
	log.Println("\nTesting reserve mode...")
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			reservation, err := limiter.Reserve(ctx, "test-key", 1)
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

	// 测试不同的时间窗口
	log.Println("\nTesting different time windows...")
	keys := []string{"1s-window", "5s-window", "10s-window"}
	windows := []time.Duration{time.Second, 5 * time.Second, 10 * time.Second}

	for i, key := range keys {
		log.Printf("\nTesting %s...", key)
		start := time.Now()
		count := 0
		for time.Since(start) < windows[i] {
			allowed, _ := limiter.Allow(ctx, key, 1)
			if allowed {
				count++
			}
			time.Sleep(time.Millisecond * 50)
		}
		log.Printf("Allowed %d requests in %v", count, windows[i])
	}

	// 测试精度影响
	log.Println("\nTesting precision impact...")
	precisions := []time.Duration{
		time.Millisecond * 100,
		time.Millisecond * 500,
		time.Second,
	}
	for _, precision := range precisions {
		limiter = ratelimit.NewSlidingWindow(
			ratelimit.WithRate(10),
			ratelimit.WithWindow(time.Second),
			ratelimit.WithPrecision(precision),
		)

		log.Printf("\nTesting with precision %v...", precision)
		start := time.Now()
		count := 0
		for time.Since(start) < time.Second {
			allowed, _ := limiter.Allow(ctx, "precision-test", 1)
			if allowed {
				count++
			}
			time.Sleep(time.Millisecond * 50)
		}
		log.Printf("Allowed %d requests with precision %v", count, precision)
	}
}
