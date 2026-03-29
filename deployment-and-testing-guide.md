# 安全平台部署指南

本文档指导如何完整部署安全平台（Server + Agent + 插件），并进行功能测试。

---

## 一、系统概述

### 1.1 架构说明

```
┌─────────────────────────────────────────────────────────────────┐
│                         Server (server)                          │
│  ┌──────────────────┐  ┌──────────────────┐  ┌───────────────┐ │
│  │   gRPC Service   │  │   HTTP API       │  │  PostgreSQL   │ │
│  │   (port 50051)   │  │   (port 8080)    │  │   Database    │ │
│  └────────┬─────────┘  └──────────────────┘  └───────────────┘ │
└───────────│─────────────────────────────────────────────────────┘
            │ gRPC 双向流
┌───────────▼─────────────────────────────────────────────────────┐
│                         Agent                                    │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    Plugin Manager                         │   │
│  └──────────┬──────────────┬──────────────┬─────────────────┘   │
│             │              │              │                      │
│  ┌──────────▼───┐  ┌───────▼──────┐  ┌───▼───────────┐         │
│  │  Collector   │  │   Baseline   │  │   Detector    │         │
│  │  (资产采集)  │  │  (基线检查)  │  │  (暴力破解)   │         │
│  └──────────────┘  └──────────────┘  └───────────────┘         │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 组件说明

| 组件 | 说明 | 端口 |
|------|------|------|
| Server (server) | 服务端，接收数据、下发任务 | gRPC: 50051, HTTP: 8080 |
| Agent | 部署在目标主机的采集代理 | - |
| Collector | 资产采集插件（进程、端口、用户等） | - |
| Baseline | 基线安全检查插件 | - |
| Detector | 入侵检测插件（SSH/FTP暴力破解） | - |

---

## 二、环境准备

### 2.1 操作系统要求

- Linux (推荐 Ubuntu 20.04+ / CentOS 7+)
- Go 1.21+
- PostgreSQL 12+

### 2.2 数据库准备

```bash
# 1. 安装 PostgreSQL (Ubuntu)
sudo apt update && sudo apt install -y postgresql postgresql-contrib

# 2. 创建数据库
sudo -u postgres createdb soc

# 3. 设置密码
sudo -u postgres psql -c "ALTER USER postgres PASSWORD 'happy';"

# 4. 导入表结构
sudo -u postgres psql -d soc -f /home/work/goProject/src/BeeGuard/init_asset_db.sql

# 5. 验证
sudo -u postgres psql -d soc -c "\dt"
```

### 2.3 网络配置

确保以下端口可访问：
- **50051**: gRPC 服务端口（Agent 与 Server 通信）
- **8080**: HTTP API 端口（管理接口）
- **5432**: PostgreSQL 端口

---

## 三、Server (server) 部署

### 3.1 编译部署

```bash
cd /home/work/goProject/src/BeeGuard/server

# 编译
make build

# 部署到 /opt/cloudsec/server/
make deploy
```

### 3.2 配置文件

配置文件位置：`/opt/cloudsec/server/conf/server.yaml`

```yaml
server:
  port: 50051                   # gRPC 服务端口
  http_port: 8080               # HTTP API 端口
  max_recv_msg_size: 16         # 最大接收消息大小 (MB)
  max_send_msg_size: 16         # 最大发送消息大小 (MB)

database:
  host: localhost
  port: 5432
  user: postgres
  password: "happy"             # 修改为实际密码
  database: soc
  pool_size: 10

log:
  level: info                   # debug, info, warn, error
```

### 3.3 启动服务

```bash
# 前台运行（调试）
/opt/cloudsec/server/bin/server -config /opt/cloudsec/server/conf/server.yaml

# 后台运行（生产）
nohup /opt/cloudsec/server/bin/server -config /opt/cloudsec/server/conf/server.yaml \
    > /opt/cloudsec/server/logs/server/server.log 2>&1 &
```

### 3.4 验证服务

```bash
# 检查端口
netstat -tlnp | grep -E "50051|8080"

# 检查进程
ps aux | grep server

# 测试 HTTP API
curl http://localhost:8080/api/agents
```

---

## 四、Agent 部署

### 4.1 编译部署

```bash
cd /home/work/goProject/src/BeeGuard/agent

