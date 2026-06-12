// Package health 提供健康检查聚合功能
//
// 接收业务服务上报的健康状态，聚合后暴露 /health, /ready
package health

import (
	"sync"
	"time"
)

// Aggregator 健康状态聚合器
type Aggregator struct {
	mu       sync.RWMutex
	services map[string]*ServiceHealth
}

// ServiceHealth 服务健康状态
type ServiceHealth struct {
	Status     string            `json:"status"`
	Checks     map[string]string `json:"checks"`
	LastReport time.Time         `json:"last_report"`
}

// NewAggregator 创建聚合器
func NewAggregator() *Aggregator {
	return &Aggregator{
		services: make(map[string]*ServiceHealth),
	}
}

// Report 上报健康状态
func (a *Aggregator) Report(serviceName string, status string, checks map[string]string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.services[serviceName] = &ServiceHealth{
		Status:     status,
		Checks:     checks,
		LastReport: time.Now(),
	}
}

// GetHealth 获取聚合后的健康状态
func (a *Aggregator) GetHealth() (string, map[string]*ServiceHealth) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	services := make(map[string]*ServiceHealth)
	overallStatus := "healthy"

	for name, health := range a.services {
		// 检查是否超时（30 秒未上报视为不健康）
		if time.Since(health.LastReport) > 30*time.Second {
			services[name] = &ServiceHealth{
				Status:     "unhealthy",
				Checks:     health.Checks,
				LastReport: health.LastReport,
			}
			overallStatus = "unhealthy"
		} else {
			services[name] = health
			if health.Status != "healthy" {
				overallStatus = "unhealthy"
			}
		}
	}

	return overallStatus, services
}

// GetReadiness 获取就绪状态
func (a *Aggregator) GetReadiness() (bool, map[string]bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	services := make(map[string]bool)
	allReady := true

	for name, health := range a.services {
		// 检查是否超时
		if time.Since(health.LastReport) > 30*time.Second {
			services[name] = false
			allReady = false
		} else {
			ready := health.Status == "healthy"
			services[name] = ready
			if !ready {
				allReady = false
			}
		}
	}

	return allReady, services
}
