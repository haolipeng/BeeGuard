#!/bin/bash

# 测试数据填充脚本
BASE_URL="http://localhost:8081"
echo "🌱 开始填充测试数据..."

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 数据填充结果统计
declare -A DATA_FILL_RESULTS

# 添加数据函数
add_data() {
    local module_name=$1
    local url=$2
    local data=$3
    local description=$4
    
    echo -e "${BLUE}📥 添加$module_name数据...${NC}"
    echo "请求URL: $url"
    
    local response=$(curl -s -X POST "$url" \
        -H "Content-Type: application/json" \
        -d "$data")
    
    if [[ $response == *"message"* ]] && [[ $response != *"error"* ]]; then
        echo -e "${GREEN}✅ $description: 成功${NC}"
        DATA_FILL_RESULTS["$module_name"]="✅ 成功"
        echo "响应: $response"
    else
        echo -e "${RED}❌ $description: 失败${NC}"
        DATA_FILL_RESULTS["$module_name"]="❌ 失败"
        echo "错误响应: $response"
    fi
    echo ""
}

# 1. 添加基线模板数据
echo "=== 1. 基线模板数据 ==="
add_data "基线模板" "${BASE_URL}/api1/baseline/templates/create" '{
    "template_name": "Linux安全基线检查模板",
    "os_type": "linux",
    "description": "用于Linux服务器的安全基线检查",
    "is_enabled": 1,
    "created_by": "admin"
}' "创建Linux基线模板"

add_data "基线模板" "${BASE_URL}/api1/baseline/templates/create" '{
    "template_name": "Windows安全基线检查模板", 
    "os_type": "windows",
    "description": "用于Windows服务器的安全基线检查",
    "is_enabled": 1,
    "created_by": "admin"
}' "创建Windows基线模板"

# 2. 添加基线检查项数据（需要先获取基线模板ID）
echo "=== 2. 基线检查项数据 ==="
TEMPLATE_ID=1  # 假设第一个模板ID为1

add_data "基线检查项" "${BASE_URL}/api1/baseline/items/create" '{
    "baseline_id": '$TEMPLATE_ID',
    "item_name": "SSH服务配置检查",
    "category": "系统服务",
    "risk_level": "high",
    "check_type": "config",
    "check_script": "cat /etc/ssh/sshd_config | grep -E \"^PermitRootLogin|^PasswordAuthentication\"",
    "expected_value": "PermitRootLogin no\\nPasswordAuthentication no",
    "description": "检查SSH服务安全配置",
    "is_enabled": 1
}' "创建SSH配置检查项"

add_data "基线检查项" "${BASE_URL}/api1/baseline/items/create" '{
    "baseline_id": '$TEMPLATE_ID',
    "item_name": "防火墙状态检查",
    "category": "网络安全",
    "risk_level": "medium",
    "check_type": "command",
    "check_script": "systemctl is-active firewalld",
    "expected_value": "active",
    "description": "检查防火墙服务运行状态",
    "is_enabled": 1
}' "创建防火墙检查项"

add_data "基线检查项" "${BASE_URL}/api1/baseline/items/create" '{
    "baseline_id": '$TEMPLATE_ID',
    "item_name": "用户密码策略检查",
    "category": "账户安全",
    "risk_level": "high",
    "check_type": "config",
    "check_script": "cat /etc/login.defs | grep PASS_MAX_DAYS",
    "expected_value": "PASS_MAX_DAYS   90",
    "description": "检查用户密码最大使用期限",
    "is_enabled": 1
}' "创建密码策略检查项"

# 3. 添加基线模板与主机关联数据
echo "=== 3. 基线关联数据 ==="
add_data "基线关联" "${BASE_URL}/api1/baseline/links/create" '{
    "baseline_template_id": '$TEMPLATE_ID',
    "target_range": "[1, 2, 3]",
    "scan_frequency": "daily"
}' "创建基线模板关联"

# 4. 添加入侵检测规则数据
echo "=== 4. 入侵检测规则数据 ==="
add_data "入侵检测规则" "${BASE_URL}/api1/hids_rules/rules" '{
    "rule_name": "Web攻击检测规则",
    "rule_feature": "GET.*\\.php\\?|POST.*sqlmap|union.*select",
    "rule_level": "高",
    "threat_type": "Web攻击",
    "trigger_action": "记录日志并告警",
    "rule_status": "生效中",
    "rule_description": "检测常见的Web应用攻击模式"
}' "创建Web攻击检测规则"