# 编译 Agent + 所有插件
make build

# 部署到 /opt/cloudsec/agent/
make deploy
```

### 4.2 配置文件

配置文件位置：`/opt/cloudsec/agent/agent.yaml`

```yaml
# Server 连接地址
server: "127.0.0.1:50051"

# 连接超时时间（秒）
connect_timeout: 30

# Agent 工作目录
working_directory: "/opt/cloudsec/agent/data/agent"

# 插件目录
plugins_directory: "/opt/cloudsec/agent/plugins"

# 连接失败最大重试次数
retry_max_count: 10

# 重试间隔（秒）
retry_interval: 5
```

### 4.3 启动 Agent

```bash
# 前台运行（调试，需要 root 权限）
sudo /opt/cloudsec/agent/bin/agent -config /opt/cloudsec/agent/agent.yaml

# 后台运行（生产）
sudo nohup /opt/cloudsec/agent/bin/agent -config /opt/cloudsec/agent/agent.yaml \
    > /opt/cloudsec/agent/logs/agent/agent.log 2>&1 &
```

### 4.5 验证连接

```bash
# 检查 Agent 日志
tail -f /opt/cloudsec/agent/logs/agent/agent.log

# 通过 API 查看已连接的 Agent
curl http://localhost:8080/api/agents
```

预期输出：
```json
{
  "agents": [
    {
      "agent_id": "xxx-xxx-xxx",
      "hostname": "your-host",
      "ipv4": ["192.168.1.100"],
      "version": "1.0.0"
    }
  ],
  "total": 1
}
```

---

## 五、插件测试

### 5.1 Collector 插件测试

通过 HTTP API 下发采集任务，验证数据采集功能。

#### 获取 Agent ID

```bash
# 先获取 Agent ID
AGENT_ID=$(curl -s http://localhost:8080/api/agents | jq -r '.agents[0].agent_id')
echo "Agent ID: $AGENT_ID"
```

#### 进程采集 (DataType: 5050)

```bash
curl -X POST http://localhost:8080/api/task \
  -H "Content-Type: application/json" \
  -d "{
    \"agent_id\": \"$AGENT_ID\",
    \"object_name\": \"collector\",
    \"data_type\": 5050,
    \"data\": \"{}\",
    \"token\": \"task-process-001\"
  }"
```

#### 端口采集 (DataType: 5051)

```bash
curl -X POST http://localhost:8080/api/task \
  -H "Content-Type: application/json" \
  -d "{
    \"agent_id\": \"$AGENT_ID\",
    \"object_name\": \"collector\",
    \"data_type\": 5051,
    \"data\": \"{}\",
    \"token\": \"task-port-001\"
  }"
```

#### 用户账号采集 (DataType: 5052)

```bash
curl -X POST http://localhost:8080/api/task \
  -H "Content-Type: application/json" \
  -d "{
    \"agent_id\": \"$AGENT_ID\",
    \"object_name\": \"collector\",
    \"data_type\": 5052,
    \"data\": \"{}\",
    \"token\": \"task-user-001\"
  }"
```

#### 其他采集任务

| DataType | 任务类型 |
|----------|---------|
| 5054 | 系统服务采集 |
| 5055 | 软件包采集 |
| 5056 | 容器采集 |
| 5057 | 可疑环境变量检测 |
| 5062 | 内核模块采集 |

#### 验证采集结果

```bash
# 查看 Server 日志
tail -f /opt/cloudsec/server/logs/server/server.log

# 查询数据库
sudo -u postgres psql -d soc -c "SELECT count(*) FROM asset_process;"
sudo -u postgres psql -d soc -c "SELECT count(*) FROM asset_port;"
sudo -u postgres psql -d soc -c "SELECT count(*) FROM asset_account;"
```

---

### 5.2 Detector 插件测试（暴力破解检测）

Detector 插件监控 SSH/FTP 登录日志，检测暴力破解行为。

#### 检测规则说明

**SSH 检测规则** (`/opt/cloudsec/agent/plugins/detector/config/rules/ssh.yaml`):
- 同一 IP 在 120 秒内认证失败 6 次触发告警
- 告警后 60 秒内不重复告警

**FTP 检测规则** (`/opt/cloudsec/agent/plugins/detector/config/rules/ftp.yaml`):
- 同一 IP 在 120 秒内登录失败 6 次触发告警
- 同一 IP 在 60 秒内连接 10 次触发告警

#### 模拟 SSH 暴力破解测试

```bash
# 方法1: 使用错误密码尝试 SSH 登录
for i in {1..10}; do
  sshpass -p 'wrongpassword' ssh -o StrictHostKeyChecking=no testuser@localhost 2>/dev/null
  sleep 1
