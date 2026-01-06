#!/bin/bash

# Baseline 插件快速测试脚本
# 自动完成：编译插件 -> 准备目录 -> 运行测试

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# 从 tests/e2e/baseline 向上三级到 agent 根目录
AGENT_DIR="$(dirname "$(dirname "$(dirname "$SCRIPT_DIR")")")"
BASELINE_DIR="$AGENT_DIR/business_plugins/baseline"
PLUGIN_DIR="/tmp/plugin/baseline"
TEST_DIR="$SCRIPT_DIR"

echo -e "${GREEN}=== Baseline Plugin Test Script ===${NC}"
echo ""

# 步骤 1: 编译 baseline 插件
echo -e "${YELLOW}[Step 1/3] 编译 baseline 插件...${NC}"
cd "$BASELINE_DIR"

if [ ! -f "main.go" ]; then
    echo -e "${RED}错误: 找不到 main.go 文件${NC}"
    exit 1
fi

echo "  执行: go mod tidy"
go mod tidy > /dev/null 2>&1 || true

echo "  执行: go build -o baseline main.go"
if go build -o baseline main.go; then
    echo -e "${GREEN}  ✓ Baseline 插件编译成功${NC}"
else
    echo -e "${RED}  ✗ Baseline 插件编译失败${NC}"
    exit 1
fi

# 检查编译结果
if [ ! -f "baseline" ]; then
    echo -e "${RED}错误: 编译后未找到 baseline 可执行文件${NC}"
    exit 1
fi

echo ""

# 步骤 2: 准备插件目录
echo -e "${YELLOW}[Step 2/3] 准备插件目录...${NC}"
mkdir -p "$PLUGIN_DIR"

if [ -f "$PLUGIN_DIR/baseline" ]; then
    echo "  备份旧的插件文件..."
    mv "$PLUGIN_DIR/baseline" "$PLUGIN_DIR/baseline.bak.$(date +%s)" 2>/dev/null || true
fi

echo "  复制插件到: $PLUGIN_DIR"
cp "$BASELINE_DIR/baseline" "$PLUGIN_DIR/baseline"
chmod +x "$PLUGIN_DIR/baseline"

if [ -f "$PLUGIN_DIR/baseline" ]; then
    echo -e "${GREEN}  ✓ 插件已准备完成${NC}"
    echo "  插件路径: $PLUGIN_DIR/baseline"
    echo "  插件大小: $(du -h "$PLUGIN_DIR/baseline" | cut -f1)"
else
    echo -e "${RED}  ✗ 插件准备失败${NC}"
    exit 1
fi

echo ""

# 步骤 3: 运行测试程序
echo -e "${YELLOW}[Step 3/3] 运行测试程序...${NC}"
cd "$TEST_DIR"

if [ ! -f "main.go" ]; then
    echo -e "${RED}错误: 找不到测试程序 main.go${NC}"
    exit 1
fi

echo "  执行: go mod tidy"
go mod tidy > /dev/null 2>&1 || true

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}测试程序将自动执行以下操作：${NC}"
echo "  1. 启动 plugin daemon"
echo "  2. 加载 baseline 插件"
echo "  3. 发送测试任务"
echo "  4. 接收并打印结果"
echo ""
echo -e "${YELLOW}按 Ctrl+C 停止测试${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 运行测试程序
go run main.go

