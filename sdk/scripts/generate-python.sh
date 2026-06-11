#!/bin/bash
# 生成 Python SDK

set -e

echo "=== 生成 Python SDK ==="

# 检查依赖
if ! command -v python3 &> /dev/null; then
    echo "错误: 请安装 Python3"
    exit 1
fi

# 创建输出目录
mkdir -p generated/python/platform_kit

# 创建 __init__.py
touch generated/python/platform_kit/__init__.py

# 生成健康检查
echo "生成 health/v1..."
python3 -m grpc_tools.protoc \
    --proto_path=proto \
    --python_out=generated/python \
    --grpc_python_out=generated/python \
    proto/health/v1/health.proto

# 生成监控指标
echo "生成 metrics/v1..."
python3 -m grpc_tools.protoc \
    --proto_path=proto \
    --python_out=generated/python \
    --grpc_python_out=generated/python \
    proto/metrics/v1/metrics.proto

# 生成服务注册
echo "生成 service/v1..."
python3 -m grpc_tools.protoc \
    --proto_path=proto \
    --python_out=generated/python \
    --grpc_python_out=generated/python \
    proto/service/v1/service.proto

echo "=== Python SDK 生成完成 ==="
