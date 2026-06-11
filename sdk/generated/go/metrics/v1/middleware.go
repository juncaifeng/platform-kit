package metricsv1

import (
	"net/http"
	"strconv"
	"time"
)

// HTTPMiddleware HTTP 指标中间件
func HTTPMiddleware(m *Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 活跃连接 +1
			m.ActiveConnections.Inc()
			defer m.ActiveConnections.Dec()

			// 记录请求体大小
			if r.ContentLength > 0 {
				m.HTTPRequestSize.WithLabelValues(r.Method, r.URL.Path).Observe(float64(r.ContentLength))
			}

			// 包装 ResponseWriter 以捕获状态码和响应大小
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r)

			duration := time.Since(start).Seconds()
			status := strconv.Itoa(rw.statusCode)

			// 记录指标
			m.HTTPRequestsTotal.WithLabelValues(r.Method, status, r.URL.Path).Inc()
			m.HTTPRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)

			if rw.written > 0 {
				m.HTTPResponseSize.WithLabelValues(r.Method, r.URL.Path).Observe(float64(rw.written))
			}
		})
	}
}

// responseWriter 包装 ResponseWriter 以捕获状态码和响应大小
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}
