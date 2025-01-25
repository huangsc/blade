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
	"github.com/huangsc/blade/mq"
)

type Message struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	// Kafka配置
	config := &mq.KafkaConfig{
		Brokers:  "localhost:9092",
		Username: "your-username",
		Password: "your-password",
	}

	// 创建生产者
	producer, err := mq.NewProducer(config)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	// 创建消费者
	consumer, err := mq.NewConsumer(config, "example-group")
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	// 订阅主题
	topic := "example-topic"
	err = consumer.Subscribe([]string{topic})
	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %v", err)
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动消费者协程
	go func() {
		err := consumer.Consume(ctx, func(msg *kafka.Message) error {
			var message Message
			if err := json.Unmarshal(msg.Value, &message); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				return nil
			}
			log.Printf("Received message: %+v", message)
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Printf("Consumer error: %v", err)
		}
	}()

	// 发送示例消息
	message := Message{
		ID:        "1",
		Content:   "Hello, Kafka!",
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Fatalf("Failed to marshal message: %v", err)
	}

	err = producer.SendMessage(ctx, topic, message.ID, data)
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	log.Printf("Sent message: %+v", message)

	// 等待信号
	<-sigChan
	log.Println("Shutting down...")
}
