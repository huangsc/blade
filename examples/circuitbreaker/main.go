package main

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/huangsc/blade/circuitbreaker"
)

var (
	// ErrTimeout 模拟超时错误
	ErrTimeout = errors.New("timeout")

	// ErrInternalError 模拟内部错误
	ErrInternalError = errors.New("internal error")
)

func main() {
	// 创建熔断器
	breaker := circuitbreaker.NewBreaker("example",
		circuitbreaker.WithMaxRequests(3),         // 半开状态下最多允许3个请求
		circuitbreaker.WithInterval(time.Minute),  // 统计周期为1分钟
		circuitbreaker.WithTimeout(time.Second*5), // 熔断超时时间为5秒
		circuitbreaker.WithReadyToTrip(func(counts circuitbreaker.Counts) bool {
			// 连续5次失败触发熔断
			return counts.ConsecutiveFailures >= 5
		}),
		circuitbreaker.WithOnStateChange(func(name string, from circuitbreaker.State, to circuitbreaker.State) {
			log.Printf("Circuit breaker %s changed from %v to %v\n", name, from, to)
		}),
	)

	// 测试熔断器
	testCircuitBreaker(breaker)
}

func testCircuitBreaker(breaker circuitbreaker.Breaker) {
	var wg sync.WaitGroup
	ctx := context.Background()

	// 测试正常请求
	log.Println("\nTesting normal requests...")
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := breaker.Execute(ctx, func() error {
				// 模拟正常请求
				time.Sleep(time.Millisecond * 100)
				return nil
			})
			if err != nil {
				log.Printf("Request %d failed: %v", i, err)
				return
			}
			log.Printf("Request %d succeeded", i)
		}(i)
	}
	wg.Wait()

	// 测试失败请求
	log.Println("\nTesting failing requests...")
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := breaker.Execute(ctx, func() error {
				// 模拟失败请求
				return ErrInternalError
			})
			if err != nil {
				log.Printf("Request %d failed: %v", i, err)
				return
			}
			log.Printf("Request %d succeeded", i)
		}(i)
	}
	wg.Wait()

	// 等待熔断器超时
	log.Println("\nWaiting for timeout...")
	time.Sleep(time.Second * 5)

	// 测试半开状态
	log.Println("\nTesting half-open state...")
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := breaker.Execute(ctx, func() error {
				// 随机成功或失败
				if rand.Float64() < 0.5 {
					return ErrInternalError
				}
				return nil
			})
			if err != nil {
				log.Printf("Request %d failed: %v", i, err)
				return
			}
			log.Printf("Request %d succeeded", i)
		}(i)
	}
	wg.Wait()

	// 测试两阶段执行
	log.Println("\nTesting two-step execution...")
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			// 第一阶段: 检查是否允许请求
			done, err := breaker.Allow()
			if err != nil {
				log.Printf("Request %d rejected: %v", i, err)
				return
			}

			// 执行请求
			success := rand.Float64() >= 0.5
			if success {
				log.Printf("Request %d succeeded", i)
			} else {
				log.Printf("Request %d failed", i)
			}

			// 第二阶段: 报告结果
			done(success)
		}(i)
	}
	wg.Wait()

	// 打印最终统计信息
	counts := breaker.Counts()
	log.Printf("\nFinal counts: %+v\n", counts)
	log.Printf("Final state: %v\n", breaker.State())
}

// 模拟不稳定的服务
func unstableService() error {
	// 随机延迟
	delay := time.Duration(rand.Int63n(100)) * time.Millisecond
	time.Sleep(delay)

	// 随机错误
	switch rand.Intn(3) {
	case 0:
		return nil
	case 1:
		return ErrTimeout
	default:
		return ErrInternalError
	}
}
