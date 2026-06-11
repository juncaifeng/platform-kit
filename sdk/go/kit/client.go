// Package kit 提供 Kit Sidecar 客户端 SDK
//
// 业务服务通过此 SDK 调用 Kit Sidecar 的能力
// 支持多语言：Go、Python、Java、Node.js
package kit

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client Kit Sidecar 客户端
type Client struct {
	conn   *grpc.Conn
	config *Config
}

// Config 客户端配置
type Config struct {
	// Kit Sidecar 地址
	KitAddr string `yaml:"kitAddr" env:"KIT_SIDECAR_ADDR" default:"localhost:9090"`
}

// New 创建 Kit 客户端
func New(cfg *Config) (*Client, error) {
	conn, err := grpc.Dial(cfg.KitAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to kit sidecar: %w", err)
	}

	return &Client{
		conn:   conn,
		config: cfg,
	}, nil
}

// Close 关闭连接
func (c *Client) Close() error {
	return c.conn.Close()
}

// Register 注册服务
func (c *Client) Register(ctx context.Context, service *ServiceInfo) error {
	// TODO: 调用 gRPC API
	return nil
}

// Deregister 注销服务
func (c *Client) Deregister(ctx context.Context, serviceID string) error {
	// TODO: 调用 gRPC API
	return nil
}

// PublishEvent 发布事件
func (c *Client) PublishEvent(ctx context.Context, subject string, data []byte) error {
	// TODO: 调用 gRPC API
	return nil
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Name    string            `json:"name"`
	ID      string            `json:"id"`
	Address string            `json:"address"`
	Port    int               `json:"port"`
	Meta    map[string]string `json:"meta"`
}
