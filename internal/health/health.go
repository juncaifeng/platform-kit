package health

import (
	"encoding/json"
	"net/http"
	"sync"
)

// Status 健康状态
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// LiveResponse 存活检查响应 - /health
type LiveResponse struct {
	Status Status            `json:"status"`
	Errors map[string]string `json:"errors,omitempty"`
}

// ReadyResponse 就绪检查响应 - /ready
type ReadyResponse struct {
	Ready  bool              `json:"ready"`
	Errors map[string]string `json:"errors,omitempty"`
}

// Checker 健康检查器
type Checker struct {
	mu     sync.RWMutex
	checks map[string]CheckFunc
}

// CheckFunc 检查函数
type CheckFunc func() error

// NewChecker 创建健康检查器
func NewChecker() *Checker {
	return &Checker{
		checks: make(map[string]CheckFunc),
	}
}

// Register 注册检查项
func (c *Checker) Register(name string, check CheckFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks[name] = check
}

// LiveHandler 存活检查处理器 - /health
// 响应格式: {"status": "healthy|unhealthy|degraded"}
func (c *Checker) LiveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.mu.RLock()
		defer c.mu.RUnlock()

		response := LiveResponse{
			Status: StatusHealthy,
			Errors: make(map[string]string),
		}

		for name, check := range c.checks {
			if err := check(); err != nil {
				response.Status = StatusUnhealthy
				response.Errors[name] = err.Error()
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if response.Status != StatusHealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		json.NewEncoder(w).Encode(response)
	}
}

// ReadyHandler 就绪检查处理器 - /ready
// 响应格式: {"ready": true|false}
func (c *Checker) ReadyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.mu.RLock()
		defer c.mu.RUnlock()

		response := ReadyResponse{
			Ready:  true,
			Errors: make(map[string]string),
		}

		for name, check := range c.checks {
			if err := check(); err != nil {
				response.Ready = false
				response.Errors[name] = err.Error()
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if !response.Ready {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		json.NewEncoder(w).Encode(response)
	}
}

// Handler 返回标准健康检查处理器（存活检查）
func (c *Checker) Handler() http.HandlerFunc {
	return c.LiveHandler()
}
