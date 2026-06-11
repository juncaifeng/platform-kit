#!/bin/bash
# 生成所有 SDK

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== 生成所有 SDK ==="

"$SCRIPT_DIR/generate-go.sh"
"$SCRIPT_DIR/generate-python.sh"

echo "=== 所有 SDK 生成完成 ==="
