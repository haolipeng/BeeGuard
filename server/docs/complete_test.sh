#!/bin/bash

# 完整API功能测试脚本
BASE_URL="http://localhost:8081"
echo "🚀 开始完整API功能测试..."

# 测试结果收集数组
declare -A TEST_RESULTS

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试函数
test_api() {
    local module_name=$1
    local url=$2
    local description=$3
    
    echo "📋 测试模块: $module_name"
    echo "请求URL: $url"
    
    local response=$(curl -s -X GET "$url" -H "Content-Type: application/json")
    
    if [[ $response == *"data"* ]] || [[ $response == *"pagination"* ]]; then
        echo -e "${GREEN}✅ $description: 成功${NC}"
        TEST_RESULTS["$module_name"]="✅ 成功"
        echo "响应预览: $(echo $response | cut -c1-100)..."
    else
        echo -e "${RED}❌ $description: 失败${NC}"
        TEST_RESULTS["$module_name"]="❌ 失败"
        echo "错误响应: $response"
    fi
    echo ""
}

# 1. 入侵检测规则管理
test_api "入侵检测规则" "${BASE_URL}/api1/hids_rules/rules?page=1&limit=5" "获取入侵检测规则列表"

# 2. 漏洞检测规则管理  
test_api "漏洞检测规则" "${BASE_URL}/api1/vuln_rules/rules?page=1&limit=5" "获取漏洞检测规则列表"

# 3. 基线模板管理
test_api "基线模板" "${BASE_URL}/api1/baseline/templates?page=1&limit=5" "获取基线模板列表"

# 4. 基线模板与主机关联
test_api "基线关联" "${BASE_URL}/api1/baseline/links?page=1&limit=5" "获取基线模板关联列表"

# 5. 基线检查项管理
test_api "基线检查项" "${BASE_URL}/api1/baseline/items?page=1&limit=5" "获取基线检查项列表"

# 6. 基线检查结果明细
test_api "检查结果明细" "${BASE_URL}/api1/baseline/details?page=1&limit=5" "获取基线检查结果明细列表"

# 7. 基线检查主机统计
test_api "主机统计" "${BASE_URL}/api1/baseline/host_views?page=1&limit=5" "获取基线检查主机统计列表"

# 8. 基线检查项统计
test_api "检查项统计" "${BASE_URL}/api1/baseline/item_views?page=1&limit=5" "获取基线检查项统计列表"

# 9. Agent客户端管理
test_api "Agent管理" "${BASE_URL}/api1/system/agents?page=1&limit=5" "获取Agent客户端列表"

# 10. 容器漏洞扫描结果
test_api "容器漏洞" "${BASE_URL}/api1/vulns/image_details?page=1&limit=5" "获取容器漏洞扫描结果列表"

# 输出测试结果摘要
echo ""
echo "📊 完整测试结果汇总:"
echo "========================"
success_count=0
total_count=0

for module in "${!TEST_RESULTS[@]}"; do
    result="${TEST_RESULTS[$module]}"
    echo "$module: $result"
    ((total_count++))
    if [[ $result == *"✅"* ]]; then
        ((success_count++))
    fi
done

echo ""
echo "📈 测试统计:"
echo "总测试模块: $total_count"
echo "成功模块: $success_count"
echo "失败模块: $((total_count - success_count))"
echo "成功率: $((success_count * 100 / total_count))%"

if [ $success_count -eq $total_count ]; then
    echo -e "${GREEN}🎉 所有API测试通过!${NC}"
else
    echo -e "${YELLOW}⚠️  部分API测试失败，请检查相关模块${NC}"
fi

echo ""
echo "🏁 完整API测试完成!"