// Package app 提供应用启动框架
//
// 统一处理: 启动、健康检查、优雅停止、信号处理
// 业务代码只需注册路由和业务逻辑，其他由 Kit 处理
package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/juncaifeng/platform-kit/kit/config"
	"github.com/juncaifeng/platform-kit/kit/event"
	"github.com/juncaifeng/platform-kit/kit/transport"
	httpTransport "github.com/juncaifeng/platform-kit/kit/transport/http"
)

// App 应用实例
type App struct {
	cfg        *config.Config
	servers    []transport.Server
	eventPub   event.Publisher
	eventSub   event.Consumer
	startFn    func(ctx context.Context) error
	stopFn     func(ctx context.Context) error
	mu         sync.Mutex
}

// Option 应用选项
type Option func(*App)

// WithConfig 设置配置
func WithConfig(cfg *config.Config) Option {
	return func(a *App) {
		a.cfg = cfg
	}
}

// WithHTTPServer 添加 HTTP 服务器
func WithHTTPServer(address string, opts ...httpTransport.Option) Option {
	return func(a *App) {
		server := httpTransport.NewServer(address, opts...)
		a.servers = append(a.servers, server)
	}
}

// WithEventPublisher 设置事件发布器
func WithEventPublisher(pub event.Publisher) Option {
	return func(a *App) {
		a.eventPub = pub
	}
}

// WithEventConsumer 设置事件消费者
func WithEventConsumer(sub event.Consumer) Option {
	return func(a *App) {
		a.eventSub = sub
	}
}

// WithStartFunc 设置自定义启动函数
func WithStartFunc(fn func(ctx context.Context) error) Option {
	return func(a *App) {
		a.startFn = fn
	}
}

// WithStopFunc 设置自定义停止函数
func WithStopFunc(fn func(ctx context.Context) error) Option {
	return func(a *App) {
		a.stopFn = fn
	}
}

// New 创建应用实例
func New(opts ...Option) *App {
	a := &App{
		cfg:     config.DefaultConfig(),
		servers: make([]transport.Server, 0),
	}

	for _, opt := range opts {
		opt(a)
	}

	return a
}

// HTTPServer 获取 HTTP 服务器（用于注册业务路由）
func (a *App) HTTPServer() *httpTransport.Server {
	for _, s := range a.servers {
		if httpServer, ok := s.(*httpTransport.Server); ok {
			return httpServer
		}
	}
	return nil
}

// EventPublisher 获取事件发布器
func (a *App) EventPublisher() event.Publisher {
	return a.eventPub
}

// Run 运行应用
// 1. 启动所有服务器
// 2. 等待退出信号
// 3. 优雅停止
func (a *App) Run(ctx context.Context) error {
	// 创建可取消的 context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 信号处理
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	// 启动所有服务器
	var wg sync.WaitGroup
	errCh := make(chan error, len(a.servers))

	for _, server := range a.servers {
		wg.Add(1)
		go func(s transport.Server) {
			defer wg.Done()
			if err := s.Start(ctx); err != nil {
				errCh <- err
			}
		}(server)
	}

	// 执行自定义启动函数
	if a.startFn != nil {
		if err := a.startFn(ctx); err != nil {
			return fmt.Errorf("start function error: %w", err)
		}
	}

	slog.Info("application started",
		"service", a.cfg.Service.Name,
		"port", a.cfg.Service.Port,
	)

	// 等待退出信号或错误
	select {
	case sig := <-sigCh:
		slog.Info("received signal, shutting down", "signal", sig)
	case err := <-errCh:
		return err
	}

	// 优雅停止
	return a.shutdown(ctx)
}

// shutdown 优雅停止
func (a *App) shutdown(ctx context.Context) error {
	// 创建超时 context
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	slog.Info("shutting down application...")

	// 执行自定义停止函数
	if a.stopFn != nil {
		if err := a.stopFn(shutdownCtx); err != nil {
			slog.Error("stop function error", "error", err)
		}
	}

	// 停止所有服务器
	for _, server := range a.servers {
		if err := server.Stop(shutdownCtx); err != nil {
			slog.Error("server stop error", "error", err)
		}
	}

	// 关闭事件发布器
	if a.eventPub != nil {
		if err := a.eventPub.Close(); err != nil {
			slog.Error("event publisher close error", "error", err)
		}
	}

	// 关闭事件消费者
	if a.eventSub != nil {
		if err := a.eventSub.Close(); err != nil {
			slog.Error("event consumer close error", "error", err)
		}
	}

	slog.Info("application stopped")
	return nil
}
