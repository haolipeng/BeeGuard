# Agent 远程卸载测试流程

本文档描述通过 server HTTP API 远程卸载 Agent 的端到端测试流程：调用卸载 API → server 通过 gRPC 下发卸载命令（DataType 1061）→ Agent 生成卸载脚本并退出 → 脚本完成清理。

---

## 一、概述

### 卸载流程

```
HTTP Client              server Server              Agent                          卸载脚本
    │                         │                       │                               │
    │ POST /api/agent/uninstall                       │                               │
    │─────────────────────────→│                       │                               │
    │                         │  gRPC Command          │                               │
    │                         │  DataType=1061         │                               │
    │                         │───────────────────────→│                               │
    │  {"success":true}       │                       │  1. 生成 /tmp/cloudsec-uninstall-<PID>.sh
    │←─────────────────────────│                       │  2. systemd-run --scope 启动脚本
    │                         │                       │  3. agent.Cancel() 开始退出
    │                         │                       │──────────────────────────────→│
    │                         │                       │  (agent 退出中...)            │ 等待 agent PID 退出（60s）
    │                         │                       ×  agent 进程结束              │ 超时则 kill -9
    │                         │                       │                               │ systemctl disable
    │                         │                       │                               │ systemctl stop
    │                         │                       │                               │ dpkg --purge / rpm -e
    │                         │                       │                               │ rm -rf /opt/cloudsec/agent
    │                         │                       │                               │ rm -f 脚本自身
    │                         │                       │                               ×
```

### 关键技术点

| 项目 | 说明 |
|------|------|
| API 端点 | `POST /api/agent/uninstall` |
| gRPC 命令 | DataType 1061, ObjectName `cloudsec-agent` |
| 脚本启动方式 | `systemd-run --scope`（独立 cgroup，不受 `systemctl stop` 影响）；不可用时回退到直接执行 |
| 服务管理 | systemd service，`Restart=always`，`KillMode=control-group` |
| 包管理器 | 自动识别 dpkg（Debian/Ubuntu）或 rpm（CentOS/RHEL） |

### 前置条件

- Linux 操作系统（Ubuntu/CentOS）
- root 权限
- 本地 PostgreSQL 数据库可访问
- server Server 已部署（`/opt/cloudsec/server/`）
- Agent 已编译（`make build`）

---

## 二、环境准备

### 2.1 数据库准备

确保本地 PostgreSQL 可访问：

```bash
PGPASSWORD=root psql -h 127.0.0.1 -p 5432 -U postgres -d soc -c "SELECT 1;"
```

> 如果数据库不存在，参见 [local-integration-testing.md](local-integration-testing.md) 的 2.2 节创建数据库。

### 2.2 修改 server 数据库配置

确保 `/opt/cloudsec/server/conf/server.yaml` 的 `database` 部分指向本地：

```yaml
database:
  host: 127.0.0.1
  port: 5432
  user: postgres
  password: "root"
  database: soc
```

### 2.3 部署 Agent

编译并部署最新代码到 `/opt/cloudsec/agent/`：

```bash
cd /home/work/goProject/src/company/agent
make build && make deploy
```

### 2.4 修改 Agent 配置

确保 `/opt/cloudsec/agent/agent.yaml` 的 `server` 指向本地 server：

```yaml
server: "127.0.0.1:50051"
```

---

## 三、启动服务

### 3.1 启动 server Server

**Terminal A**：

```bash
cd /opt/cloudsec/server
sudo ./bin/server -config conf/server.yaml
```

启动成功判定：

```
INFO  gRPC Server 启动，监听端口 :50051
INFO  [HTTP] HTTP API Server 启动，监听端口 :8081
```

### 3.2 通过 install.sh 安装 Agent

**Terminal B**：

```bash
curl -fsSL http://:8081/install.sh | sudo bash
```

安装完成后，确认 agent.yaml 中 `server` 为 `127.0.0.1:50051`（install.sh 可能使用 server 配置中的远程地址，需手动修正）：

```bash
grep 'server' /opt/cloudsec/agent/agent.yaml
```

如果不是 `127.0.0.1:50051`，手动修改后重启 Agent：

```bash
sudo sed -i 's|server:.*|server: "127.0.0.1:50051"|' /opt/cloudsec/agent/agent.yaml
sudo systemctl restart cloudsec-agent
```

### 3.3 验证 Agent 在线

等待 5 秒后执行：

```bash
# 方式一：HTTP API
curl -s http://127.0.0.1:8081/api/agents | python3 -m json.tool
```

预期响应中 `agents` 数组包含已注册的 Agent。

```bash
# 方式二：查询数据库
PGPASSWORD=root psql -h 127.0.0.1 -U postgres -d soc -c \
  "SELECT agent_id, host_name, connection_status FROM agent_info WHERE connection_status = 1;"
```

**记录 `agent_id`**，后续卸载命令需要使用。`connection_status = 1` 表示 Agent 在线。

```bash
# 方式三：检查 systemd 服务状态
sudo systemctl status cloudsec-agent
```

