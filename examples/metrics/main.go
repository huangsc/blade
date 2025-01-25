package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/huangsc/blade/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// 创建指标收集器
	m := metrics.NewPrometheusMetrics(
		metrics.WithNamespace("knife"),
		metrics.WithSubsystem("example"),
		metrics.WithConstLabels(metrics.Labels{
			"service": "example",
			"version": "v1.0.0",
		}),
	)

	// 创建计数器
	requestCounter := m.Counter("http_requests_total", metrics.Labels{
		"method": "GET",
		"path":   "/hello",
	})

	// 创建仪表盘
	cpuUsage := m.Gauge("cpu_usage", metrics.Labels{
		"core": "0",
	})

	// 创建直方图
	requestDuration := m.Histogram("request_duration_seconds", metrics.Labels{
		"handler": "hello",
	})

	// 创建摘要
	requestLatency := m.Summary("request_latency_seconds", metrics.Labels{
		"handler": "hello",
	})

	// 模拟请求处理
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 增加请求计数
		requestCounter.Inc()

		// 模拟 CPU 使用率
		cpuUsage.Set(rand.Float64() * 100)

		// 模拟处理延迟
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

		// 记录请求持续时间
		duration := time.Since(start).Seconds()
		requestDuration.Observe(duration)
		requestLatency.Observe(duration)

		fmt.Fprintf(w, "Hello, World!")
	})

	// 暴露 Prometheus 指标
	http.Handle("/metrics", promhttp.Handler())

	// 启动服务器
	fmt.Println("Server is running on http://localhost:8080")
	fmt.Println("Metrics are exposed on http://localhost:8080/metrics")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
