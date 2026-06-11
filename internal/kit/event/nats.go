package event

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
)

// NATSEventBus NATS 事件总线
type NATSEventBus struct {
	conn *nats.Conn
}

// newNATSEventBus 创建 NATS 事件总线
func newNATSEventBus(address string) (*NATSEventBus, error) {
	conn, err := nats.Connect(address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	return &NATSEventBus{conn: conn}, nil
}

// Publish 发布事件
func (b *NATSEventBus) Publish(ctx context.Context, subject string, data []byte, headers map[string]string) error {
	msg := &nats.Msg{
		Subject: subject,
		Data:    data,
	}

	// 添加 headers
	if len(headers) > 0 {
		msg.Header = make(nats.Header)
		for k, v := range headers {
			msg.Header.Set(k, v)
		}
	}

	return b.conn.PublishMsg(msg)
}

// Subscribe 订阅事件
func (b *NATSEventBus) Subscribe(ctx context.Context, subject string, handler EventHandler) error {
	_, err := b.conn.Subscribe(subject, func(msg *nats.Msg) {
		platformMsg := &Message{
			Subject:   msg.Subject,
			Data:      msg.Data,
			MessageID: msg.Header.Get("Nats-Msg-Id"),
		}

		// 复制 headers
		if msg.Header != nil {
			platformMsg.Headers = make(map[string]string)
			for k, v := range msg.Header {
				if len(v) > 0 {
					platformMsg.Headers[k] = v[0]
				}
			}
		}

		if err := handler(platformMsg); err != nil {
			// 记录错误日志
			fmt.Printf("event handler error: %v\n", err)
		}
	})

	return err
}

// Close 关闭连接
func (b *NATSEventBus) Close() error {
	b.conn.Close()
	return nil
}
