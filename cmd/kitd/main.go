// Kit Sidecar 主程序
//
// 作为独立进程运行，提供以下能力：
// 1. 服务注册（Consul/CloudMap/K8s）
// 2. 健康检查（/health, /ready）
// 3. 事件发布/消费（NATS/Kafka）
// 4. 指标采集（/metrics）
// 5. gRPC API（供业务进程调用）
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/juncaifeng/platform-kit/internal/kit/event"
	"github.com/juncaifeng/platform-kit/internal/kit/registry"
)

func main() {
	// 初始化日志
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// 加载配置
	cfg := loadConfig()

	slog.Info("starting kit sidecar",
		"service", cfg.ServiceName,
		"registry", cfg.RegistryType,
		"event_bus", cfg.EventBusType,
	)

	// 创建注册中心
	reg, err := registry.New(registry.Config{
		Type:    cfg.RegistryType,
		Address: cfg.RegistryAddr,
	})
	if err != nil {
		slog.Error("failed to create registry", "error", err)
		os.Exit(1)
	}
	defer reg.Close()

	// 创建事件总线
	eventBus, err := event.New(event.Config{
		Type:    cfg.EventBusType,
		Address: cfg.EventBusAddr,
	})
	if err != nil {
		slog.Error("failed to create event bus", "error", err)
		os.Exit(1)
	}
	defer eventBus.Close()

	// 启动 HTTP 服务器（健康检查 + 指标）
	httpServer := startHTTPServer(cfg, reg)

	// 启动 gRPC 服务器
	grpcServer := startGRPCServer(cfg, reg, eventBus)

	// 注册服务
	service := &registry.Service{
		ID:      fmt.Sprintf("%s-%s", cfg.ServiceName, cfg.PodIP),
		Name:    cfg.ServiceName,
		Address: cfg.PodIP,
		Port:    cfg.ServicePort,
		Meta: map[string]string{
			"type":             cfg.ServiceType,
			"gateway_exposed":  cfg.GatewayExposed,
			"gateway_prefix":   cfg.GatewayPrefix,
			"events_produced":  cfg.EventsProduced,
			"events_consumed":  cfg.EventsConsumed,
		},
	}

	if err := reg.Register(context.Background(), service); err != nil {
		slog.Error("failed to register service", "error", err)
	} else {
		slog.Info("service registered", "id", service.ID)
	}

	// 信号处理
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	sig := <-sigCh
	slog.Info("received signal, shutting down", "signal", sig)

	// 注销服务
	if err := reg.Deregister(context.Background(), service.ID); err != nil {
		slog.Error("failed to deregister service", "error", err)
	}

	// 停止服务器
	grpcServer.GracefulStop()
	httpServer.Shutdown(context.Background())

	slog.Info("kit sidecar stopped")
}

// Config 配置
type Config struct {
	// 服务信息
	ServiceName string
	ServicePort int
	ServiceType string
	PodIP       string

	// 注册中心
	RegistryType string
	RegistryAddr string

	// 事件总线
	EventBusType string
	EventBusAddr string

	// 网关
	GatewayExposed string
	GatewayPrefix  string

	// 事件
	EventsProduced string
	EventsConsumed string

	// Kit Sidecar 端口
	KitGRPCPort int
	KitHTTPPort int
}

func loadConfig() *Config {
	return &Config{
		ServiceName:    getEnv("SERVICE_NAME", "unknown"),
		ServicePort:    getEnvInt("SERVICE_PORT", 8080),
		ServiceType:    getEnv("SERVICE_TYPE", "http"),
		PodIP:          getEnv("POD_IP", "127.0.0.1"),
		RegistryType:   getEnv("REGISTRY_TYPE", "consul"),
		RegistryAddr:   getEnv("REGISTRY_ADDR", "http://localhost:8500"),
		EventBusType:   getEnv("EVENT_BUS_TYPE", "nats"),
		EventBusAddr:   getEnv("EVENT_BUS_ADDR", "nats://localhost:4222"),
		GatewayExposed: getEnv("GATEWAY_EXPOSED", "false"),
		GatewayPrefix:  getEnv("GATEWAY_PREFIX", ""),
		EventsProduced: getEnv("EVENTS_PRODUCED", ""),
		EventsConsumed: getEnv("EVENTS_CONSUMED", ""),
		KitGRPCPort:    getEnvInt("KIT_GRPC_PORT", 9090),
		KitHTTPPort:    getEnvInt("KIT_HTTP_PORT", 9091),
	}
}

func startHTTPServer(cfg *Config, reg registry.Registry) *http.Server {
	mux := http.NewServeMux()

	// 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// 就绪检查
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ready":true}`))
	})

	// 指标
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("# Prometheus metrics placeholder\n"))
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.KitHTTPPort),
		Handler: mux,
	}

	go func() {
		slog.Info("kit HTTP server starting", "port", cfg.KitHTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	return server
}

func startGRPCServer(cfg *Config, reg registry.Registry, eventBus event.EventBus) *grpc.Server {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.KitGRPCPort))
	if err != nil {
		slog.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()

	// 注册健康检查服务
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	// TODO: 注册 Kit 服务
	// kitv1.RegisterKitServiceServer(grpcServer, &kitService{
	//     registry: reg,
	//     eventBus: eventBus,
	// })

	go func() {
		slog.Info("kit gRPC server starting", "port", cfg.KitGRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("gRPC server error", "error", err)
		}
	}()

	return grpcServer
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intVal int
		if _, err := fmt.Sscanf(value, "%d", &intVal); err == nil {
			return intVal
		}
	}
	return defaultValue
}
