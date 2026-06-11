package event

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// NATSPublisher NATS 事件发布器
type NATSPublisher struct {
	conn *nats.Conn
	js   nats.JetStreamContext
	cfg  *NATSConfig
}

// newNATSPublisher 创建 NATS 发布器
func newNATSPublisher(cfg *NATSConfig) (*NATSPublisher, error) {
	opts := []nats.Option{
		nats.Name("platform-kit-publisher"),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(10),
	}

	conn, err := nats.Connect(cfg.Address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	p := &NATSPublisher{
		conn: conn,
		cfg:  cfg,
	}

	if cfg.JetStream {
		js, err := conn.JetStream()
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to create jetstream context: %w", err)
		}
		p.js = js
	}

	return p, nil
}

// Publish 发布事件
func (p *NATSPublisher) Publish(ctx context.Context, subject string, data []byte) error {
	msg := &nats.Msg{
		Subject: subject,
		Data:    data,
	}

	if p.js != nil {
		_, err := p.js.PublishMsg(msg)
		return err
	}

	return p.conn.PublishMsg(msg)
}

// Close 关闭连接
func (p *NATSPublisher) Close() error {
	p.conn.Close()
	return nil
}

// NATSConsumer NATS 事件消费者
type NATSConsumer struct {
	conn     *nats.Conn
	js       nats.JetStreamContext
	cfg      *NATSConfig
	subs     []*nats.Subscription
}

// newNATSConsumer 创建 NATS 消费者
func newNATSConsumer(cfg *NATSConfig) (*NATSConsumer, error) {
	opts := []nats.Option{
		nats.Name("platform-kit-consumer"),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(10),
	}

	conn, err := nats.Connect(cfg.Address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	c := &NATSConsumer{
		conn: conn,
		cfg:  cfg,
		subs: make([]*nats.Subscription, 0),
	}

	if cfg.JetStream {
		js, err := conn.JetStream()
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to create jetstream context: %w", err)
		}
		c.js = js
	}

	return c, nil
}

// Subscribe 订阅事件
func (c *NATSConsumer) Subscribe(ctx context.Context, subject string, handler Handler) error {
	natsHandler := func(msg *nats.Msg) {
		platformMsg := &Message{
			Subject: msg.Subject,
			Data:    msg.Data,
			ackFn: func() error {
				return msg.Ack()
			},
			nackFn: func() error {
				return msg.Nak()
			},
		}

		if err := handler(ctx, platformMsg); err != nil {
			// 消费失败，Nak 触发重试
			msg.Nak()
		} else {
			// 消费成功，Ack
			msg.Ack()
		}
	}

	var sub *nats.Subscription
	var err error

	if c.js != nil {
		sub, err = c.js.Subscribe(subject, natsHandler)
	} else {
		sub, err = c.conn.Subscribe(subject, natsHandler)
	}

	if err != nil {
		return fmt.Errorf("failed to subscribe %s: %w", subject, err)
	}

	c.subs = append(c.subs, sub)
	return nil
}

// Close 关闭连接
func (c *NATSConsumer) Close() error {
	for _, sub := range c.subs {
		sub.Unsubscribe()
	}
	c.conn.Close()
	return nil
}
