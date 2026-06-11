# Platform Kit

微服务平台工具包 - 三层架构（Infra / Kit / Dev）

## 架构设计

```
┌─────────────────────────────────────────────────────────┐
│                    Dev 业务开发层                        │
│  业务 Handler、领域模型、业务规则                        │
│  开发者只写业务代码，不感知基础设施                       │
└─────────────────────────────────────────────────────────┘
                          │
                          │ go get 引入
                          ▼
┌─────────────────────────────────────────────────────────┐
│                    Kit 脚手架层                          │
│  ├── app/          应用启动框架                         │
│  ├── config/       配置加载                             │
│  ├── event/        事件封装（NATS/Kafka）               │
│  ├── registry/     服务注册（Consul/CloudMap/K8s）      │
│  └── transport/    传输层（HTTP/gRPC）                  │
└─────────────────────────────────────────────────────────┘
                          │
                          │ 自动处理
                          ▼
┌─────────────────────────────────────────────────────────┐
│                    Infra 基础设施层                      │
│  APISIX 网关、Consul、NATS、服务注册同步器、CI/CD       │
│  运维管理，服务无感知                                    │
└─────────────────────────────────────────────────────────┘
```

## 核心特性

- **插件化架构**：组件可替换（NATS→Kafka、Consul→CloudMap）
- **配置驱动**：切换组件只需改配置，业务代码零改动
- **标准接口**：`/health`、`/ready`、`/metrics` 自动实现
- **优雅停止**：SIGTERM 信号自动处理
- **事件封装**：统一的事件发布/消费接口

## 业务开发者使用

### 第一步：引入依赖

```bash
go get github.com/juncaifeng/platform-kit/kit@v1.0.1
```

### 第二步：编写业务代码

```go
package main

import (
    "context"
    "github.com/juncaifeng/platform-kit/kit/app"
    "github.com/juncaifeng/platform-kit/kit/config"
    "github.com/juncaifeng/platform-kit/kit/event"
    httpTransport "github.com/juncaifeng/platform-kit/kit/transport/http"
)

func main() {
    // 1. 加载配置
    cfg, _ := config.Load("")

    // 2. 创建事件发布器
    eventPub, _ := event.NewPublisher(event.EventBusConfig{
        Type: cfg.EventBus.Type,
        NATS: &event.NATSConfig{Address: cfg.EventBus.NATS.Address},
    })

    // 3. 创建应用
    app := app.New(
        app.WithConfig(cfg),
        app.WithHTTPServer(":8080"),
        app.WithEventPublisher(eventPub),
    )

    // 4. 注册业务路由
    httpServer := app.HTTPServer()
    httpServer.HandleFunc("POST /api/v1/orders", createOrderHandler)

    // 5. 运行（Kit 自动处理健康检查、优雅停止）
    app.Run(context.Background())
}
```

### 第三步：配置环境变量

```bash
# 服务配置
SERVICE_NAME=order-service
SERVICE_PORT=8080

# 注册中心（可选切换）
REGISTRY_TYPE=consul          # consul / cloudmap / k8s-service
REGISTRY_ADDR=consul:8500

# 事件总线（可选切换）
EVENT_BUS_TYPE=nats           # nats / kafka
NATS_ADDRESS=nats://nats:4222
```

## Kit 能力清单

| 能力 | 说明 | 开发者感知 |
|------|------|-----------|
| 标准接口 | `/health`、`/ready`、`/metrics` 自动实现 | ❌ 无需编写 |
| 优雅停止 | SIGTERM 信号自动处理 | ❌ 无需编写 |
| 事件封装 | 统一的发布/消费接口 | ✅ 调用 SDK |
| 服务注册 | 自动注册到注册中心 | ❌ 无需编写 |
| 配置加载 | YAML + 环境变量 | ❌ 自动加载 |

## 组件切换

只需修改配置，业务代码零改动：

```yaml
# 切换事件总线：NATS → Kafka
eventBus:
  type: kafka
  kafka:
    brokers: ["kafka:9092"]

# 切换注册中心：Consul → Cloud Map
registry:
  type: cloudmap
  cloudmap:
    namespace: prod.internal
    region: ap-northeast-1
```

## 项目结构

