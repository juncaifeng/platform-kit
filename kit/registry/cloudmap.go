package registry

import (
	"context"
	"fmt"
)

// CloudMapRegistry AWS Cloud Map 服务注册
type CloudMapRegistry struct {
	cfg *CloudMapConfig
}

// newCloudMapRegistry 创建 Cloud Map 注册中心
func newCloudMapRegistry(cfg *CloudMapConfig) (*CloudMapRegistry, error) {
	// TODO: 实现 AWS Cloud Map 注册
	return nil, fmt.Errorf("cloudmap registry not implemented yet")
}

// Register 注册服务
func (r *CloudMapRegistry) Register(ctx context.Context, service *Service) error {
	return fmt.Errorf("cloudmap registry not implemented yet")
}

// Deregister 注销服务
func (r *CloudMapRegistry) Deregister(ctx context.Context, serviceID string) error {
	return fmt.Errorf("cloudmap registry not implemented yet")
}

// Heartbeat 心跳
func (r *CloudMapRegistry) Heartbeat(ctx context.Context, serviceID string) error {
	return fmt.Errorf("cloudmap registry not implemented yet")
}

// Close 关闭连接
func (r *CloudMapRegistry) Close() error {
	return nil
}
