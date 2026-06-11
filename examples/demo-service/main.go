package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/your-org/platform-kit/internal/health"
)

type Response struct {
	Service string `json:"service"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 创建健康检查器
	checker := health.NewChecker()

	// 注册检查项（示例）
	checker.Register("database", func() error {
		// 这里检查数据库连接
		return nil
	})

	checker.Register("cache", func() error {
		// 这里检查缓存连接
		return nil
	})

	// 注册健康检查路由
	http.HandleFunc("/health", checker.LiveHandler())
	http.HandleFunc("/ready", checker.ReadyHandler())

	// 首页
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Response{
			Service: "demo-service",
			Status:  "running",
			Message: "Hello from demo-service!",
		})
	})

	// 业务接口
	http.HandleFunc("/api/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Response{
			Service: "demo-service",
			Status:  "ok",
			Message: "This is a demo API endpoint",
		})
	})

	slog.Info("starting demo-service", "port", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
