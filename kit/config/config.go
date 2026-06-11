// Package config 提供配置加载能力
//
// 支持 YAML 文件 + 环境变量覆盖
// 配置优先级: 环境变量 > 环境配置文件 > 服务模板 > 默认值
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config 应用配置
type Config struct {
	// 服务信息
	Service ServiceConfig `yaml:"service"`

	// 注册中心配置
	Registry RegistryConfig `yaml:"registry"`

	// 事件总线配置
	EventBus EventBusConfig `yaml:"eventBus"`

	// 网关配置
	Gateway GatewayConfig `yaml:"gateway"`

	// 传输层配置
	Transport TransportConfig `yaml:"transport"`
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name     string `yaml:"name"`
	Port     int    `yaml:"port"`
	Protocol string `yaml:"protocol"` // http / grpc
}

// RegistryConfig 注册中心配置
type RegistryConfig struct {
	Type    string `yaml:"type"`    // consul / cloudmap / k8s-service
	Address string `yaml:"address"`
}

// EventBusConfig 事件总线配置
type EventBusConfig struct {
	Type  string      `yaml:"type"` // nats / kafka / sqs
	NATS  *NATSConfig `yaml:"nats"`
	Kafka *KafkaConfig `yaml:"kafka"`
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

// GatewayConfig 网关配置
type GatewayConfig struct {
	Exposed bool   `yaml:"exposed"`
	Prefix  string `yaml:"prefix"`
}

// TransportConfig 传输层配置
type TransportConfig struct {
	HTTP HTTPConfig `yaml:"http"`
	GRPC GRPCConfig `yaml:"grpc"`
}

// HTTPConfig HTTP 配置
type HTTPConfig struct {
	Address         string `yaml:"address"`
	ShutdownTimeout int    `yaml:"shutdownTimeout"` // 秒
}

// GRPCConfig gRPC 配置
type GRPCConfig struct {
	Address         string `yaml:"address"`
	ShutdownTimeout int    `yaml:"shutdownTimeout"` // 秒
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Service: ServiceConfig{
			Port:     8080,
			Protocol: "http",
		},
		Registry: RegistryConfig{
			Type:    "consul",
			Address: "http://localhost:8500",
		},
		EventBus: EventBusConfig{
			Type: "nats",
			NATS: &NATSConfig{
				Address:   "nats://localhost:4222",
				JetStream: false,
			},
		},
		Transport: TransportConfig{
			HTTP: HTTPConfig{
				Address:         ":8080",
				ShutdownTimeout: 30,
			},
			GRPC: GRPCConfig{
				Address:         ":9090",
				ShutdownTimeout: 30,
			},
		},
	}
}

// Load 加载配置
// 1. 加载 YAML 文件
// 2. 环境变量覆盖
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	// 加载 YAML 文件
	if path != "" {
		if err := loadYAML(path, cfg); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// 环境变量覆盖
	loadEnv(cfg)

	return cfg, nil
}

// loadYAML 加载 YAML 文件
func loadYAML(path string, cfg *Config) error {
	// TODO: 实现 YAML 解析
	// 可以使用 gopkg.in/yaml.v3
	return nil
}

// loadEnv 从环境变量加载配置
func loadEnv(cfg *Config) {
	// 服务配置
	if v := os.Getenv("SERVICE_NAME"); v != "" {
		cfg.Service.Name = v
	}
	if v := os.Getenv("SERVICE_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Service.Port = port
		}
	}
	if v := os.Getenv("SERVICE_PROTOCOL"); v != "" {
		cfg.Service.Protocol = v
	}

	// 注册中心配置
	if v := os.Getenv("REGISTRY_TYPE"); v != "" {
		cfg.Registry.Type = v
	}
	if v := os.Getenv("REGISTRY_ADDR"); v != "" {
		cfg.Registry.Address = v
	}

	// 事件总线配置
	if v := os.Getenv("EVENT_BUS_TYPE"); v != "" {
		cfg.EventBus.Type = v
	}
	if v := os.Getenv("NATS_ADDRESS"); v != "" {
		if cfg.EventBus.NATS == nil {
			cfg.EventBus.NATS = &NATSConfig{}
		}
		cfg.EventBus.NATS.Address = v
	}
	if v := os.Getenv("NATS_JETSTREAM"); v != "" {
		if cfg.EventBus.NATS == nil {
			cfg.EventBus.NATS = &NATSConfig{}
		}
		cfg.EventBus.NATS.JetStream = strings.ToLower(v) == "true"
	}

	// 网关配置
	if v := os.Getenv("GATEWAY_EXPOSED"); v != "" {
		cfg.Gateway.Exposed = strings.ToLower(v) == "true"
	}
	if v := os.Getenv("GATEWAY_PREFIX"); v != "" {
		cfg.Gateway.Prefix = v
	}
}
