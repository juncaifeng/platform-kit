// Package registry 定义服务注册接口
//
// 业务代码通过此接口注册服务，不感知底层实现（Consul/CloudMap/K8s）
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
	Check   *HealthCheck      `json:"check,omitempty"`
}

// HealthCheck 健康检查配置
type HealthCheck struct {
	HTTP     string `json:"http,omitempty"`
	Interval string `json:"interval,omitempty"`
	Timeout  string `json:"timeout,omitempty"`
}

// RegistryConfig 注册中心配置
type RegistryConfig struct {
	Type    string        `yaml:"type"` // consul / cloudmap / k8s-service
	Consul  *ConsulConfig `yaml:"consul"`
	CloudMap *CloudMapConfig `yaml:"cloudmap"`
}

// ConsulConfig Consul 配置
type ConsulConfig struct {
	Address string `yaml:"address"`
	Token   string `yaml:"token"`
}

// CloudMapConfig AWS Cloud Map 配置
type CloudMapConfig struct {
	Namespace string `yaml:"namespace"`
	Region    string `yaml:"region"`
}

// NewRegistry 根据配置创建注册中心
func NewRegistry(cfg RegistryConfig) (Registry, error) {
	switch cfg.Type {
	case "consul":
		if cfg.Consul == nil {
			return nil, fmt.Errorf("consul config is required")
		}
		return newConsulRegistry(cfg.Consul)
	case "cloudmap":
		if cfg.CloudMap == nil {
			return nil, fmt.Errorf("cloudmap config is required")
		}
		return newCloudMapRegistry(cfg.CloudMap)
	case "k8s-service":
		return newK8sRegistry(), nil
	default:
		return nil, fmt.Errorf("unsupported registry type: %s", cfg.Type)
	}
}
