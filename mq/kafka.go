package mq

import (
	"context"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// KafkaConfig Kafka配置
type KafkaConfig struct {
	Brokers  string
	Username string
	Password string
}

// Producer Kafka生产者
type Producer struct {
	producer *kafka.Producer
	config   *KafkaConfig
}

// Consumer Kafka消费者
type Consumer struct {
	consumer *kafka.Consumer
	config   *KafkaConfig
}

// NewProducer 创建生产者
func NewProducer(config *KafkaConfig) (*Producer, error) {
	conf := &kafka.ConfigMap{
		"bootstrap.servers": config.Brokers,
		"security.protocol": "SASL_PLAINTEXT",
		"sasl.mechanisms":   "PLAIN",
		"sasl.username":     config.Username,
		"sasl.password":     config.Password,
	}

	producer, err := kafka.NewProducer(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %v", err)
	}

	return &Producer{
		producer: producer,
		config:   config,
	}, nil
}

// Close 关闭生产者
func (p *Producer) Close() {
	p.producer.Close()
}

// SendMessage 发送消息
func (p *Producer) SendMessage(ctx context.Context, topic string, key string, value []byte) error {
	deliveryChan := make(chan kafka.Event)
	defer close(deliveryChan)

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:       []byte(key),
		Value:     value,
		Timestamp: time.Now(),
	}

	err := p.producer.Produce(msg, deliveryChan)
	if err != nil {
		return fmt.Errorf("failed to produce message: %v", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-deliveryChan:
		m := e.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			return fmt.Errorf("delivery failed: %v", m.TopicPartition.Error)
		}
	}

	return nil
}

// NewConsumer 创建消费者
func NewConsumer(config *KafkaConfig, groupID string) (*Consumer, error) {
	conf := &kafka.ConfigMap{
		"bootstrap.servers": config.Brokers,
		"security.protocol": "SASL_PLAINTEXT",
		"sasl.mechanisms":   "PLAIN",
		"sasl.username":     config.Username,
		"sasl.password":     config.Password,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	}

	consumer, err := kafka.NewConsumer(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %v", err)
	}

	return &Consumer{
		consumer: consumer,
		config:   config,
	}, nil
}

// Close 关闭消费者
func (c *Consumer) Close() error {
	return c.consumer.Close()
}

// Subscribe 订阅主题
func (c *Consumer) Subscribe(topics []string) error {
	return c.consumer.SubscribeTopics(topics, nil)
}

// Consume 消费消息
func (c *Consumer) Consume(ctx context.Context, handler func(*kafka.Message) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				return fmt.Errorf("error reading message: %v", err)
			}

			if err := handler(msg); err != nil {
				return fmt.Errorf("error handling message: %v", err)
			}
		}
	}
}
