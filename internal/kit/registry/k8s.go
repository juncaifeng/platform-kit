package registry

import "context"

// K8sRegistry K8s Service 服务注册
type K8sRegistry struct{}

// newK8sRegistry 创建 K8s 注册中心
func newK8sRegistry() *K8sRegistry {
	return &K8sRegistry{}
}

// Register 注册服务（K8s 环境下为空实现）
func (r *K8sRegistry) Register(ctx context.Context, service *Service) error {
	return nil
}

// Deregister 注销服务（K8s 环境下为空实现）
func (r *K8sRegistry) Deregister(ctx context.Context, serviceID string) error {
	return nil
}

// Heartbeat 心跳（K8s 环境下为空实现）
func (r *K8sRegistry) Heartbeat(ctx context.Context, serviceID string) error {
	return nil
}

// Close 关闭连接
func (r *K8sRegistry) Close() error {
	return nil
}