```
platform-kit/
├── kit/                        # 🆕 脚手架实现层
│   ├── app/                   # 应用启动框架
│   ├── config/                # 配置加载
│   ├── event/                 # 事件封装
│   │   ├── event.go           # 接口定义
│   │   ├── nats.go            # NATS 实现
│   │   └── kafka.go           # Kafka 实现（占位）
│   ├── registry/              # 服务注册
│   │   ├── registry.go        # 接口定义
│   │   ├── consul.go          # Consul 实现
│   │   ├── cloudmap.go        # Cloud Map 实现（占位）
│   │   └── k8s.go             # K8s 实现
│   ├── transport/             # 传输层
│   │   └── http/              # HTTP 服务器
│   └── middleware/            # 中间件
├── sdk/                        # 契约层（Proto/OpenAPI）
├── cmd/                        # 基础镜像
│   ├── registerd/             # 后台注册守护进程
│   └── entrypoint/            # 入口包装器
├── examples/                   # 示例服务
│   └── order-service/         # 订单服务示例
└── Dockerfile.base             # 基础镜像
```

## 设计原则

1. **接口抽象**：所有组件通过接口定义，支持插件化替换
2. **配置驱动**：切换组件只需改配置，业务代码零改动
3. **约定优于配置**：合理的默认值，开箱即用
4. **关注点分离**：开发写业务，运维改配置，平台维护 Kit

## 快速开始

### 1. 构建基础镜像

```bash
docker build -f Dockerfile.base -t platform-kit:latest .
```

### 2. 业务服务使用基础镜像

```dockerfile
FROM platform-kit:latest

WORKDIR /app
COPY ./my-service .

ENV SERVICE_NAME=my-service
ENV SERVICE_PORT=8080
ENV REGISTRY_ADDR=http://consul:8500

CMD ["./my-service"]
```

### 3. 启动服务

```bash
docker-compose up -d
```

## 环境变量

| 变量 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `SERVICE_NAME` | **是** | - | 服务名称 |
| `SERVICE_PORT` | **是** | - | 服务端口 |
| `REGISTER_MODE` | 否 | `daemon` | 运行模式：`daemon`/`once` |
| `REGISTRY_TYPE` | 否 | `consul` | 注册中心类型 |
| `REGISTRY_ADDR` | 否 | `http://localhost:8500` | 注册中心地址 |
| `SERVICE_ADDR` | 否 | 自动检测 | 服务地址 |
| `HEALTH_CHECK_PATH` | 否 | `/health` | 健康检查路径 |
| `HEALTH_CHECK_INTERVAL` | 否 | `10s` | 健康检查间隔 |
| `GATEWAY_EXPOSED` | 否 | `false` | 是否暴露到网关 |
| `GATEWAY_PREFIX` | 否 | - | 网关路由前缀 |
| `EVENTS_PRODUCED` | 否 | - | 产生的事件 |
| `EVENTS_CONSUMED` | 否 | - | 消费的事件 |

## SDK 使用

### 安装 SDK

```bash
# 使用固定版本（推荐）
go get github.com/juncaifeng/platform-kit/sdk/go/health/v1@v1.0.1
go get github.com/juncaifeng/platform-kit/sdk/go/metrics/v1@v1.0.1
go get github.com/juncaifeng/platform-kit/sdk/go/service/v1@v1.0.1

# 使用最新版本（sdk 分支）
go get github.com/juncaifeng/platform-kit/sdk/go/health/v1@sdk
```

### 版本管理

| 版本 | 说明 |
|------|------|
| `v1.0.0` | 初始版本 |
| `v1.0.1` | 自动递增（CI 生成） |
| `v1.x.x` | 每次 proto 变更自动递增 |

### SDK 更新流程

```
修改 proto/openapi → 推送 master → CI 自动生成 → 自动打 tag → go get @版本号
```

### 标准接口规范

| 接口 | 用途 | 调用方 | 响应格式 | 是否必须 |
|------|------|--------|----------|----------|
| `/health` | 存活检查 | Consul / K8s / 负载均衡 | `{"status":"healthy"}` | 必须 |
| `/ready` | 就绪检查 | K8s / 滚动更新 | `{"ready":true}` | 必须 |
| `/metrics` | 监控指标 | Prometheus / 监控系统 | Prometheus 格式 | 必须 |
| `/debug/pprof` | 性能分析 | 开发/运维手动 | Go pprof 页面 | 可选 |
| `/config` | 配置查看 | 运维调试 | `{"config":{...}}` | 可选 |

### Prometheus 指标

| 指标名 | 类型 | 说明 |
|--------|------|------|
| `{prefix}_http_requests_total` | Counter | HTTP 请求总数（按 method, status, path） |
| `{prefix}_http_request_duration_seconds` | Histogram | 请求耗时分布 |
| `{prefix}_http_request_size_bytes` | Summary | 请求体大小 |
| `{prefix}_http_response_size_bytes` | Summary | 响应体大小 |
| `{prefix}_active_connections` | Gauge | 当前活跃连接数 |
| `{prefix}_event_published_total` | Counter | 事件发布总数（按 subject） |
| `{prefix}_event_consumed_total` | Counter | 事件消费总数（按 subject） |
| `{prefix}_event_publish_duration_seconds` | Histogram | 事件发布耗时 |
| `{prefix}_event_consume_duration_seconds` | Histogram | 事件消费耗时 |
| `{prefix}_registry_heartbeat_duration_seconds` | Histogram | 注册中心心跳耗时 |
| `go_goroutines` | Gauge | Go 协程数（runtime 自带） |
| `go_memstats_alloc_bytes` | Gauge | 内存分配（runtime 自带） |

