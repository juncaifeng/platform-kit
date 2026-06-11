# Platform Kit

微服务平台工具包 - Kit Sidecar 架构

## 架构设计

```
┌─────────────────────────────────────────────────────────┐
│                    业务服务容器                          │
├─────────────────────────────────────────────────────────┤
│  业务进程（Go/Python/Java/Node.js）                     │
│    └── 调用 Kit Sidecar gRPC API                       │
├─────────────────────────────────────────────────────────┤
│  Kit Sidecar 进程                                       │
│    ├── /health, /ready 端点（自动注册）                 │
│    ├── 服务注册（Consul/CloudMap/K8s）                  │
│    ├── 事件发布/消费（NATS/Kafka）                      │
│    ├── 指标采集（/metrics）                             │
│    └── gRPC API（:9090）供业务进程调用                  │
└─────────────────────────────────────────────────────────┘
```

## 核心优势

| 优势 | 说明 |
|------|------|
| **多语言支持** | 业务服务可用任何语言（Go/Python/Java/Node.js） |
| **统一架构** | 所有能力通过 Kit Sidecar 提供 |
| **业务无感知** | 业务代码只关注业务逻辑 |
| **组件可替换** | NATS→Kafka、Consul→CloudMap 只需改配置 |

## 业务开发者使用

### Go 服务

```go
import "github.com/juncaifeng/platform-kit/sdk/go/kit"

// 连接 Kit Sidecar
client, _ := kit.New(&kit.Config{
    KitAddr: os.Getenv("KIT_SIDECAR_ADDR"),
})
defer client.Close()

// 注册服务
client.Register(ctx, &kit.ServiceInfo{
    Name: "order-service",
    Port: 8080,
})

// 发布事件
client.PublishEvent(ctx, "order.created.v1", data)
```

### Python 服务

```python
from kit_sdk import KitClient

# 连接 Kit Sidecar
kit = KitClient(os.getenv("KIT_SIDECAR_ADDR"))

# 注册服务
kit.register(name="order-service", port=8080)

# 发布事件
kit.publish_event("order.created.v1", data)
```

## 部署方式

### Docker Compose

```yaml
services:
  # Kit Sidecar
  order-service-kit:
    build:
      context: .
      dockerfile: Dockerfile.kit
    environment:
      - SERVICE_NAME=order-service
      - REGISTRY_TYPE=consul
      - REGISTRY_ADDR=http://consul:8500
      - EVENT_BUS_TYPE=nats
      - EVENT_BUS_ADDR=nats://nats:4222

  # 业务服务
  order-service:
    build: ./examples/order-service
    environment:
      - KIT_SIDECAR_ADDR=order-service-kit:9090
    depends_on:
      - order-service-kit
```

### K8s（Pod 内 Sidecar）

```yaml
apiVersion: v1
kind: Pod
spec:
  containers:
    # 业务容器
    - name: order-service
      image: order-service:latest
      env:
        - name: KIT_SIDECAR_ADDR
          value: "localhost:9090"
    
    # Kit Sidecar 容器
    - name: kit-sidecar
      image: platform-kit:latest
      env:
        - name: SERVICE_NAME
          value: "order-service"
        - name: REGISTRY_TYPE
          value: "consul"
```

## 组件切换

只需修改 Kit Sidecar 配置，业务代码零改动：

```bash
# 切换事件总线：NATS → Kafka
EVENT_BUS_TYPE=kafka
EVENT_BUS_ADDR=kafka:9092

# 切换注册中心：Consul → K8s Service
REGISTRY_TYPE=k8s-service
```

## 项目结构

```
platform-kit/
├── cmd/
│   └── kitd/                  # Kit Sidecar 主程序
├── internal/
│   └── kit/
│       ├── registry/          # 服务注册实现
│       │   ├── registry.go    # 接口定义
│       │   ├── consul.go      # Consul 实现
│       │   └── k8s.go         # K8s 实现
│       └── event/             # 事件总线实现
│           ├── event.go       # 接口定义
│           ├── nats.go        # NATS 实现
│           └── kafka.go       # Kafka 实现
├── proto/
│   └── kit/v1/                # Kit Sidecar gRPC API
│       └── kit.proto
├── sdk/                       # 客户端 SDK（多语言）
│   └── go/
│       └── kit/               # Go SDK
├── examples/                  # 示例服务
├── Dockerfile.kit             # Kit Sidecar 镜像
└── docker-compose.kit.yml     # 部署示例
```

## SDK vs Kit

| 组件 | 定位 | 使用方式 |
|------|------|----------|
| **SDK** | 客户端库（多语言） | 业务服务引入依赖 |
| **Kit** | Sidecar 服务 | 独立进程运行 |

## 环境变量

### Kit Sidecar

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `SERVICE_NAME` | - | 服务名称 |
| `SERVICE_PORT` | 8080 | 服务端口 |
| `REGISTRY_TYPE` | consul | 注册中心类型 |
| `REGISTRY_ADDR` | http://localhost:8500 | 注册中心地址 |
| `EVENT_BUS_TYPE` | nats | 事件总线类型 |
| `EVENT_BUS_ADDR` | nats://localhost:4222 | 事件总线地址 |
| `KIT_GRPC_PORT` | 9090 | gRPC API 端口 |
| `KIT_HTTP_PORT` | 9091 | HTTP 端口（健康检查） |

### 业务服务

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `KIT_SIDECAR_ADDR` | localhost:9090 | Kit Sidecar 地址 |
