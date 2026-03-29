# Collector 插件端到端测试指南

本文档描述如何手动测试 Collector 插件的完整流程，包括所有资产采集功能。

## 测试架构

```
┌─────────────────┐     gRPC      ┌─────────────────┐     Pipe      ┌─────────────────┐
│     Server      │◄─────────────►│     Agent       │◄─────────────►│ Collector Plugin│
│   (server)       │   50051端口    │                 │               │                 │
└────────┬────────┘               └─────────────────┘               └─────────────────┘
         │
         ▼
┌─────────────────┐
│   PostgreSQL    │
│   (soc 数据库)   │
└─────────────────┘
```

**数据流程：**
1. Server 启动，监听 gRPC (50051) 和 HTTP (8080) 端口
2. Agent 连接 Server
3. 通过 HTTP API 下发插件配置，触发 Agent 加载 Collector 插件
4. 通过 HTTP API 下发采集任务（指定 DataType）
5. Collector 插件执行采集并将数据返回给 Agent
6. Agent 通过 gRPC 将数据上报给 Server
7. Server 解析数据并写入 PostgreSQL 数据库

---

## 数据类型说明

| DataType | 名称 | 说明 | 对应数据库表 |
|----------|------|------|-------------|
| 5050 | Process | 进程采集 | asset_process |
| 5051 | Port | 端口采集 | asset_port |
| 5052 | User | 用户账号采集 | asset_account |
| 5054 | Service | 系统服务采集 | asset_system_service |
| 5055 | Software | 软件包采集 | asset_software |
| 5056 | Container | 容器采集 | asset_container |
| 5057 | EnvSuspicious | 可疑环境变量检测 | asset_env_suspicious |
| 5062 | Kmod | 内核模块采集 | asset_kmod |
| 5100 | TaskResult | 任务执行结果 | - |

> **注意：** `asset_host` 表会在收到任何数据包时自动从包头信息更新，无需单独触发。

---

## 前置条件

### 1. 编译部署

```bash
# 编译并部署 Server
cd /home/work/goProject/src/BeeGuard/server
make build && make deploy

# 编译并部署 Agent 和插件
cd /home/work/goProject/src/BeeGuard/agent
make build && make deploy
```

### 2. 配置文件

#### Server 配置

位置：`/opt/cloudsec/server/conf/server.yaml`

```yaml
server:
  port: 50051
  http_port: 8080
  max_recv_msg_size: 16
  max_send_msg_size: 16

database:
  host: localhost
  port: 5432
  user: postgres
  password: "happy"
  database: soc
  pool_size: 10

log:
  level: info
```

#### Agent 配置

位置：`/opt/cloudsec/agent/agent.yaml`

```yaml
server: "127.0.0.1:50051"
connect_timeout: 30
working_directory: "/opt/cloudsec/agent/data/agent"
plugins_directory: "/opt/cloudsec/agent/plugins"
retry_max_count: 10
retry_interval: 5
```

### 3. 数据库准备

```bash
# 初始化数据库（如需重建）
cd /home/work/goProject/src/BeeGuard/server
./rebuild_asset_db.sh
```

---

## 测试步骤

### 步骤一：启动 Server

**终端 1：**

```bash
# 前台运行（便于查看日志）
/opt/cloudsec/server/bin/server -config /opt/cloudsec/server/conf/server.yaml

# 或后台运行
nohup /opt/cloudsec/server/bin/server -config /opt/cloudsec/server/conf/server.yaml >> /opt/cloudsec/server/logs/server/server.log 2>&1 &
```

**预期输出：**
```
2026/01/27 13:21:31 配置加载成功: grpc_port=50051, http_port=8080, log_level=info
2026/01/27 13:21:31 [DB] PostgreSQL 连接成功: localhost:5432/soc
2026/01/27 13:21:31 gRPC Server 启动，监听端口 :50051
2026/01/27 13:21:31 [API] HTTP server starting on :8080
```

---

### 步骤二：启动 Agent

**终端 2：**

```bash
# 前台运行
sudo /opt/cloudsec/agent/bin/agent -config /opt/cloudsec/agent/agent.yaml

# 或后台运行
sudo nohup /opt/cloudsec/agent/bin/agent -config /opt/cloudsec/agent/agent.yaml >> /tmp/agent.log 2>&1 &
```

**预期输出：**
```
agent start running!
2026/01/27 13:24:42 INFO config initialized successfully
2026-01-27T13:24:42.707+0800    INFO    agent/main.go:60    ++++++++++++++++++++++++++++++running++++++++++++++++++++++++++++++
2026-01-27T13:24:42.707+0800    INFO    transport/connection.go:145    connected to server    {"server": "127.0.0.1:50051"}
```

**Server 端会显示：**
```
2026/01/27 13:24:42 [Transfer] Agent 连接 agent_id=6bb9735d-66ee-556a-8981-62d127daf308 hostname=ubuntu version=8383128-dirty
```

