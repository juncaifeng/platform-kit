// Package metricsv1 提供 Prometheus 指标定义和 /metrics 端点
//
// 规范要求的指标:
//   - {prefix}_http_requests_total         Counter   HTTP 请求总数（按 method, status, path）
//   - {prefix}_http_request_duration_seconds Histogram 请求耗时分布
//   - {prefix}_http_request_size_bytes     Summary   请求体大小
//   - {prefix}_http_response_size_bytes    Summary   响应体大小
//   - {prefix}_active_connections          Gauge     当前活跃连接数
//   - {prefix}_event_published_total       Counter   事件发布总数（按 subject）
//   - {prefix}_event_consumed_total        Counter   事件消费总数（按 subject）
//   - {prefix}_event_publish_duration_seconds Histogram 事件发布耗时
//   - {prefix}_event_consume_duration_seconds Histogram 事件消费耗时
//   - {prefix}_registry_heartbeat_duration_seconds Histogram 注册中心心跳耗时
//   - go_goroutines                        Gauge     Go 协程数
//   - go_memstats_alloc_bytes              Gauge     内存分配
package metricsv1

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics 包含所有规范定义的指标
type Metrics struct {
	// HTTP 指标
	HTTPRequestsTotal       *prometheus.CounterVec
	HTTPRequestDuration     *prometheus.HistogramVec
	HTTPRequestSize         *prometheus.SummaryVec
	HTTPResponseSize        *prometheus.SummaryVec
	ActiveConnections       prometheus.Gauge

	// 事件指标
	EventPublishedTotal     *prometheus.CounterVec
	EventConsumedTotal      *prometheus.CounterVec
	EventPublishDuration    *prometheus.HistogramVec
	EventConsumeDuration    *prometheus.HistogramVec

	// 注册中心指标
	RegistryHeartbeatDuration *prometheus.HistogramVec
}

// New 创建指标实例
func New(prefix string) *Metrics {
	m := &Metrics{
		// HTTP 指标
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: prefix + "_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "status", "path"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    prefix + "_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		HTTPRequestSize: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       prefix + "_http_request_size_bytes",
				Help:       "HTTP request body size in bytes",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			[]string{"method", "path"},
		),
		HTTPResponseSize: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       prefix + "_http_response_size_bytes",
				Help:       "HTTP response body size in bytes",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			[]string{"method", "path"},
		),
		ActiveConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: prefix + "_active_connections",
				Help: "Number of active connections",
			},
		),

		// 事件指标
		EventPublishedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: prefix + "_event_published_total",
				Help: "Total number of events published",
			},
			[]string{"subject"},
		),
		EventConsumedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: prefix + "_event_consumed_total",
				Help: "Total number of events consumed",
			},
			[]string{"subject"},
		),
		EventPublishDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    prefix + "_event_publish_duration_seconds",
				Help:    "Event publish duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"subject"},
		),
		EventConsumeDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    prefix + "_event_consume_duration_seconds",
				Help:    "Event consume duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"subject"},
		),

		// 注册中心指标
		RegistryHeartbeatDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    prefix + "_registry_heartbeat_duration_seconds",
				Help:    "Registry heartbeat duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"registry"},
		),
	}

	return m
}

// Register 注册所有指标到 Prometheus
func (m *Metrics) Register() {
	prometheus.MustRegister(
		m.HTTPRequestsTotal,
		m.HTTPRequestDuration,
		m.HTTPRequestSize,
		m.HTTPResponseSize,
		m.ActiveConnections,
		m.EventPublishedTotal,
		m.EventConsumedTotal,
		m.EventPublishDuration,
		m.EventConsumeDuration,
		m.RegistryHeartbeatDuration,
	)
}

// Handler 返回 /metrics 端点处理器
func Handler() http.Handler {
	return promhttp.Handler()
}

// HandlerFunc 返回 /metrics 端点处理函数
func HandlerFunc() http.HandlerFunc {
	return promhttp.Handler().ServeHTTP
}

// CollectRuntimeMetrics 收集 Go 运行时指标
// go_goroutines, go_memstats_alloc_bytes 等由 prometheus 自动收集
func CollectRuntimeMetrics() {
	prometheus.MustRegister(prometheus.NewGoCollector())
}
