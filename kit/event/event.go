// Package event 定义事件发布/消费接口
//
// 业务代码通过此接口发布和消费事件，不感知底层实现（NATS/Kafka/SQS）
// 切换事件总线只需修改配置，业务代码零改动
package event

import (
	"context"
	"fmt"
)

// Publisher 事件发布接口
type Publisher interface {
	// Publish 发布事件
	// subject 格式: {domain}.{event}.{version}，如 order.created.v1
	Publish(ctx context.Context, subject string, data []byte) error

	// Close 关闭连接
	Close() error
}

// Consumer 事件消费接口
type Consumer interface {
	// Subscribe 订阅事件
	// subject 支持通配符: order.*.*, order.>
	Subscribe(ctx context.Context, subject string, handler Handler) error

	// Close 关闭连接
	Close() error
}

// Handler 事件处理函数
type Handler func(ctx context.Context, msg *Message) error

// Message 事件消息
type Message struct {
	Subject string
	Data    []byte
	Headers map[string]string

	// ack 用于手动确认（异步服务需要）
	ackFn   func() error
	nackFn  func() error
}

// Ack 手动确认消息
func (m *Message) Ack() error {
	if m.ackFn != nil {
		return m.ackFn()
	}
	return nil
}

// Nack 拒绝消息（触发重试）
func (m *Message) Nack() error {
	if m.nackFn != nil {
		return m.nackFn()
	}
	return nil
}

// EventBus 事件总线配置
type EventBusConfig struct {
	Type     string      `yaml:"type"`     // nats / kafka / sqs
	NATS     *NATSConfig `yaml:"nats"`
	Kafka    *KafkaConfig `yaml:"kafka"`
}

// NATSConfig NATS 配置
type NATSConfig struct {
	Address   string `yaml:"address"`
	JetStream bool   `yaml:"jetStream"`
}

// KafkaConfig Kafka 配置
type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
}

// NewPublisher 根据配置创建事件发布器
func NewPublisher(cfg EventBusConfig) (Publisher, error) {
	switch cfg.Type {
	case "nats":
		if cfg.NATS == nil {
			return nil, fmt.Errorf("nats config is required")
		}
		return newNATSPublisher(cfg.NATS)
	case "kafka":
		if cfg.Kafka == nil {
			return nil, fmt.Errorf("kafka config is required")
		}
		return newKafkaPublisher(cfg.Kafka)
	default:
		return nil, fmt.Errorf("unsupported event bus type: %s", cfg.Type)
	}
}

// NewConsumer 根据配置创建事件消费者
func NewConsumer(cfg EventBusConfig) (Consumer, error) {
	switch cfg.Type {
	case "nats":
		if cfg.NATS == nil {
			return nil, fmt.Errorf("nats config is required")
		}
		return newNATSConsumer(cfg.NATS)
	case "kafka":
		if cfg.Kafka == nil {
			return nil, fmt.Errorf("kafka config is required")
		}
		return newKafkaConsumer(cfg.Kafka)
	default:
		return nil, fmt.Errorf("unsupported event bus type: %s", cfg.Type)
	}
}
