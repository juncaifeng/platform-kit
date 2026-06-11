package registry

import "time"

// Registry 服务注册接口
type Registry interface {
	// Register 注册服务
	Register() error

	// Deregister 注销服务
	Deregister() error

	// Heartbeat 心跳
	Heartbeat() error

	// StartHeartbeat 启动心跳循环
	StartHeartbeat(interval time.Duration, stopCh <-chan struct{})
}