done

# 方法2: 直接写入测试日志（模拟）
for i in {1..10}; do
  echo "$(date '+%b %d %H:%M:%S') localhost sshd[12345]: Failed password for root from 192.168.1.100 port 22 ssh2" \
    | sudo tee -a /var/log/auth.log
  sleep 1
done
```

#### 查看告警结果

```bash
# 查看 Detector 日志
tail -f /opt/cloudsec/agent/logs/plugins/detector/detector.log

# 查询告警表
sudo -u postgres psql -d soc -c "SELECT * FROM alert_brute_force ORDER BY created_at DESC LIMIT 10;"
```

#### 下发检测配置

通过 API 动态更新检测规则：

```bash
curl -X POST http://localhost:8080/api/detector/config \
  -H "Content-Type: application/json" \
  -d "{
    \"agent_id\": \"$AGENT_ID\",
    \"service\": \"ssh\",
    \"config\": {
      \"enabled\": true,
      \"rules\": [
        {
          \"name\": \"auth_failure_brute_force\",
          \"description\": \"SSH认证失败暴力破解检测\",
          \"action\": \"failed\",
          \"frequency\": 5,
          \"timeframe\": 60,
          \"level\": 10,
          \"ignore\": 30,
          \"group_by\": \"source_ip\"
        }
      ],
      \"whitelist\": [\"127.0.0.1\", \"192.168.1.1\"]
    }
  }"
```

---

### 5.3 Detector 插件测试（高危命令检测）

Detector 插件通过 Linux Audit 子系统监控命令执行，检测高危命令和反弹Shell行为。

#### 检测规则说明

高危命令检测规则位于 `/opt/cloudsec/agent/plugins/detector/config/rules/command.yaml`：

**检测类别：**
- **反弹Shell检测**：检测 bash/nc/python/perl 等反弹Shell命令
- **权限提升检测**：检测 sudo、chmod 777、setuid 等危险操作
- **文件删除检测**：检测 rm -rf 等危险删除命令
- **日志篡改检测**：检测清除日志的命令
- **网络扫描检测**：检测 nmap 等扫描工具
- **服务停止检测**：检测停止安全服务的命令

#### 启动高危命令检测

高危命令检测需要 root 权限（访问 audit 子系统）：

```bash
# 确保 Agent 以 root 运行
sudo /opt/cloudsec/agent/bin/agent -config /opt/cloudsec/agent/agent.yaml
```

#### 模拟高危命令测试

```bash
# 测试1: 危险 chmod 操作
chmod 777 /tmp/testfile 2>/dev/null

# 测试2: 文件删除（安全测试，不会真删除）
touch /tmp/test_rm_target && rm -rf /tmp/test_rm_target

# 测试3: 模拟反弹Shell命令（不会真正执行，仅触发检测）
echo "test" > /dev/tcp/127.0.0.1/9999 2>/dev/null || true

# 测试4: nmap 扫描（如果安装了 nmap）
which nmap && nmap -sP 127.0.0.1 2>/dev/null || true
```

#### 查看高危命令告警

```bash
# 查看 Detector 日志
tail -f /opt/cloudsec/agent/logs/plugins/detector/detector.log

# 查询高危命令告警表
PGPASSWORD="happy" psql -h localhost -U postgres -d soc \
  -c "SELECT id, command, command_type, \"user\", alert_time FROM alert_dangerous_command ORDER BY created_at DESC LIMIT 10;"

# 查询反弹Shell告警表
PGPASSWORD="happy" psql -h localhost -U postgres -d soc \
  -c "SELECT id, command_line, shell_type, target_host, target_port FROM alert_reverse_shell ORDER BY created_at DESC LIMIT 10;"