预期输出 `Active: active (running)`，且 CGroup 中包含 agent 和各插件进程。

---

## 四、执行远程卸载

### 4.1 记录卸载前状态

卸载前先记录当前状态，用于卸载后对比验证：

```bash
echo "=== Agent 进程 ==="
pgrep -af 'cloudsec.*agent'

echo "=== systemd 服务状态 ==="
systemctl is-active cloudsec-agent
systemctl is-enabled cloudsec-agent

echo "=== 安装包 ==="
dpkg -l cloudsec-agent 2>/dev/null || rpm -q cloudsec-agent 2>/dev/null || echo "无安装包"

echo "=== 安装目录 ==="
ls /opt/cloudsec/agent/bin/agent 2>/dev/null && echo "存在" || echo "不存在"

echo "=== Agent PID ==="
pgrep -x agent
```

### 4.2 发送卸载命令

将 `<AGENT_ID>` 替换为 3.3 步骤中记录的 `agent_id`：

```bash
curl -X POST http://127.0.0.1:8081/api/agent/uninstall \
  -H 'Content-Type: application/json' \
  -d '{"agent_id":"<AGENT_ID>"}'
```

**预期响应**：

```json
{"success":true,"message":"Uninstall command sent to agent"}
```

**异常响应**：

| 响应 | 原因 |
|------|------|
| `{"success":false,"message":"Agent not found"}` | agent_id 错误或 Agent 不在线 |
| `{"success":false,"message":"Command queue full"}` | 命令队列满，稍后重试 |
| `{"success":false,"message":"Invalid request: ..."}` | 请求格式错误，检查 JSON 字段 |

### 4.3 观察卸载过程

卸载命令发送后，可观察以下过程：

**server 日志（Terminal A）**：

```
INFO  [Transfer] 发送命令: agent_id=xxx, data_type=1061, object_name=cloudsec-agent
```

**Agent 日志**（查看持久化日志，因为 Agent 即将退出）：

```bash
# 实时观察 Agent 日志
sudo tail -f /opt/cloudsec/agent/logs/agent/agent.log
```

预期日志顺序：

```
WARN  received uninstall command, will uninstall agent  {"data_type": 1061, "object_name": "cloudsec-agent"}
INFO  uninstall script created    {"path": "/tmp/cloudsec-uninstall-<PID>.sh"}
INFO  uninstall script started    {"script_pid": <SCRIPT_PID>}
INFO  uninstall script launched successfully, agent will exit
```

**验证脚本在独立 cgroup 中运行**（可选，在 Agent 退出前快速执行）：

```bash
# 查找卸载脚本的 PID
SCRIPT_PID=$(pgrep -f 'cloudsec-uninstall')

# 对比 cgroup：脚本应在 run-xxx.scope，而非 cloudsec-agent.service
cat /proc/$SCRIPT_PID/cgroup
# 预期: 0::/system.slice/run-xxxx.scope

AGENT_PID=$(pgrep -x agent)
cat /proc/$AGENT_PID/cgroup
# 预期: 0::/system.slice/cloudsec-agent.service
```

### 4.4 等待卸载完成

Agent 退出需要一定时间（各插件有序关闭），卸载脚本会等待 Agent 进程退出后继续执行。整个过程约需 60-90 秒。

```bash
# 等待卸载完成（轮询检查）
echo "等待卸载完成..."
for i in $(seq 1 90); do
    if ! pgrep -x agent > /dev/null 2>&1 && \
       ! systemctl is-active cloudsec-agent > /dev/null 2>&1; then
        echo "Agent 进程已退出，等待脚本完成清理..."
        sleep 10
        break
    fi
    sleep 1
done
echo "检查完成"
```

---

## 五、验证卸载结果

### 5.1 逐项检查

依次检查以下 6 项，**全部通过**才算卸载成功：

```bash
echo "========== 卸载结果验证 =========="

echo "--- 1. Agent 进程 ---"
pgrep -af 'cloudsec.*agent' && echo "[FAIL] Agent 进程仍在运行" || echo "[PASS] 无 Agent 进程"

echo "--- 2. systemd 服务状态 ---"
systemctl is-active cloudsec-agent 2>&1
# 预期: inactive 或 failed（非 active）

echo "--- 3. systemd 开机自启 ---"
systemctl is-enabled cloudsec-agent 2>&1
# 预期: disabled 或 "Failed to get unit file state"（service 文件已被包管理器删除）

echo "--- 4. 安装包 ---"
dpkg -l cloudsec-agent 2>&1 | grep -q '^ii' && echo "[FAIL] dpkg 包仍存在" || echo "[PASS] dpkg 包已卸载"
rpm -q cloudsec-agent 2>&1 | grep -q 'not installed' && echo "[PASS] rpm 包已卸载" || true

echo "--- 5. 安装目录 ---"
ls /opt/cloudsec/agent/ 2>/dev/null && echo "[FAIL] 安装目录仍存在" || echo "[PASS] 安装目录已清理"

echo "--- 6. 卸载脚本 ---"
ls /tmp/cloudsec-uninstall-*.sh 2>/dev/null && echo "[WARN] 存在历史卸载脚本（可能是之前的残留）" || echo "[PASS] 卸载脚本已自删除"
```

