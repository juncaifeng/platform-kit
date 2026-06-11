package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

type Mode string

const (
	ModeDaemon Mode = "daemon" // 后台守护进程模式
	ModeOnce   Mode = "once"   // 注册一次后退出
)

type RegistryType string

const (
	RegistryConsul   RegistryType = "consul"
	RegistryCloudMap RegistryType = "cloudmap"
	RegistryK8s      RegistryType = "k8s-service"
)

type Config struct {
	// 运行模式
	Mode Mode

	// 注册中心配置
	RegistryType RegistryType
	RegistryAddr string

	// 服务信息
	ServiceName string
	ServiceAddr string
	ServicePort int
	ServiceType string

	// 元数据
	EventsProduced string
	EventsConsumed string
	GatewayExposed string
	GatewayPrefix  string

	// 健康检查
	HealthCheckPath     string
	HealthCheckInterval string

	// TTL
	TTL int

	// Unix Socket 路径（用于进程间通信）
	SocketPath string
}

func Load() (*Config, error) {
	cfg := &Config{
		Mode:                Mode(getEnv("REGISTER_MODE", "daemon")),
		RegistryType:        RegistryType(getEnv("REGISTRY_TYPE", "consul")),
		RegistryAddr:        getEnv("REGISTRY_ADDR", "http://localhost:8500"),
		ServiceName:         getEnv("SERVICE_NAME", ""),
		ServiceAddr:         getEnv("SERVICE_ADDR", ""),
		ServicePort:         getEnvInt("SERVICE_PORT", 8080),
		ServiceType:         getEnv("SERVICE_TYPE", "http"),
		EventsProduced:      getEnv("EVENTS_PRODUCED", ""),
		EventsConsumed:      getEnv("EVENTS_CONSUMED", ""),
		GatewayExposed:      getEnv("GATEWAY_EXPOSED", "false"),
		GatewayPrefix:       getEnv("GATEWAY_PREFIX", ""),
		HealthCheckPath:     getEnv("HEALTH_CHECK_PATH", "/health"),
		HealthCheckInterval: getEnv("HEALTH_CHECK_INTERVAL", "10s"),
		TTL:                 getEnvInt("TTL", 30),
		SocketPath:          getEnv("REGISTER_SOCKET", "/var/run/platform-register.sock"),
	}

	// 自动检测 ServiceAddr
	if cfg.ServiceAddr != "" {
		ips, err := net.LookupHost(cfg.ServiceAddr)
		if err == nil && len(ips) > 0 {
			cfg.ServiceAddr = ips[0]
		}
	} else {
		cfg.ServiceAddr = detectServiceAddr()
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.ServiceName == "" {
		return fmt.Errorf("SERVICE_NAME is required")
	}
	if c.ServicePort <= 0 {
		return fmt.Errorf("SERVICE_PORT must be positive")
	}
	if c.RegistryAddr == "" {
		return fmt.Errorf("REGISTRY_ADDR is required")
	}
	return nil
}

func detectServiceAddr() string {
	if addr := os.Getenv("POD_IP"); addr != "" {
		return addr
	}

	if serviceAddr := os.Getenv("SERVICE_ADDR"); serviceAddr != "" {
		ips, err := net.LookupHost(serviceAddr)
		if err == nil && len(ips) > 0 {
			return ips[0]
		}
		return serviceAddr
	}

	hostname, _ := os.Hostname()
	if hostname != "" {
		addrs, err := net.LookupHost(hostname)
		if err == nil && len(addrs) > 0 {
			return addrs[0]
		}
	}

	return "127.0.0.1"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