```

#### DataType 说明

| DataType | 说明 |
|----------|------|
| 6003 | 高危命令告警 |
| 6004 | 反弹Shell告警 |

---

### 5.4 Baseline 插件测试

（基线检查插件测试方法待补充）

---

## 六、API 接口测试

### 6.1 查询 Agent 列表

```bash
curl http://localhost:8080/api/agents
```

### 6.2 查询指定 Agent

```bash
curl http://localhost:8080/api/agents/$AGENT_ID
```

### 6.3 下发采集任务

```bash
curl -X POST http://localhost:8080/api/task \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "your-agent-id",
    "object_name": "collector",
    "data_type": 5050,
    "data": "{}",
    "token": "task-001"
  }'
```

### 6.4 下发检测器配置

```bash
curl -X POST http://localhost:8080/api/detector/config \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "your-agent-id",
    "service": "ssh",
    "config": {
      "enabled": true,
      "rules": [...],
      "whitelist": ["127.0.0.1"]
    }
  }'
```

---

## 七、常见问题排查

### 7.1 Agent 无法连接 Server

**检查项：**
1. Server 是否已启动：`ps aux | grep server`
2. 端口是否监听：`netstat -tlnp | grep 50051`
3. 防火墙是否放行：`iptables -L -n`
4. Agent 配置的 server 地址是否正确

### 7.2 数据库连接失败

**检查项：**
1. PostgreSQL 是否运行：`systemctl status postgresql`
2. 数据库是否存在：`sudo -u postgres psql -l`
3. 用户权限是否正确
4. server.yaml 中密码是否正确

### 7.3 插件加载失败

**检查项：**
1. 插件文件是否存在：`ls -la /opt/cloudsec/agent/plugins/`
2. 插件是否有执行权限：`chmod 755 /opt/cloudsec/agent/plugins/*/`
3. Agent 日志错误信息：`tail -f /opt/cloudsec/agent/logs/agent/agent.log`

### 7.4 采集任务无响应

**检查项：**
1. 确认 Agent 已连接：`curl http://localhost:8080/api/agents`
2. 检查任务下发响应是否成功
3. 检查 Agent 和插件日志

---

## 八、清理与维护

### 8.1 停止服务

```bash
# 停止 Agent
sudo pkill -f "/opt/cloudsec/agent/bin/agent"

# 停止 Server
pkill -f "/opt/cloudsec/server/bin/server"
```

### 8.2 清理日志

```bash
# 清理所有日志
sudo rm -rf /opt/cloudsec/agent/logs/agent/*
sudo rm -rf /opt/cloudsec/agent/logs/plugins/*
sudo rm -rf /opt/cloudsec/server/logs/server/*
```

### 8.3 清理运行时数据

```bash
sudo rm -rf /opt/cloudsec/agent/data/*
```

### 8.4 清理编译产物

```bash
cd /home/work/goProject/src/BeeGuard/agent && make clean
cd /home/work/goProject/src/BeeGuard/server && make clean
```

### 8.5 完全卸载

```bash
# 停止所有服务
sudo pkill -f "/opt/cloudsec/"

# 删除部署目录
sudo rm -rf /opt/cloudsec

# 删除数据库（可选）
sudo -u postgres dropdb soc
```

---

## 附录：DataType 参考

| DataType | 类型 | 说明 |
|----------|------|------|
| 1060 | 命令 | 关闭 Agent |
| 5050 | 采集 | 进程采集 |
| 5051 | 采集 | 端口采集 |
| 5052 | 采集 | 用户账号采集 |
| 5054 | 采集 | 系统服务采集 |
| 5055 | 采集 | 软件包采集 |
| 5056 | 采集 | 容器采集 |
| 5057 | 采集 | 可疑环境变量检测 |
| 5062 | 采集 | 内核模块采集 |
| 5100 | 响应 | 任务执行结果 |
| 6001 | 告警 | SSH 暴力破解告警 |
| 6002 | 告警 | FTP 暴力破解告警 |
| 6003 | 告警 | 高危命令告警 |
| 6004 | 告警 | 反弹Shell告警 |
| 6010 | 配置 | 检测器配置更新 |