### 5.2 预期结果汇总

| 检查项 | 预期结果 | 说明 |
|--------|---------|------|
| Agent 进程 | 无 cloudsec-agent 相关进程 | 包括 agent 主进程和所有插件子进程 |
| systemd 服务 | `inactive` 或 `failed` | 非 `active` 状态 |
| systemd 开机自启 | `disabled` 或 service 文件不存在 | 不会在重启后自动启动 |
| dpkg/rpm 包 | 已卸载 | `dpkg -l` 查不到或 `rpm -q` 返回 not installed |
| 安装目录 | `/opt/cloudsec/agent/` 不存在 | 目录已被 `rm -rf` 清理 |
| 卸载脚本 | `/tmp/cloudsec-uninstall-<PID>.sh` 不存在 | 脚本最后一步自删除 |

### 5.3 数据库验证

卸载后 Agent 断开连接，server 会更新连接状态：

```sql
PGPASSWORD=root psql -h 127.0.0.1 -U postgres -d soc -c \
  "SELECT agent_id, connection_status, last_connected_at FROM agent_info WHERE agent_id = '<AGENT_ID>';"
```

预期 `connection_status = 0`（离线）。

---

## 六、常见问题

### 6.1 API 返回成功但 Agent 未卸载

**现象**：`curl` 返回 `{"success":true}`，但 Agent 进程仍在运行，安装目录仍存在。

**排查步骤**：

1. 检查 Agent 日志是否收到卸载命令：

   ```bash
   grep 'uninstall' /opt/cloudsec/agent/logs/agent/agent.log
   ```

   - 无 `received uninstall command` → Agent 与 server 的 gRPC 连接异常
   - 有 `failed to start uninstall script` → 脚本启动失败，检查 `/tmp` 是否可写

2. 检查卸载脚本是否存在：

   ```bash
   ls -la /tmp/cloudsec-uninstall-*.sh
   ```

   - 脚本存在但未执行完 → 检查脚本的 cgroup 是否与 agent 相同（见 4.3）

3. 检查 systemd-run 是否可用：

   ```bash
   which systemd-run
   systemd-run --scope echo "test"
   ```

   > 注意：即使 `systemd-run` 不可用，Agent 也会回退到直接执行脚本（新进程会话），卸载仍可正常完成。

### 6.2 Agent 卸载后又自动重启

**原因**：systemd service 配置了 `Restart=always`。

**排查**：卸载脚本应在 `systemctl stop` 之前执行 `systemctl disable`。如果仍然重启，说明脚本未完整执行。参见 6.1 排查。

### 6.3 dpkg 包残留

**现象**：Agent 进程已停止，但 `dpkg -l cloudsec-agent` 仍显示已安装。

**排查**：卸载脚本中 `dpkg --purge` 可能执行失败。手动清理：

```bash
sudo dpkg --purge cloudsec-agent 2>/dev/null || true
sudo rm -rf /opt/cloudsec/agent
```

### 6.4 卸载脚本残留在 /tmp

**现象**：`/tmp/cloudsec-uninstall-<PID>.sh` 文件仍存在。

**排查**：脚本在执行 `rm -f "$0"` 之前被中断。可安全删除：

```bash
rm -f /tmp/cloudsec-uninstall-*.sh
```

---

## 七、重新安装（用于重复测试）

卸载验证完成后，如需重复测试，执行以下步骤重新安装 Agent：

```bash
# 1. 确认清理干净
pgrep -af 'cloudsec.*agent' && sudo killall agent || echo "无残留进程"
sudo dpkg --purge cloudsec-agent 2>/dev/null || sudo rpm -e cloudsec-agent 2>/dev/null || true
sudo rm -rf /opt/cloudsec/agent
sudo rm -f /etc/systemd/system/cloudsec-agent.service
sudo systemctl daemon-reload

# 2. 重新安装
curl -fsSL http://127.0.0.1:8081/install.sh | sudo bash

# 3. 修正 server 地址（如需要）
sudo sed -i 's|server:.*|server: "127.0.0.1:50051"|' /opt/cloudsec/agent/agent.yaml
sudo systemctl restart cloudsec-agent

# 4. 验证 Agent 在线
sleep 5
curl -s http://127.0.0.1:8081/api/agents | python3 -m json.tool
```

确认 Agent 在线后，重复第四节步骤执行卸载测试。

---

## 八、测试后清理

```bash
# 停止 server（Terminal A 按 Ctrl+C）

# 清理残留的卸载脚本
rm -f /tmp/cloudsec-uninstall-*.sh

# 恢复 server 配置（如果修改过数据库配置）
cp /opt/cloudsec/server/conf/server.yaml.bak /opt/cloudsec/server/conf/server.yaml 2>/dev/null
```