---

### 步骤三：下发插件配置

在 Agent 加载插件之前，需要先下发插件配置：

```bash
# 获取 Agent ID
AGENT_ID=$(curl --noproxy '*' -s http://127.0.0.1:8080/api/agents | jq -r '.[0].agent_id')
echo "Agent ID: $AGENT_ID"

# 下发插件配置
curl --noproxy '*' -s -X POST http://127.0.0.1:8080/api/config \
  -H "Content-Type: application/json" \
  -d "{\"agent_id\": \"$AGENT_ID\", \"plugins\": [{\"name\": \"collector\", \"version\": \"1.0.0\", \"type\": \"binary\"}]}"
```

**预期输出：**
```json
{"success":true,"message":"Plugin config sent to agent"}
```

**Agent 日志会显示：**
```
2026-01-27T13:32:29.378+0800    INFO    transport/transfer.go:214    received config command    {"plugin_count": 1, "plugins": ["collector"]}
2026-01-27T13:32:29.378+0800    INFO    plugin/plugin_linux.go:124    plugin's process will start    {"plugin": "collector"}
2026-01-27T13:32:29.379+0800    INFO    plugin/plugin.go:235    plugin has been loaded    {"plugin": "collector"}
```

---

### 步骤四：下发采集任务

#### 方式一：使用脚本（推荐）

```bash
cd /home/work/goProject/src/BeeGuard/server/test_data

# 查看帮助
./send_task.sh help

# 列出所有可用任务
./send_task.sh list

# 发送单个任务
./send_task.sh process      # 进程采集
./send_task.sh port         # 端口采集
./send_task.sh user         # 用户账号采集

# 发送所有任务
./send_task.sh all
```

#### 方式二：使用 curl + JSON 文件

```bash
cd /home/work/goProject/src/BeeGuard/server/test_data

# 发送进程采集任务
curl --noproxy '*' -s -X POST http://127.0.0.1:8080/api/task \
  -H "Content-Type: application/json" \
  -d @task_process.json

# 发送端口采集任务
curl --noproxy '*' -s -X POST http://127.0.0.1:8080/api/task \
  -H "Content-Type: application/json" \
  -d @task_port.json
```

**预期输出：**
```json
{"success":true,"message":"Task sent to agent"}
```

**Agent 日志：**
```
2026-01-27T13:32:50.716+0800    INFO    transport/transfer.go:173    received task command    {"object_name": "collector", "data_type": 5050}
2026-01-27T13:32:50.716+0800    INFO    transport/transfer.go:199    task sent to plugin successfully    {"plugin": "collector"}
```

---

### 步骤五：验证数据库数据

```bash
# 查看所有表的数据量
PGPASSWORD=happy psql -h localhost -U postgres -d soc -c "
SELECT 'asset_host' as table_name, COUNT(*) as row_count FROM asset_host
UNION ALL SELECT 'asset_port', COUNT(*) FROM asset_port
UNION ALL SELECT 'asset_account', COUNT(*) FROM asset_account
UNION ALL SELECT 'asset_process', COUNT(*) FROM asset_process
UNION ALL SELECT 'asset_database', COUNT(*) FROM asset_database
UNION ALL SELECT 'asset_web_service', COUNT(*) FROM asset_web_service
UNION ALL SELECT 'asset_system_service', COUNT(*) FROM asset_system_service
UNION ALL SELECT 'asset_software', COUNT(*) FROM asset_software
UNION ALL SELECT 'asset_container', COUNT(*) FROM asset_container
UNION ALL SELECT 'asset_env_suspicious', COUNT(*) FROM asset_env_suspicious
UNION ALL SELECT 'asset_kmod', COUNT(*) FROM asset_kmod
ORDER BY row_count DESC;
"

# 查看主机信息
PGPASSWORD=happy psql -h localhost -U postgres -d soc -c \
  "SELECT agent_id, host_name, host_ip, os_type, os_version FROM asset_host;"

# 查看进程数据（前10条）
PGPASSWORD=happy psql -h localhost -U postgres -d soc -c \
  "SELECT name, path, run_name, status FROM asset_process LIMIT 10;"
```

---

## 测试数据文件

位置：`/home/work/goProject/src/BeeGuard/server/test_data/`

| 文件 | DataType | 用途 |
|------|----------|------|
| `task_process.json` | 5050 | 进程采集 |
| `task_port.json` | 5051 | 端口采集 |
| `task_user.json` | 5052 | 用户账号采集 |
| `task_service.json` | 5054 | 系统服务采集 |
| `task_software.json` | 5055 | 软件包采集 |
| `task_container.json` | 5056 | 容器采集 |
| `task_env_suspicious.json` | 5057 | 可疑环境变量检测 |
| `task_kmod.json` | 5062 | 内核模块采集 |
| `plugin_config.json` | - | 插件配置 |
| `send_task.sh` | - | 任务发送脚本 |
