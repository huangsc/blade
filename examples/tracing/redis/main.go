package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/huangsc/blade/tracing"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
)

// User 用户结构
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	// 创建追踪器
	tracer, err := tracing.NewOTelTracer(
		tracing.WithServiceName("redis-example"),
		tracing.WithServiceVersion("v1.0.0"),
		tracing.WithEnvironment("development"),
		tracing.WithEndpoint("localhost:4317"),
		tracing.WithSampler(1.0),
	)
	if err != nil {
		log.Fatalf("Failed to create tracer: %v", err)
	}

	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer rdb.Close()

	// 创建用户
	user := User{
		ID:        "user-1",
		Name:      "Alice",
		Email:     "alice@example.com",
		CreatedAt: time.Now(),
	}

	// 设置缓存
	if err := setCache(context.Background(), tracer, rdb, user); err != nil {
		log.Fatalf("Failed to set cache: %v", err)
	}

	// 获取缓存
	cachedUser, err := getCache(context.Background(), tracer, rdb, user.ID)
	if err != nil {
		log.Fatalf("Failed to get cache: %v", err)
	}
	log.Printf("Got user from cache: %+v", cachedUser)

	// 删除缓存
	if err := deleteCache(context.Background(), tracer, rdb, user.ID); err != nil {
		log.Fatalf("Failed to delete cache: %v", err)
	}

	// 使用管道
	if err := usePipeline(context.Background(), tracer, rdb); err != nil {
		log.Fatalf("Failed to use pipeline: %v", err)
	}

	// 使用事务
	if err := useTransaction(context.Background(), tracer, rdb); err != nil {
		log.Fatalf("Failed to use transaction: %v", err)
	}
}

func setCache(ctx context.Context, tracer tracing.Tracer, rdb *redis.Client, user User) error {
	// 创建 Span
	ctx, span := tracer.Start(ctx, "redis.set",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithSpanAttributes(
			attribute.String("db.system", "redis"),
			attribute.String("db.operation", "SET"),
			attribute.String("cache.key", fmt.Sprintf("user:%s", user.ID)),
		),
	)
	defer span.End()

	// 序列化用户数据
	value, err := json.Marshal(user)
	if err != nil {
		span.SetError(err)
		return err
	}

	// 设置缓存
	key := fmt.Sprintf("user:%s", user.ID)
	err = rdb.Set(ctx, key, value, 24*time.Hour).Err()
	if err != nil {
		span.SetError(err)
		return err
	}

	span.SetAttributes(
		attribute.Int("cache.value.size", len(value)),
		attribute.String("cache.ttl", "24h"),
	)

	return nil
}

func getCache(ctx context.Context, tracer tracing.Tracer, rdb *redis.Client, userID string) (*User, error) {
	// 创建 Span
	ctx, span := tracer.Start(ctx, "redis.get",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithSpanAttributes(
			attribute.String("db.system", "redis"),
			attribute.String("db.operation", "GET"),
			attribute.String("cache.key", fmt.Sprintf("user:%s", userID)),
		),
	)
	defer span.End()

	// 获取缓存
	key := fmt.Sprintf("user:%s", userID)
	value, err := rdb.Get(ctx, key).Bytes()
	if err != nil {
		span.SetError(err)
		return nil, err
	}

	// 反序列化用户数据
	var user User
	if err := json.Unmarshal(value, &user); err != nil {
		span.SetError(err)
		return nil, err
	}

	span.SetAttributes(
		attribute.Int("cache.value.size", len(value)),
		attribute.Bool("cache.hit", true),
	)

	return &user, nil
}

func deleteCache(ctx context.Context, tracer tracing.Tracer, rdb *redis.Client, userID string) error {
	// 创建 Span
	ctx, span := tracer.Start(ctx, "redis.delete",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithSpanAttributes(
			attribute.String("db.system", "redis"),
			attribute.String("db.operation", "DEL"),
			attribute.String("cache.key", fmt.Sprintf("user:%s", userID)),
		),
	)
	defer span.End()

	// 删除缓存
	key := fmt.Sprintf("user:%s", userID)
	err := rdb.Del(ctx, key).Err()
	if err != nil {
		span.SetError(err)
		return err
	}

	return nil
}

func usePipeline(ctx context.Context, tracer tracing.Tracer, rdb *redis.Client) error {
	// 创建 Span
	ctx, span := tracer.Start(ctx, "redis.pipeline",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithSpanAttributes(
			attribute.String("db.system", "redis"),
			attribute.String("db.operation", "PIPELINE"),
		),
	)
	defer span.End()

	// 创建管道
	pipe := rdb.Pipeline()

	// 添加命令
	incr := pipe.Incr(ctx, "pipeline_counter")
	pipe.Expire(ctx, "pipeline_counter", time.Hour)

	// 执行管道
	_, err := pipe.Exec(ctx)
	if err != nil {
		span.SetError(err)
		return err
	}

	// 获取计数器值
	counter, err := incr.Result()
	if err != nil {
		span.SetError(err)
		return err
	}

	span.SetAttributes(
		attribute.Int64("pipeline.counter", counter),
		attribute.Int("pipeline.commands", 2),
	)

	return nil
}

func useTransaction(ctx context.Context, tracer tracing.Tracer, rdb *redis.Client) error {
	// 创建 Span
	ctx, span := tracer.Start(ctx, "redis.transaction",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithSpanAttributes(
			attribute.String("db.system", "redis"),
			attribute.String("db.operation", "MULTI"),
		),
	)
	defer span.End()

	// 执行事务
	err := rdb.Watch(ctx, func(tx *redis.Tx) error {
		// 获取计数器值
		n, err := tx.Get(ctx, "tx_counter").Int()
		if err != nil && err != redis.Nil {
			return err
		}

		// 事务操作
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, "tx_counter", n+1, time.Hour)
			return nil
		})
		return err
	}, "tx_counter")

	if err != nil {
		span.SetError(err)
		return err
	}

	span.SetAttributes(
		attribute.String("transaction.key", "tx_counter"),
		attribute.String("transaction.type", "WATCH"),
	)

	return nil
}
