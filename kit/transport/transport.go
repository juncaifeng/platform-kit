// Package transport 定义传输层接口
package transport

import "context"

// Server 传输层服务器接口
type Server interface {
	// Start 启动服务器
	Start(ctx context.Context) error

	// Stop 优雅停止服务器
	Stop(ctx context.Context) error
}

// HealthChecker 健康检查接口
type HealthChecker interface {
	// Live 存活检查
	Live(ctx context.Context) error

	// Ready 就绪检查
	Ready(ctx context.Context) error
}
