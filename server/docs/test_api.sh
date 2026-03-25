#!/bin/bash

# API测试脚本
BASE_URL="http://localhost:8081"
echo "🚀 开始API功能测试..."

# 测试结果收集
TEST_RESULTS=()

# 1. 测试Agent客户端管理
echo "📋 测试1: Agent客户端管理"
echo "创建Agent..."
AGENT_CREATE_RESPONSE=$(curl -s -X POST "${BASE_URL}/api1/system/agents/create" \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "TEST-AGENT-001",
    "agent_version": "1.0.0",
    "connection_status": 1,
    "host_name": "test-server-01",
    "host_ip": "192.168.1.100",
    "os_type": "linux"
  }')

echo "Agent创建响应: $AGENT_CREATE_RESPONSE"
TEST_RESULTS+=("Agent创建: $(echo $AGENT_CREATE_RESPONSE | grep -q 'message.*成功' && echo '✅ 成功' || echo '❌ 失败')")

echo "获取Agent列表..."
AGENT_LIST_RESPONSE=$(curl -s -X GET "${BASE_URL}/api1/system/agents?page=1&limit=5")
echo "Agent列表响应: $AGENT_LIST_RESPONSE"
TEST_RESULTS+=("Agent列表: $(echo $AGENT_LIST_RESPONSE | grep -q 'data' && echo '✅ 成功' || echo '❌ 失败')")

# 2. 测试漏洞检测规则管理
echo "📋 测试2: 漏洞检测规则管理"
echo "获取漏洞规则列表..."
VULN_LIST_RESPONSE=$(curl -s -X GET "${BASE_URL}/api1/vuln_rules/rules?page=1&limit=5")
echo "漏洞规则列表响应: $VULN_LIST_RESPONSE"
TEST_RESULTS+=("漏洞规则列表: $(echo $VULN_LIST_RESPONSE | grep -q 'data' && echo '✅ 成功' || echo '❌ 失败')")

# 3. 测试基线模板管理
echo "📋 测试3: 基线模板管理"
echo "获取基线模板列表..."
BASELINE_LIST_RESPONSE=$(curl -s -X GET "${BASE_URL}/api1/baseline/templates?page=1&limit=5")
echo "基线模板列表响应: $BASELINE_LIST_RESPONSE"
TEST_RESULTS+=("基线模板列表: $(echo $BASELINE_LIST_RESPONSE | grep -q 'data' && echo '✅ 成功' || echo '❌ 失败')")

# 4. 测试入侵检测规则管理
echo "📋 测试4: 入侵检测规则管理"
echo "获取入侵检测规则列表..."
HIDS_LIST_RESPONSE=$(curl -s -X GET "${BASE_URL}/api1/hids_rules/rules?page=1&limit=5")
echo "入侵检测规则列表响应: $HIDS_LIST_RESPONSE"
TEST_RESULTS+=("入侵检测规则列表: $(echo $HIDS_LIST_RESPONSE | grep -q 'data' && echo '✅ 成功' || echo '❌ 失败')")

# 输出测试结果摘要
echo ""
echo "📊 测试结果汇总:"
echo "=================="
for result in "${TEST_RESULTS[@]}"; do
    echo "$result"
done

echo ""
echo "🏁 API测试完成!"