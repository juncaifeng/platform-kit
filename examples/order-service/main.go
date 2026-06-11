// 订单服务示例 - 展示业务开发者如何使用 Kit
//
// 开发者只需关注:
// 1. 业务 Handler
// 2. 领域模型
// 3. 业务逻辑
//
// 不需要关注:
// - 健康检查 (/health, /ready)
// - 监控指标 (/metrics)
// - 服务注册
// - 优雅停止
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/juncaifeng/platform-kit/kit/app"
	"github.com/juncaifeng/platform-kit/kit/config"
	"github.com/juncaifeng/platform-kit/kit/event"
	httpTransport "github.com/juncaifeng/platform-kit/kit/transport/http"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// 2. 创建事件发布器
	eventPub, err := event.NewPublisher(event.EventBusConfig{
		Type: cfg.EventBus.Type,
		NATS: &event.NATSConfig{
			Address:   cfg.EventBus.NATS.Address,
			JetStream: cfg.EventBus.NATS.JetStream,
		},
	})
	if err != nil {
		slog.Error("failed to create event publisher", "error", err)
		os.Exit(1)
	}

	// 3. 创建应用
	application := app.New(
		app.WithConfig(cfg),
		app.WithHTTPServer(":8080"),
		app.WithEventPublisher(eventPub),
	)

	// 4. 注册业务路由（开发者只需关注这部分）
	httpServer := application.HTTPServer()
	registerRoutes(httpServer, eventPub)

	// 5. 运行应用（Kit 自动处理健康检查、优雅停止等）
	if err := application.Run(context.Background()); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}

// registerRoutes 注册业务路由
func registerRoutes(server *httpTransport.Server, eventPub event.Publisher) {
	// 创建 Handler
	orderHandler := &OrderHandler{
		eventPub: eventPub,
	}

	// 注册路由
	server.HandleFunc("POST /api/v1/orders", orderHandler.CreateOrder)
	server.HandleFunc("GET /api/v1/orders/{id}", orderHandler.GetOrder)
}

// OrderHandler 订单处理器
type OrderHandler struct {
	eventPub event.Publisher
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Amount    float64 `json:"amount"`
}

// CreateOrderResponse 创建订单响应
type CreateOrderResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

// CreateOrder 创建订单
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// 业务逻辑
	orderID := generateOrderID()

	// 发布事件（不关心底层是 NATS 还是 Kafka）
	eventData, _ := json.Marshal(map[string]interface{}{
		"order_id":    orderID,
		"product_id":  req.ProductID,
		"quantity":    req.Quantity,
		"amount":      req.Amount,
	})

	if err := h.eventPub.Publish(r.Context(), "order.created.v1", eventData); err != nil {
		slog.Error("failed to publish event", "error", err)
		// 不影响主流程，记录日志即可
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateOrderResponse{
		OrderID: orderID,
		Status:  "created",
	})
}

// GetOrder 获取订单
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("id")

	// 业务逻辑：查询订单
	order := map[string]interface{}{
		"order_id": orderID,
		"status":   "created",
		"amount":   100.00,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func generateOrderID() string {
	return "ORD-" + time.Now().Format("20060102150405")
}
