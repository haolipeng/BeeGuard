#!/bin/bash

# Collector 插件测试脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# 从 tests/e2e/collector 向上三级到 agent 根目录
AGENT_DIR="$(dirname "$(dirname "$(dirname "$SCRIPT_DIR")")")"
COLLECTOR_DIR="$AGENT_DIR/business_plugins/collector"
PLUGIN_DIR="/tmp/plugin/collector"
TEST_DIR="$SCRIPT_DIR"

echo -e "${GREEN}=== Collector Plugin Test Script ===${NC}"
echo ""

# 步骤 1: 准备 collector 插件
echo -e "${YELLOW}[Step 1/2] 准备 collector 插件...${NC}"
mkdir -p "$PLUGIN_DIR"

if [ ! -f "$COLLECTOR_DIR/collector" ]; then
    echo -e "${RED}错误: 找不到 collector 可执行文件${NC}"
    echo "请先编译 collector 插件:"
    echo "  cd $COLLECTOR_DIR && go build -o collector main.go process.go"
    exit 1
fi

echo "  复制插件到: $PLUGIN_DIR"
cp "$COLLECTOR_DIR/collector" "$PLUGIN_DIR/collector"
chmod +x "$PLUGIN_DIR/collector"

if [ -f "$PLUGIN_DIR/collector" ]; then
    echo -e "${GREEN}  ✓ Collector 插件已准备完成${NC}"
    echo "  插件路径: $PLUGIN_DIR/collector"
    echo "  插件大小: $(du -h "$PLUGIN_DIR/collector" | cut -f1)"
else
    echo -e "${RED}  ✗ 插件准备失败${NC}"
    exit 1
fi

echo ""

# 步骤 2: 运行测试程序
echo -e "${YELLOW}[Step 2/2] 运行测试程序...${NC}"
cd "$TEST_DIR"

if [ ! -f "test_collector" ]; then
    echo -e "${RED}错误: 找不到 test_collector 可执行文件${NC}"
    echo "请先编译测试程序:"
    echo "  cd $TEST_DIR && go build -o test_collector main.go"
    exit 1
fi

echo "  启动测试程序（将运行 30 秒）..."
echo "  按 Ctrl+C 可以提前退出"
echo ""
./test_collector

echo ""
echo -e "${GREEN}测试完成！${NC}"

