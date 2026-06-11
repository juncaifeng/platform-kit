// Package event 定义事件总线接口
package event

import (
	"context"
	"fmt"
)

// EventBus 事件总线接口
type EventBus interface {
	// Publish 发布事件
	Publish(ctx context.Context, subject string, data []byte, headers map[string]string) error

	// Subscribe 订阅事件
	Subscribe(ctx context.Context, subject string, handler EventHandler) error

	// Close 关闭连接
	Close() error
}

// EventHandler 事件处理函数
type EventHandler func(msg *Message) error

// Message 事件消息
type Message struct {
	Subject   string
	Data      []byte
	Headers   map[string]string
	MessageID string
}

// Config 事件总线配置
type Config struct {
	Type    string // nats / kafka
	Address string
}

// New 根据配置创建事件总线
func New(cfg Config) (EventBus, error) {
	switch cfg.Type {
	case "nats":
		return newNATSEventBus(cfg.Address)
	case "kafka":
		return newKafkaEventBus(cfg.Address)
	default:
		return nil, fmt.Errorf("unsupported event bus type: %s", cfg.Type)
	}
}
