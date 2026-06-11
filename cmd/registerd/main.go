package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/juncaifeng/platform-kit/internal/config"
	"github.com/juncaifeng/platform-kit/internal/registry"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	slog.Info("starting registerd",
		"mode", cfg.Mode,
		"registry", cfg.RegistryType,
		"service", cfg.ServiceName,
	)

	var reg registry.Registry
	switch cfg.RegistryType {
	case config.RegistryConsul:
		reg, err = registry.NewConsulRegistry(cfg)
	default:
		slog.Error("unsupported registry type", "type", cfg.RegistryType)
		os.Exit(1)
	}

	if err != nil {
		slog.Error("failed to create registry", "error", err)
		os.Exit(1)
	}

	if err := reg.Register(); err != nil {
		slog.Error("failed to register service", "error", err)
		os.Exit(1)
	}

	if cfg.Mode == config.ModeOnce {
		slog.Info("once mode: service registered, exiting")
		os.Exit(0)
	}

	// 守护进程模式：持续运行
	slog.Info("daemon mode: starting heartbeat")

	stopCh := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	// 计算心跳间隔
	interval, err := time.ParseDuration(cfg.HealthCheckInterval)
	if err != nil {
		interval = 10 * time.Second
	}

	go reg.StartHeartbeat(interval, stopCh)

	sig := <-sigCh
	slog.Info("received signal, shutting down", "signal", sig)
	close(stopCh)

	if err := reg.Deregister(); err != nil {
		slog.Error("failed to deregister service", "error", err)
	}

	slog.Info("service deregistered, exiting")
}
