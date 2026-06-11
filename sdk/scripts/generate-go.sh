#!/bin/bash
# 生成 Go SDK

set -e

echo "=== 生成 Go SDK ==="

# 检查依赖
if ! command -v protoc &> /dev/null; then
    echo "错误: 请安装 protobuf"
    echo "  brew install protobuf"
    exit 1
fi

if ! command -v protoc-gen-go &> /dev/null; then
    echo "错误: 请安装 Go protobuf 插件"
    echo "  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
    echo "  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    exit 1
fi

# 创建输出目录
mkdir -p generated/go

# 生成健康检查
echo "生成 health/v1..."
protoc \
    --proto_path=proto \
    --proto_path=$(go env GOPATH)/src \
    --go_out=generated/go \
    --go_opt=paths=source_relative \
    --go-grpc_out=generated/go \
    --go-grpc_opt=paths=source_relative \
    proto/health/v1/health.proto

# 生成监控指标
echo "生成 metrics/v1..."
protoc \
    --proto_path=proto \
    --proto_path=$(go env GOPATH)/src \
    --go_out=generated/go \
    --go_opt=paths=source_relative \
    --go-grpc_out=generated/go \
    --go-grpc_opt=paths=source_relative \
    proto/metrics/v1/metrics.proto

# 生成服务注册
echo "生成 service/v1..."
protoc \
    --proto_path=proto \
    --proto_path=$(go env GOPATH)/src \
    --go_out=generated/go \
    --go_opt=paths=source_relative \
    --go-grpc_out=generated/go \
    --go-grpc_opt=paths=source_relative \
    proto/service/v1/service.proto

echo "=== Go SDK 生成完成 ==="
