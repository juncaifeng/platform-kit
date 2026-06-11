# Platform Kit SDK

微服务平台 SDK - API-First 方式定义契约并生成多语言 SDK。

## 架构设计

```
┌─────────────────────────────────────────────────────────┐
│                    API 契约层                            │
├─────────────────────────────────────────────────────────┤
│  proto/                      openapi/                   │
│  ├── health/v1/              └── platform-kit.yaml      │
│  ├── metrics/v1/                                        │
│  └── service/v1/                                        │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼ 代码生成
┌─────────────────────────────────────────────────────────┐
│                    生成的 SDK                            │
├─────────────────────────────────────────────────────────┤
│  generated/                                             │
│  ├── go/                                                │
│  │   ├── health/v1/                                     │
│  │   ├── metrics/v1/                                    │
│  │   └── service/v1/                                    │
│  └── python/                                            │
│      ├── health/v1/                                     │
│      ├── metrics/v1/                                    │
│      └── service/v1/                                    │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼ 使用
┌─────────────────────────────────────────────────────────┐
│                    业务服务                              │
├─────────────────────────────────────────────────────────┤
│  - 实现 gRPC 服务接口                                    │
│  - 使用 HTTP 中间件                                      │
│  - 自动获得健康检查、监控等能力                           │
└─────────────────────────────────────────────────────────┘
```

## API 契约

### Proto 定义

| 服务 | 用途 | 接口 |
|------|------|------|
| `HealthService` | 健康检查 | `Live`, `Ready` |
| `MetricsService` | 监控指标 | `GetMetrics`, `GetServiceInfo` |
| `ServiceRegistry` | 服务注册 | `Register`, `Deregister`, `Heartbeat` |

### OpenAPI 规范

| 接口 | 用途 | 调用方 |
|------|------|--------|
| `/health` | 存活检查 | Consul / K8s / 负载均衡 |
| `/ready` | 就绪检查 | K8s / 滚动更新 |
| `/metrics` | 监控指标 | Prometheus |
| `/config` | 配置查看 | 运维调试 |

## 生成 SDK

### 前置条件

```bash
# 安装 protoc
brew install protobuf

# 安装 Go 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 安装 Python 工具
pip install grpcio grpcio-tools
```

### 生成命令

```bash
# 生成所有 SDK
./scripts/generate-all.sh

# 单独生成
./scripts/generate-go.sh
./scripts/generate-python.sh
```

## 使用示例

### Go

```go
package main

import (
    "context"
    "log"
    "net"

    healthv1 "github.com/your-org/platform-kit/sdk/generated/go/health/v1"
    "google.golang.org/grpc"
    "google.golang.org/protobuf/types/known/emptypb"
)

// 实现健康检查服务
type healthServer struct {
    healthv1.UnimplementedHealthServiceServer
}

func (s *healthServer) Live(ctx context.Context, in *emptypb.Empty) (*healthv1.LiveResponse, error) {
    return &healthv1.LiveResponse{
        Status: "healthy",
    }, nil
}

func (s *healthServer) Ready(ctx context.Context, in *emptypb.Empty) (*healthv1.ReadyResponse, error) {
    return &healthv1.ReadyResponse{
        Ready: true,
    }, nil
}

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    s := grpc.NewServer()
    healthv1.RegisterHealthServiceServer(s, &healthServer{})

    log.Println("gRPC server listening on :50051")
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
```

### Python

```python
import grpc
from concurrent import futures
from sdk.generated.python.health.v1 import health_pb2, health_pb2_grpc

class HealthService(health_pb2_grpc.HealthServiceServicer):
    def Live(self, request, context):
        return health_pb2.LiveResponse(status="healthy")
    
    def Ready(self, request, context):
        return health_pb2.ReadyResponse(ready=True)

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    health_pb2_grpc.add_HealthServiceServicer_to_server(HealthService(), server)
    server.add_insecure_port('[::]:50051')
    server.start()
    server.wait_for_termination()

if __name__ == '__main__':
    serve()
```

## 目录结构

```
sdk/
├── proto/                      # API 契约定义
│   ├── buf.yaml                # Buf 配置
│   ├── health/v1/              # 健康检查接口
│   ├── metrics/v1/             # 监控指标接口
│   └── service/v1/             # 服务注册接口
├── openapi/                    # OpenAPI 规范
│   └── platform-kit.yaml
├── scripts/                    # 代码生成脚本
│   ├── generate-go.sh
│   ├── generate-python.sh
│   └── generate-all.sh
└── generated/                  # 生成的 SDK
    ├── go/
    └── python/
```

## 扩展指南

### 添加新接口

1. 在 `proto/` 目录添加新的 `.proto` 文件
2. 运行 `./scripts/generate-all.sh` 生成代码
3. 在业务服务中实现新接口

### 添加新语言

1. 在 `scripts/` 目录添加新的生成脚本
2. 更新 `generate-all.sh` 调用新脚本
3. 在 `generated/` 目录查看生成结果

### 版本管理

- Proto 文件使用 `v1`, `v2` 版本目录
- 不兼容变更需要创建新版本目录
- 旧版本保持向后兼容
