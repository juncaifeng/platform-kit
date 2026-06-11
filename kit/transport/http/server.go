// Package http 提供 HTTP 传输层实现
//
// 自动注册标准接口: /health, /ready, /metrics
package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Server HTTP 服务器
type Server struct {
	mux      *http.ServeMux
	server   *http.Server
	checker  HealthChecker
	metrics  MetricsCollector
}

// HealthChecker 健康检查器
type HealthChecker interface {
	Live(ctx context.Context) error
	Ready(ctx context.Context) error
}

// MetricsCollector 指标收集器
type MetricsCollector interface {
	Handler() http.Handler
}

// Option 服务器选项
type Option func(*Server)

// WithHealthChecker 设置健康检查器
func WithHealthChecker(checker HealthChecker) Option {
	return func(s *Server) {
		s.checker = checker
	}
}

// WithMetrics 设置指标收集器
func WithMetrics(collector MetricsCollector) Option {
	return func(s *Server) {
		s.metrics = collector
	}
}

// NewServer 创建 HTTP 服务器
func NewServer(address string, opts ...Option) *Server {
	s := &Server{
		mux: http.NewServeMux(),
	}

	for _, opt := range opts {
		opt(s)
	}

	// 注册标准接口
	s.registerStandardEndpoints()

	s.server = &http.Server{
		Addr:    address,
		Handler: s.mux,
	}

	return s
}

// registerStandardEndpoints 注册标准接口
func (s *Server) registerStandardEndpoints() {
	// /health 存活检查
	s.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if s.checker != nil {
			if err := s.checker.Live(r.Context()); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status": "unhealthy",
					"error":  err.Error(),
				})
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "healthy",
		})
	})

	// /ready 就绪检查
	s.mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		if s.checker != nil {
			if err := s.checker.Ready(r.Context()); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"ready": false,
					"error": err.Error(),
				})
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ready": true,
		})
	})

	// /metrics 监控指标
	if s.metrics != nil {
		s.mux.Handle("/metrics", s.metrics.Handler())
	}

	// /config 配置查看（可选）
	s.mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service": "platform-kit",
		})
	})
}

// Handle 注册路由
func (s *Server) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

// HandleFunc 注册路由函数
func (s *Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.mux.HandleFunc(pattern, handler)
}

// Start 启动服务器
func (s *Server) Start(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		s.Stop(context.Background())
	}()

	fmt.Printf("HTTP server starting on %s\n", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server error: %w", err)
	}
	return nil
}

// Stop 优雅停止服务器
func (s *Server) Stop(ctx context.Context) error {
	fmt.Println("HTTP server shutting down...")
	return s.server.Shutdown(ctx)
}
