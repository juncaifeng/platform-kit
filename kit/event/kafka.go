package event

import (
	"context"
	"fmt"
)

// KafkaPublisher Kafka 事件发布器（占位实现）
type KafkaPublisher struct {
	cfg *KafkaConfig
}

// newKafkaPublisher 创建 Kafka 发布器
func newKafkaPublisher(cfg *KafkaConfig) (*KafkaPublisher, error) {
	// TODO: 实现 Kafka 发布器
	return nil, fmt.Errorf("kafka publisher not implemented yet")
}

// Publish 发布事件
func (p *KafkaPublisher) Publish(ctx context.Context, subject string, data []byte) error {
	return fmt.Errorf("kafka publisher not implemented yet")
}

// Close 关闭连接
func (p *KafkaPublisher) Close() error {
	return nil
}

// KafkaConsumer Kafka 事件消费者（占位实现）
type KafkaConsumer struct {
	cfg *KafkaConfig
}

// newKafkaConsumer 创建 Kafka 消费者
func newKafkaConsumer(cfg *KafkaConfig) (*KafkaConsumer, error) {
	// TODO: 实现 Kafka 消费者
	return nil, fmt.Errorf("kafka consumer not implemented yet")
}

// Subscribe 订阅事件
func (c *KafkaConsumer) Subscribe(ctx context.Context, subject string, handler Handler) error {
	return fmt.Errorf("kafka consumer not implemented yet")
}

// Close 关闭连接
func (c *KafkaConsumer) Close() error {
	return nil
}
