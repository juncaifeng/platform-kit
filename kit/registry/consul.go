package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
)

// ConsulRegistry Consul 服务注册
type ConsulRegistry struct {
	client *api.Client
	cfg    *ConsulConfig
}

// newConsulRegistry 创建 Consul 注册中心
func newConsulRegistry(cfg *ConsulConfig) (*ConsulRegistry, error) {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = cfg.Address
	if cfg.Token != "" {
		consulConfig.Token = cfg.Token
	}

	client, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return &ConsulRegistry{
		client: client,
		cfg:    cfg,
	}, nil
}

// Register 注册服务
func (r *ConsulRegistry) Register(ctx context.Context, service *Service) error {
	reg := &api.AgentServiceRegistration{
		ID:      service.ID,
		Name:    service.Name,
		Address: service.Address,
		Port:    service.Port,
		Meta:    service.Meta,
	}

	if service.Check != nil {
		reg.Check = &api.AgentServiceCheck{
			HTTP:     service.Check.HTTP,
			Interval: service.Check.Interval,
			Timeout:  service.Check.Timeout,
		}
	}

	if err := r.client.Agent().ServiceRegister(reg); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	return nil
}

// Deregister 注销服务
func (r *ConsulRegistry) Deregister(ctx context.Context, serviceID string) error {
	if err := r.client.Agent().ServiceDeregister(serviceID); err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}
	return nil
}

// Heartbeat 心跳
func (r *ConsulRegistry) Heartbeat(ctx context.Context, serviceID string) error {
	checkID := "service:" + serviceID
	if err := r.client.Agent().UpdateTTL(checkID, "service is healthy", api.HealthPassing); err != nil {
		return fmt.Errorf("failed to update TTL: %w", err)
	}
	return nil
}

// Close 关闭连接
func (r *ConsulRegistry) Close() error {
	// Consul 客户端没有显式的 Close 方法
	return nil
}

// StartHeartbeat 启动心跳循环
func (r *ConsulRegistry) StartHeartbeat(ctx context.Context, serviceID string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := r.Heartbeat(ctx, serviceID); err != nil {
				// 记录日志但不中断
				fmt.Printf("heartbeat failed: %v\n", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
