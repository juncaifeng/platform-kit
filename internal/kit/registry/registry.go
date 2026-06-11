// Package registry 定义服务注册接口
package registry

import (
	"context"
	"fmt"
)

// Registry 服务注册接口
type Registry interface {
	// Register 注册服务
	Register(ctx context.Context, service *Service) error

	// Deregister 注销服务
	Deregister(ctx context.Context, serviceID string) error

	// Heartbeat 心跳
	Heartbeat(ctx context.Context, serviceID string) error

	// Close 关闭连接
	Close() error
}

// Service 服务信息
type Service struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Address string            `json:"address"`
	Port    int               `json:"port"`
	Meta    map[string]string `json:"meta"`
}

// Config 注册中心配置
type Config struct {
	Type    string // consul / cloudmap / k8s-service
	Address string
}

// New 根据配置创建注册中心
func New(cfg Config) (Registry, error) {
	switch cfg.Type {
	case "consul":
		return newConsulRegistry(cfg.Address)
	case "k8s-service":
		return newK8sRegistry(), nil
	default:
		return nil, fmt.Errorf("unsupported registry type: %s", cfg.Type)
	}
}
