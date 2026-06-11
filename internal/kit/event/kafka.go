package event

import (
	"context"
	"fmt"
)

// KafkaEventBus Kafka 事件总线（占位实现）
type KafkaEventBus struct{}

// newKafkaEventBus 创建 Kafka 事件总线
func newKafkaEventBus(address string) (*KafkaEventBus, error) {
	// TODO: 实现 Kafka 事件总线
	return nil, fmt.Errorf("kafka event bus not implemented yet")
}

// Publish 发布事件
func (b *KafkaEventBus) Publish(ctx context.Context, subject string, data []byte, headers map[string]string) error {
	return fmt.Errorf("kafka event bus not implemented yet")
}

// Subscribe 订阅事件
func (b *KafkaEventBus) Subscribe(ctx context.Context, subject string, handler EventHandler) error {
	return fmt.Errorf("kafka event bus not implemented yet")
}

// Close 关闭连接
func (b *KafkaEventBus) Close() error {
	return nil
}