### 快速接入示例

```go
package main

import (
    "net/http"

    metricsv1 "github.com/juncaifeng/platform-kit/sdk/generated/go/metrics/v1"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    // 初始化指标
    m := metricsv1.New("myapp")
    m.Register()
    metricsv1.CollectRuntimeMetrics()

    // 注册 /metrics 端点
    http.Handle("/metrics", promhttp.Handler())

    // 使用指标中间件
    handler := metricsv1.HTTPMiddleware(m)(http.DefaultServeMux)

    http.ListenAndServe(":8080", handler)
}
```

### 健康检查接入

```go
package main

import (
    "net/http"

    "github.com/juncaifeng/platform-kit/internal/health"
)

func main() {
    checker := health.NewChecker()

    // 注册检查项
    checker.Register("database", func() error {
        // 检查数据库连接
        return nil
    })

    checker.Register("redis", func() error {
        // 检查 Redis 连接
        return nil
    })

    // 注册端点
    http.HandleFunc("/health", checker.LiveHandler())
    http.HandleFunc("/ready", checker.ReadyHandler())

    http.ListenAndServe(":8080", nil)
}
```

### 生成 SDK（可选）

如需修改接口定义并重新生成：

```bash
cd sdk

# 生成所有 SDK
./scripts/generate-all.sh

# 单独生成 Go SDK
./scripts/generate-go.sh
```

## 与旧方案对比

| 维度 | 旧方案（Sidecar） | 新方案（基础镜像） |
|------|-------------------|-------------------|
| 容器数量 | 每个服务 2 个容器 | 每个服务 1 个容器 |
| 资源占用 | 较高 | 较低 |
| 配置复杂度 | 需要配置两个容器 | 只需配置一个容器 |
| 网络通信 | 容器间通信 | 进程间通信 |
| 部署复杂度 | 较高 | 较低 |

## 项目结构

```
platform-kit/
├── cmd/
│   ├── registerd/          # 后台注册守护进程
│   └── entrypoint/         # 入口包装器
├── internal/
│   ├── config/             # 配置模块
│   ├── registry/           # 注册模块
│   └── health/             # 健康检查模块
├── sdk/                    # SDK 模块（API-First）
│   ├── proto/              # Proto 定义（gRPC）
│   │   ├── health/v1/      # 健康检查接口
│   │   ├── metrics/v1/     # 监控指标接口
│   │   └── service/v1/     # 服务注册接口
│   ├── openapi/            # OpenAPI 规范（REST）
│   ├── scripts/            # 代码生成脚本
│   └── generated/          # 生成的 SDK
│       ├── go/
│       └── python/
├── examples/
│   └── demo-service/       # 示例服务
├── Dockerfile.base         # 基础镜像
└── docker-compose.yml      # 演示配置
```

## 构建和运行

```bash
# 构建基础镜像
docker build -f Dockerfile.base -t platform-kit:latest .

# 构建示例服务
docker build -t demo-service:latest ./examples/demo-service

# 启动所有服务
docker-compose up -d

# 查看注册的服务
curl http://localhost:8500/v1/agent/services

# 访问示例服务
curl http://localhost:8081/
curl http://localhost:8081/health
```

## SDK 分发

### 发布流程

```bash
# 1. 推送到 GitHub
git add .
git commit -m "feat: release v1.0.0"
git push origin main

# 2. 打版本 Tag
git tag v1.0.0
git push origin v1.0.0

# 3. 业务服务引入
go get github.com/juncaifeng/platform-kit@v1.0.0
```

### 版本管理

| 版本 | 说明 |
|------|------|
| `v1.x.x` | 稳定版本，向后兼容 |
| `v0.x.x` | 开发中版本，可能有破坏性变更 |
| `v2.x.x` | 下一个大版本 |

## 设计原则

1. **API-First**：接口定义（Proto/OpenAPI）优先，SDK 生成
2. **单一职责**：每个服务只需关注自己的业务逻辑
3. **平台统一**：注册、健康检查等由平台统一处理
4. **零侵入**：业务服务无需修改代码即可获得注册能力
5. **资源优化**：避免 Sidecar 模式的资源浪费
