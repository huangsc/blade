package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
		tracing.WithServiceName("kafka-producer"),
		tracing.WithServiceVersion("v1.0.0"),
		tracing.WithEnvironment("development"),
		tracing.WithEndpoint("localhost:4317"),
		tracing.WithSampler(1.0),
	)
	if err != nil {
		log.Fatalf("Failed to create tracer: %v", err)
	}

	// 创建 Kafka 生产者
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"client.id":         "kafka-producer-example",
		"acks":              "all",
	})
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	// 处理投递报告
	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
				} else {
					log.Printf("Successfully delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	topic := "example-topic"
	for i := 0; i < 10; i++ {
		// 创建消息
		msg := Message{
			ID:        fmt.Sprintf("msg-%d", i+1),
			Content:   fmt.Sprintf("Hello, Kafka! (%d)", i+1),
			Timestamp: time.Now(),
		}

		// 创建生产者 Span
		ctx, span := tracer.Start(context.Background(), "kafka.produce",
			tracing.WithSpanKind(tracing.SpanKindProducer),
			tracing.WithSpanAttributes(
				attribute.String("messaging.system", "kafka"),
				attribute.String("messaging.destination", topic),
				attribute.String("messaging.destination_kind", "topic"),
				attribute.String("messaging.message_id", msg.ID),
			),
		)

		// 注入追踪上下文到消息头
		carrier := propagation.MapCarrier{}
		if err := tracer.Inject(ctx, carrier); err != nil {
			span.SetError(err)
			span.End()
			log.Printf("Failed to inject context: %v", err)
			continue
		}

		// 序列化消息
		value, err := json.Marshal(msg)
		if err != nil {
			span.SetError(err)
			span.End()
			log.Printf("Failed to marshal message: %v", err)
			continue
		}

		// 创建消息头
		headers := []kafka.Header{}
		for k, v := range carrier {
			headers = append(headers, kafka.Header{
				Key:   k,
				Value: []byte(v),
			})
		}

		// 发送消息
		err = producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Value:   value,
			Headers: headers,
			Key:     []byte(msg.ID),
		}, nil)
		if err != nil {
			span.SetError(err)
			span.End()
			log.Printf("Failed to produce message: %v", err)
			continue
		}

		span.End()
		time.Sleep(time.Second)
	}

	// 等待所有消息发送完成
	producer.Flush(15 * 1000)
}
