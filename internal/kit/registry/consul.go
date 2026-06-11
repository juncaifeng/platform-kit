package registry

import (
	"context"
	"fmt"

	"github.com/hashicorp/consul/api"
)

// ConsulRegistry Consul 服务注册
type ConsulRegistry struct {
	client *api.Client
}

// newConsulRegistry 创建 Consul 注册中心
func newConsulRegistry(address string) (*ConsulRegistry, error) {
	cfg := api.DefaultConfig()
	cfg.Address = address

	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return &ConsulRegistry{client: client}, nil
}

// Register 注册服务
func (r *ConsulRegistry) Register(ctx context.Context, service *Service) error {
	reg := &api.AgentServiceRegistration{
		ID:      service.ID,
		Name:    service.Name,
		Address: service.Address,
		Port:    service.Port,
		Meta:    service.Meta,
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/health", service.Address, service.Port),
			Interval: "10s",
			Timeout:  "5s",
		},
	}

	return r.client.Agent().ServiceRegister(reg)
}

// Deregister 注销服务
func (r *ConsulRegistry) Deregister(ctx context.Context, serviceID string) error {
	return r.client.Agent().ServiceDeregister(serviceID)
}

// Heartbeat 心跳
func (r *ConsulRegistry) Heartbeat(ctx context.Context, serviceID string) error {
	checkID := "service:" + serviceID
	return r.client.Agent().UpdateTTL(checkID, "healthy", api.HealthPassing)
}

// Close 关闭连接
func (r *ConsulRegistry) Close() error {
	return nil
}
