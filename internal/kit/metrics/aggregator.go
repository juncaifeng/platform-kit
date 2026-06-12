// Package metrics 提供指标聚合功能
//
// 接收业务服务上报的指标，聚合后暴露 /metrics（Prometheus 格式）
package metrics

import (
	"fmt"
	"strings"
	"sync"
)

// Aggregator 指标聚合器
type Aggregator struct {
	mu      sync.RWMutex
	metrics map[string]*MetricEntry
}

// MetricEntry 指标条目
type MetricEntry struct {
	Name   string
	Type   string
	Help   string
	Values []*MetricValue
}

// MetricValue 指标值
type MetricValue struct {
	Value  float64
	Labels map[string]string
}

// NewAggregator 创建聚合器
func NewAggregator() *Aggregator {
	return &Aggregator{
		metrics: make(map[string]*MetricEntry),
	}
}

// Report 上报指标
func (a *Aggregator) Report(serviceName string, metrics []*Metric) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, m := range metrics {
		key := m.Name

		entry, exists := a.metrics[key]
		if !exists {
			entry = &MetricEntry{
				Name:   m.Name,
				Type:   m.Type,
				Help:   m.Help,
				Values: make([]*MetricValue, 0),
			}
			a.metrics[key] = entry
		}

		// 添加服务名标签
		labels := make(map[string]string)
		for k, v := range m.Labels {
			labels[k] = v
		}
		labels["service"] = serviceName

		entry.Values = append(entry.Values, &MetricValue{
			Value:  m.Value,
			Labels: labels,
		})
	}
}

// Metric 指标
type Metric struct {
	Name   string
	Type   string
	Value  float64
	Labels map[string]string
	Help   string
}

// GetMetrics 获取 Prometheus 格式的指标
func (a *Aggregator) GetMetrics() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var sb strings.Builder

	for _, entry := range a.metrics {
		// 写入 HELP
		if entry.Help != "" {
			sb.WriteString(fmt.Sprintf("# HELP %s %s\n", entry.Name, entry.Help))
		}

		// 写入 TYPE
		sb.WriteString(fmt.Sprintf("# TYPE %s %s\n", entry.Name, entry.Type))

		// 写入指标值
		for _, v := range entry.Values {
			labels := formatLabels(v.Labels)
			sb.WriteString(fmt.Sprintf("%s%s %g\n", entry.Name, labels, v.Value))
		}
	}

	return sb.String()
}

// formatLabels 格式化标签
func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}

	var parts []string
	for k, v := range labels {
		parts = append(parts, fmt.Sprintf("%s=\"%s\"", k, v))
	}

	return "{" + strings.Join(parts, ",") + "}"
}