add_data "入侵检测规则" "${BASE_URL}/api1/hids_rules/rules" '{
    "rule_name": "文件篡改检测规则",
    "rule_feature": "chmod.*777|chown.*root.*tmp|mv.*\\/etc",
    "rule_level": "紧急",
    "threat_type": "文件操作",
    "trigger_action": "阻断并告警",
    "rule_status": "生效中", 
    "rule_description": "检测敏感文件的异常修改操作"
}' "创建文件篡改检测规则"

# 5. 添加漏洞检测规则数据
echo "=== 5. 漏洞检测规则数据 ==="
add_data "漏洞检测规则" "${BASE_URL}/api1/vuln_rules/rules" '{
    "cve_id": "CVE-2023-9999",
    "vuln_name": "测试漏洞-缓冲区溢出",
    "severity": "high",
    "cvss_score": 7.5,
    "description": "测试用的缓冲区溢出漏洞",
    "fix_suggestion": "升级到最新版本",
    "status": 0
}' "创建缓冲区溢出测试漏洞"

add_data "漏洞检测规则" "${BASE_URL}/api1/vuln_rules/rules" '{
    "cve_id": "CVE-2023-8888",
    "vuln_name": "测试漏洞-权限提升",
    "severity": "critical", 
    "cvss_score": 9.0,
    "description": "测试用的权限提升漏洞",
    "fix_suggestion": "应用安全补丁",
    "status": 0
}' "创建权限提升测试漏洞"

# 显示数据填充结果
echo ""
echo "📊 数据填充结果汇总:"
echo "======================"
success_count=0
total_count=0

for module in "${!DATA_FILL_RESULTS[@]}"; do
    result="${DATA_FILL_RESULTS[$module]}"
    echo "$module: $result"
    ((total_count++))
    if [[ $result == *"✅"* ]]; then
        ((success_count++))
    fi
done

echo ""
echo "📈 数据填充统计:"
echo "总填充模块: $total_count"
echo "成功模块: $success_count" 
echo "失败模块: $((total_count - success_count))"
echo "成功率: $((success_count * 100 / total_count))%"

if [ $success_count -eq $total_count ]; then
    echo -e "${GREEN}🎉 所有数据填充成功!${NC}"
else
    echo -e "${YELLOW}⚠️  部分数据填充失败，请检查相关模块${NC}"
fi

echo ""
echo "🔄 现在重新测试所有API列表接口..."

# 重新测试所有列表接口
echo "=== 重新测试API列表接口 ==="
test_api() {
    local module_name=$1
    local url=$2
    local description=$3
    
    echo "📋 测试模块: $module_name"
    local response=$(curl -s -X GET "$url" -H "Content-Type: application/json")
    
    if [[ $response == *"data"* ]] && [[ $response != *"[]"* ]]; then
        echo -e "${GREEN}✅ $description: 有数据返回${NC}"
        echo "数据预览: $(echo $response | jq '.data | length' 2>/dev/null || echo '无法解析')"
    elif [[ $response == *"[]"* ]]; then
        echo -e "${YELLOW}⚠️  $description: 返回空列表${NC}"
    else
        echo -e "${RED}❌ $description: 请求失败${NC}"
    fi
    echo ""
}

# 测试所有模块的列表接口
test_api "基线模板" "${BASE_URL}/api1/baseline/templates?page=1&limit=5" "获取基线模板列表"
test_api "基线检查项" "${BASE_URL}/api1/baseline/items?page=1&limit=5" "获取基线检查项列表" 
test_api "基线关联" "${BASE_URL}/api1/baseline/links?page=1&limit=5" "获取基线关联列表"
test_api "入侵检测规则" "${BASE_URL}/api1/hids_rules/rules?page=1&limit=5" "获取入侵检测规则列表"
test_api "漏洞检测规则" "${BASE_URL}/api1/vuln_rules/rules?page=1&limit=5" "获取漏洞检测规则列表"

echo "🏁 数据填充和验证完成!"