// Package servicev1 提供服务注册相关接口
package servicev1

import (
	"context"
	"time"
)

// ServiceRegistration 服务注册信息
type ServiceRegistration struct {
	// 服务名称
	Name string `json:"name"`
	// 服务地址
	Address string `json:"address"`
	// 服务端口
	Port int `json:"port"`
	// 服务类型: http | grpc
	Type string `json:"type"`
	// 健康检查路径
	HealthCheckPath string `json:"health_check_path"`
	// 网关暴露
	GatewayExposed bool `json:"gateway_exposed"`
	// 网关路由前缀
	GatewayPrefix string `json:"gateway_prefix"`
	// 产生的事件
	EventsProduced []string `json:"events_produced"`
	// 消费的事件
	EventsConsumed []string `json:"events_consumed"`
	// 自定义元数据
	Metadata map[string]string `json:"metadata"`
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	// 服务 ID
	ID string `json:"id"`
	// 服务名称
	Name string `json:"name"`
	// 服务地址
	Address string `json:"address"`
	// 服务端口
	Port int `json:"port"`
	// 服务状态: healthy | unhealthy | critical
	Status string `json:"status"`
	// 注册时间
	RegisteredAt time.Time `json:"registered_at"`
	// 最后心跳时间
	LastHeartbeat time.Time `json:"last_heartbeat"`
	// 元数据
	Metadata map[string]string `json:"metadata"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Service ServiceRegistration `json:"service"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	ServiceID string `json:"service_id"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

// DeregisterRequest 注销请求
type DeregisterRequest struct {
	ServiceID string `json:"service_id"`
}

// DeregisterResponse 注销响应
type DeregisterResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// HeartbeatRequest 心跳请求
type HeartbeatRequest struct {
	ServiceID string `json:"service_id"`
}

// HeartbeatResponse 心跳响应
type HeartbeatResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// ListServicesRequest 列出服务请求
type ListServicesRequest struct {
	Filter map[string]string `json:"filter,omitempty"`
}

// ListServicesResponse 列出服务响应
type ListServicesResponse struct {
	Services []ServiceInfo `json:"services"`
}

// ServiceRegistry 服务注册接口
type ServiceRegistry interface {
	// Register 注册服务
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)

	// Deregister 注销服务
	Deregister(ctx context.Context, req *DeregisterRequest) (*DeregisterResponse, error)

	// Heartbeat 心跳
	Heartbeat(ctx context.Context, req *HeartbeatRequest) (*HeartbeatResponse, error)

	// ListServices 列出服务
	ListServices(ctx context.Context, req *ListServicesRequest) (*ListServicesResponse, error)

	// GetService 获取服务详情
	GetService(ctx context.Context, serviceID string) (*ServiceInfo, error)
}
