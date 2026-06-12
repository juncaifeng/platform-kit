# Platform Kit

微服务平台工具包 - Kit Sidecar 架构

## 架构设计

```
┌─────────────────────────────────────────────────────────┐
│                    业务服务容器                          │
├─────────────────────────────────────────────────────────┤
│  业务进程（Go/Python/Java/Node.js）                     │
│    └── SDK（客户端）──gRPC──▶ Kit Sidecar               │
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

## SDK 使用

### Go 服务

```bash
go get github.com/juncaifeng/platform-kit/sdk/go/kit/v1@v1.0.2
```

```go
package main

import (
    "context"
    "os"

    kitv1 "github.com/juncaifeng/platform-kit/sdk/generated/go/kit/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    // 连接 Kit Sidecar
    conn, _ := grpc.Dial(
        os.Getenv("KIT_SIDECAR_ADDR"),
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    defer conn.Close()

    client := kitv1.NewKitServiceClient(conn)

    // 注册服务
    client.Register(context.Background(), &kitv1.RegisterRequest{
        ServiceName: "order-service",
        Address:     os.Getenv("POD_IP"),
        Port:        8080,
    })

    // 发布事件
    client.PublishEvent(context.Background(), &kitv1.PublishEventRequest{
        Subject: "order.created.v1",
        Data:    []byte(`{"order_id":"123","amount":100}`),
    })
}
```

### Python 服务

```python
import grpc
import os
from kit.v1 import kit_pb2, kit_pb2_grpc

# 连接 Kit Sidecar
channel = grpc.insecure_channel(os.getenv("KIT_SIDECAR_ADDR"))
client = kit_pb2_grpc.KitServiceStub(channel)

# 注册服务
client.Register(kit_pb2.RegisterRequest(
    service_name="order-service",
    address=os.getenv("POD_IP"),
    port=8080
))

# 发布事件
client.PublishEvent(kit_pb2.PublishEventRequest(
    subject="order.created.v1",
    data=b'{"order_id":"123","amount":100}'
))
```

## 部署方式

### Docker Compose

```yaml
services:
  # Kit Sidecar
  order-service-kit:
    image: registry.cn-shanghai.aliyuncs.com/big_star/platform-kit:latest
    environment:
      - SERVICE_NAME=order-service
      - SERVICE_PORT=8080
      - REGISTRY_TYPE=consul
      - REGISTRY_ADDR=http://consul:8500
      - EVENT_BUS_TYPE=nats
      - EVENT_BUS_ADDR=nats://nats:4222

  # 业务服务
  order-service:
    build: ./examples/order-service
    environment:
      - KIT_SIDECAR_ADDR=order-service-kit:9090
```

### K8s（Pod 内 Sidecar）

```yaml
apiVersion: v1
kind: Pod
spec:
  containers:
    - name: order-service
      image: order-service:latest
      env:
        - name: KIT_SIDECAR_ADDR
          value: "localhost:9090"

    - name: kit-sidecar
      image: registry.cn-shanghai.aliyuncs.com/big_star/platform-kit:latest
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
├── cmd/kitd/                  # Kit Sidecar 主程序
├── internal/kit/              # Kit 内部实现
│   ├── registry/              # 服务注册
│   │   ├── registry.go        # 接口定义
│   │   ├── consul.go          # Consul 实现
│   │   └── k8s.go             # K8s 实现
│   └── event/                 # 事件总线
│       ├── event.go           # 接口定义
│       ├── nats.go            # NATS 实现
│       └── kafka.go           # Kafka 实现
├── proto/kit/v1/              # Kit Sidecar gRPC API
│   └── kit.proto
├── sdk/generated/             # 生成的客户端 SDK
│   ├── go/kit/v1/             # Go SDK
│   └── python/kit/v1/         # Python SDK
├── examples/                  # 示例服务
├── Dockerfile.kit             # Kit Sidecar 镜像
└── .github/workflows/         # CI/CD
    ├── generate-sdk.yml       # SDK 生成
    └── build-kit.yml          # Kit 镜像构建
```

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

## API 契约

### Kit Sidecar gRPC API

```protobuf
service KitService {
  // 服务注册
  rpc Register(RegisterRequest) returns (RegisterResponse);
  rpc Deregister(DeregisterRequest) returns (DeregisterResponse);

  // 健康检查
  rpc GetHealth(google.protobuf.Empty) returns (HealthResponse);
  rpc GetReadiness(google.protobuf.Empty) returns (ReadinessResponse);

  // 事件
  rpc PublishEvent(PublishEventRequest) returns (PublishEventResponse);
  rpc SubscribeEvent(SubscribeEventRequest) returns (stream EventMessage);
}
```

## 版本管理

| 版本 | 说明 |
|------|------|
| `v1.0.2` | 当前版本（SDK 自动生成） |
| `v1.x.x` | 每次 proto 变更自动递增 |

## 设计原则

1. **Sidecar 模式**：Kit 作为独立进程，与业务服务解耦
2. **多语言支持**：SDK 从 proto 生成，支持 Go/Python/Java/Node.js
3. **接口抽象**：所有组件通过接口定义，支持插件化替换
4. **配置驱动**：切换组件只需改配置，业务代码零改动
