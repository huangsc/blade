package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/huangsc/blade/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

func main() {
	// 创建追踪器
	tracer, err := tracing.NewOTelTracer(
		tracing.WithServiceName("example"),
		tracing.WithServiceVersion("v1.0.0"),
		tracing.WithEnvironment("development"),
		tracing.WithEndpoint("localhost:4317"),
		tracing.WithSampler(1.0),
	)
	if err != nil {
		log.Fatalf("Failed to create tracer: %v", err)
	}

	// 创建 HTTP 处理函数
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		// 从请求中提取上下文
		carrier := propagation.HeaderCarrier(r.Header)
		ctx := tracer.Extract(r.Context(), carrier)

		// 创建根 Span
		ctx, span := tracer.Start(ctx, "hello",
			tracing.WithSpanKind(tracing.SpanKindServer),
			tracing.WithSpanAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.String()),
			),
		)
		defer span.End()

		// 模拟一些处理
		if err := processRequest(ctx, tracer); err != nil {
			span.SetError(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 注入上下文到响应头
		carrier = propagation.HeaderCarrier(w.Header())
		if err := tracer.Inject(ctx, carrier); err != nil {
			span.SetError(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Hello, World!")
	})

	// 启动服务器
	fmt.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func processRequest(ctx context.Context, tracer tracing.Tracer) error {
	// 创建子 Span
	ctx, span := tracer.Start(ctx, "process_request",
		tracing.WithSpanKind(tracing.SpanKindInternal),
	)
	defer span.End()

	// 添加一些事件
	span.AddEvent("processing.start",
		attribute.String("processor", "main"),
		attribute.Int("queue_length", 10),
	)

	// 模拟一些处理延迟
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

	// 模拟数据库调用
	if err := queryDatabase(ctx, tracer); err != nil {
		return fmt.Errorf("database error: %v", err)
	}

	// 添加处理结果
	span.AddEvent("processing.end",
		attribute.String("status", "success"),
		attribute.Int("items_processed", 5),
	)

	return nil
}

func queryDatabase(ctx context.Context, tracer tracing.Tracer) error {
	// 创建子 Span
	ctx, span := tracer.Start(ctx, "query_database",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithSpanAttributes(
			attribute.String("db.system", "mysql"),
			attribute.String("db.name", "example"),
			attribute.String("db.operation", "select"),
		),
	)
	defer span.End()

	// 模拟数据库查询延迟
	time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

	// 模拟随机错误
	if rand.Float64() < 0.1 {
		err := fmt.Errorf("database connection failed")
		span.SetError(err)
		return err
	}

	return nil
}
