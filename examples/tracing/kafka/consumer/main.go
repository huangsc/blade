package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/huangsc/blade/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

// Message 消息结构
type Message struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	// 创建追踪器
	tracer, err := tracing.NewOTelTracer(
		tracing.WithServiceName("kafka-consumer"),
		tracing.WithServiceVersion("v1.0.0"),
		tracing.WithEnvironment("development"),
		tracing.WithEndpoint("localhost:4317"),
		tracing.WithSampler(1.0),
	)
	if err != nil {
		log.Fatalf("Failed to create tracer: %v", err)
	}

	// 创建 Kafka 消费者
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "example-group",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	// 订阅主题
	topic := "example-topic"
	err = consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %v", err)
	}

	// 处理信号
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// 消费消息
	run := true
	for run {
		select {
		case sig := <-sigchan:
			log.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			ev := consumer.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				// 从消息头提取追踪上下文
				carrier := propagation.MapCarrier{}
				for _, header := range e.Headers {
					carrier.Set(header.Key, string(header.Value))
				}
				ctx := tracer.Extract(context.Background(), carrier)

				// 创建消费者 Span
				ctx, span := tracer.Start(ctx, "kafka.consume",
					tracing.WithSpanKind(tracing.SpanKindConsumer),
					tracing.WithSpanAttributes(
						attribute.String("messaging.system", "kafka"),
						attribute.String("messaging.destination", *e.TopicPartition.Topic),
						attribute.String("messaging.destination_kind", "topic"),
						attribute.String("messaging.consumer_group", "example-group"),
						attribute.Int("messaging.kafka.partition", int(e.TopicPartition.Partition)),
						attribute.Int64("messaging.kafka.offset", int64(e.TopicPartition.Offset)),
					),
				)

				// 处理消息
				if err := processMessage(ctx, tracer, e.Value); err != nil {
					span.SetError(err)
					log.Printf("Failed to process message: %v", err)
				}

				span.End()

			case kafka.Error:
				log.Printf("Error: %v\n", e)
				if e.Code() == kafka.ErrAllBrokersDown {
					run = false
				}
			}
		}
	}
}

func processMessage(ctx context.Context, tracer tracing.Tracer, value []byte) error {
	// 创建处理 Span
	ctx, span := tracer.Start(ctx, "process_message",
		tracing.WithSpanKind(tracing.SpanKindInternal),
	)
	defer span.End()

	// 添加处理开始事件
	span.AddEvent("processing.start")

	// 反序列化消息
	var msg Message
	if err := json.Unmarshal(value, &msg); err != nil {
		return err
	}

	// 添加消息属性
	span.SetAttributes(
		attribute.String("message.id", msg.ID),
		attribute.String("message.content", msg.Content),
		attribute.String("message.timestamp", msg.Timestamp.Format(time.RFC3339)),
	)

	// 模拟处理延迟
	time.Sleep(time.Duration(100) * time.Millisecond)

	// 添加处理结束事件
	span.AddEvent("processing.end",
		attribute.String("status", "success"),
	)

	log.Printf("Processed message: %s - %s", msg.ID, msg.Content)
	return nil
}
