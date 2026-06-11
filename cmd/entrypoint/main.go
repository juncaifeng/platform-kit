package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// 获取业务服务的命令
	args := os.Args[1:]
	if len(args) == 0 {
		slog.Error("no command specified")
		fmt.Fprintf(os.Stderr, "Usage: entrypoint <command> [args...]\n")
		os.Exit(1)
	}

	// 启动 registerd 后台进程
	registerdPath := "/app/registerd"
	if path := os.Getenv("REGISTERD_PATH"); path != "" {
		registerdPath = path
	}

	slog.Info("starting registerd", "path", registerdPath)
	registerd := exec.Command(registerdPath)
	registerd.Stdout = os.Stdout
	registerd.Stderr = os.Stderr
	registerd.Env = os.Environ()

	if err := registerd.Start(); err != nil {
		slog.Error("failed to start registerd", "error", err)
		os.Exit(1)
	}

	slog.Info("registerd started", "pid", registerd.Process.Pid)

	// 启动业务服务
	slog.Info("starting application", "command", args)
	application := exec.Command(args[0], args[1:]...)
	application.Stdin = os.Stdin
	application.Stdout = os.Stdout
	application.Stderr = os.Stderr
	application.Env = os.Environ()

	if err := application.Start(); err != nil {
		slog.Error("failed to start application", "error", err)
		// 停止 registerd
		registerd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	slog.Info("application started", "pid", application.Process.Pid)

	// 信号处理
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	// 等待任一进程退出或收到信号
	done := make(chan error, 2)

	go func() {
		done <- application.Wait()
	}()

	go func() {
		done <- registerd.Wait()
	}()

	select {
	case sig := <-sigCh:
		slog.Info("received signal", "signal", sig)
	case err := <-done:
		if err != nil {
			slog.Info("process exited with error", "error", err)
		} else {
			slog.Info("process exited")
		}
	}

	// 优雅关闭
	slog.Info("shutting down...")

	// 停止业务服务
	if application.Process != nil {
		application.Process.Signal(syscall.SIGTERM)
	}

	// 停止 registerd
	if registerd.Process != nil {
		registerd.Process.Signal(syscall.SIGTERM)
	}

	// 等待退出
	application.Wait()
	registerd.Wait()

	slog.Info("all processes stopped")
}
