# Platform Kit

微服务平台工具包 - 整合服务注册和标准 SDK。

## 核心特性

- **基础镜像模式**：业务服务基于平台基础镜像构建，自动获得注册能力
- **后台注册服务**：registerd 以后台守护进程方式运行，不影响业务服务
- **标准健康检查**：内置健康检查 SDK，统一 `/health` 和 `/ready` 接口
- **零侵入**：业务服务无需关心注册逻辑，只需实现健康检查接口

## 架构设计

```
┌─────────────────────────────────────────────────────────┐
│                    业务服务容器                          │
├─────────────────────────────────────────────────────────┤
│  entrypoint (PID 1)                                     │
│    ├── registerd (后台进程)                              │
│    │     ├── 注册到 Consul                              │
│    │     └── 心跳维护                                   │
│    └── demo-service (业务进程)                           │
│          ├── /health (健康检查)                          │
│          └── /api/v1/demo (业务接口)                     │
└─────────────────────────────────────────────────────────┘
```

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
# 公开仓库
go get github.com/juncaifeng/platform-kit@v1.0.0
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
