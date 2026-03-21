# Agent 一键安装命令测试指南

本文档描述如何在本地环境测试 `curl -fsSL http://127.0.0.1:8081/install.sh | sudo bash` 一键安装命令，包括环境准备、执行步骤、验证方法和故障排查。

---

## 一、概述

### 命令说明

```bash
curl -fsSL http://127.0.0.1:8081/install.sh | sudo bash
```

| 参数 | 含义 |
|------|------|
| `-f` | HTTP 错误时静默失败（返回非零退出码，不输出 HTML 错误页） |
| `-s` | 静默模式，不显示进度条和错误信息 |
| `-S` | 与 `-s` 配合，发生错误时仍显示错误信息 |
| `-L` | 跟随 HTTP 重定向（如 301/302） |
| `http://127.0.0.1:8081/install.sh` | server Server 的一键安装脚本端点 |
| `\| sudo bash` | 将下载的脚本通过管道传给 bash 以 root 权限执行 |

### 工作原理

```
                curl                           server Server (:8081)
┌──────────────────────┐  GET /install.sh     ┌────────────────────────────┐
│ 客户端               │────────────────────→ │ GetInstallScript()         │
│                      │                      │  - 渲染 install.sh.tpl     │
│                      │ ←─────────────────── │  - 注入 BaseURL + GRPCAddr │
│                      │  shell script        └────────────────────────────┘
│                      │
│  install.sh 执行：   │  GET /api1/agent/    ┌────────────────────────────┐
│  1. 检测包管理器     │  download?type=      │ DownloadPackage()          │
│  2. 检测系统架构     │  deb&arch=amd64      │  - 查找 package_dir 下的   │
│  3. 下载安装包  ────→│────────────────────→ │    .deb/.rpm 文件          │
│                      │ ←─────────────────── │  - 返回文件流              │
│  4. dpkg -i 安装     │  .deb package        └────────────────────────────┘
│  5. 验证服务状态     │
└──────────────────────┘
         │
         ▼ 安装完成后
┌──────────────────────┐  gRPC stream          ┌────────────────────────────┐
│ cloudsec-agent       │────────────────────→  │ server Server (:50051)      │
│ (systemd service)    │ ←─────────────────── │  - 接收 Agent 数据         │
│                      │  plugin configs      │  - 下发插件配置和任务      │
└──────────────────────┘                       └────────────────────────────┘
```

### 涉及的 API 端点

| 方法 | 端点 | 功能 | 认证 |
|------|------|------|------|
| GET | `/install.sh` | 返回动态生成的安装脚本 | 无需认证 |
| GET | `/api1/agent/download?type=deb\|rpm&arch=amd64\|arm64` | 下载安装包 | 无需认证 |
| GET | `/api1/agent/packages` | 列出可用安装包 | 无需认证 |

### 安装脚本执行流程

安装脚本（`install.sh.tpl` 渲染后）按以下步骤执行：

1. **权限检查** — 必须以 root 运行
2. **systemctl 检查** — 系统必须支持 systemd
3. **检测包管理器** — 优先 dpkg (DEB)，其次 rpm (RPM)
4. **检测系统架构** — x86_64 → amd64，aarch64 → arm64
5. **下载安装包** — 从 server 下��� .deb 或 .rpm
6. **安装包** — `dpkg -i` 或 `rpm -i`
7. **设置 gRPC 地址** — 通过 `SPECIFIED_SERVER` 环境变量传递给 postinstall 脚本
8. **验证服务状态** — 检查 cloudsec-agent 服务是否启动

---

## 二、前置条件

| 条件 | 说明 | 检查命令 |
|------|------|---------|
| server 已部署 | `/opt/cloudsec/server/bin/server` 存在 | `ls /opt/cloudsec/server/bin/server` |
| 安装包已就位 | `package_dir` 下有 .deb 或 .rpm 文件 | `ls /opt/cloudsec/server/packages/` |
| PostgreSQL 运行中 | 本地数据库可访问 | `systemctl is-active postgresql` |
| root 权限 | 安装脚本需要 root | `id -u` 应为 0 |
| curl 已安装 | 系统有 curl | `which curl` |
| systemd 可用 | 系统支持 systemctl | `which systemctl` |

---

## 三、环境准备

### 3.1 确认安装包存在

```bash
ls -la /opt/cloudsec/server/packages/
```

预期输出应包含 `.deb` 文件（Ubuntu/Debian 环境）：

```
-rw-r--r-- 1 root root 43171548 Mar 17 22:00 cloudsec-agent_914d33b-dirty-1_amd64.deb
```

如果目录为空或没有对应架构的包，需先编译打包：

```bash
cd /home/work/goProject/src/company/agent
make package-deb
cp build/cloudsec-agent_*.deb /opt/cloudsec/server/packages/
```

### 3.2 修改 server 配置（本地测试）

server 默认配置指向远程服务器，本地测试需修改为本地地址。

```bash
# 备份原始配置
sudo cp /opt/cloudsec/server/conf/server.yaml /opt/cloudsec/server/conf/server.yaml.bak
```

编辑 `/opt/cloudsec/server/conf/server.yaml`，修改以下两处：

**修改 1 — 数据库指向本地 PostgreSQL：**

```yaml
database:
  host: 127.0.0.1
  port: 5432
  user: postgres
  password: "root"
  database: soc
```

**修改 2 — install.server_addr 指向本地 gRPC：**

```yaml
install:
  enabled: true
  package_dir: /opt/cloudsec/server/packages
  server_addr: "127.0.0.1:50051"
```

> **重要**：`server_addr` 决定了安装脚本中 Agent 连接的 gRPC 地址。本地测试必须改为 `127.0.0.1:50051`，否则 Agent 安装后会尝试连接远程服务器。

### 3.3 数据库准备

```bash
# 确认 PostgreSQL 可访问
PGPASSWORD=root psql -h 127.0.0.1 -p 5432 -U postgres -c "SELECT 1;"

# 创建数据库（如不存在）
PGPASSWORD=root psql -h 127.0.0.1 -p 5432 -U postgres -c "CREATE DATABASE soc;" 2>/dev/null || true
```

### 3.4 清理已有安装（如果之前安装过）

如果系统上已有 cloudsec-agent，先卸载以验证全新安装流程：

```bash
# 停止服务
sudo systemctl stop cloudsec-agent 2>/dev/null

# 卸载包
sudo dpkg --purge cloudsec-agent 2>/dev/null

# 清理残留目录
sudo rm -rf /opt/cloudsec/agent

# 确认已卸载
dpkg -l | grep cloudsec
```

---

## 四、测试步骤

### 4.1 启动 server Server

打开 **Terminal A**：

```bash
cd /opt/cloudsec/server
sudo ./bin/server -config conf/server.yaml
```

**启动成功判定** — 必须看到以下两行日志：

```
INFO  gRPC Server 启动，监听端口 :50051
INFO  [HTTP] HTTP API Server 启动，监听端口 :8081
```

### 4.2 验证安装接口可用

打开 **Terminal B**，逐步验证三个端点：

**验证 1 — 安装脚本端点：**

```bash
curl -fsSL http://127.0.0.1:8081/install.sh
```

预期输出：一个完整的 bash 脚本，开头为：

```bash
#!/bin/bash
set -e

BASE_URL="http://127.0.0.1:8081"
GRPC_ADDR="127.0.0.1:50051"
...
```

验证要点：
- `BASE_URL` 为 `http://127.0.0.1:8081`（从请求 Host 头自动提取）
- `GRPC_ADDR` 为 `127.0.0.1:50051`（来自 server.yaml 中 `install.server_addr`）
- 脚本内容完整，无模板语法残留（如 `{{.BaseURL}}`）

**验证 2 — 安装包列表端点：**

```bash
curl -s http://127.0.0.1:8081/api1/agent/packages | python3 -m json.tool
```

预期输出：

```json
{
    "package_dir": "/opt/cloudsec/server/packages",
    "packages": [
        {
            "name": "cloudsec-agent_914d33b-dirty-1_amd64.deb",
            "size": 43171548
        }
    ]
}
```

**验证 3 — 安装包下载端点：**

```bash
# 测试下载（只看 HTTP 状态码，不保存文件）
curl -s -o /dev/null -w '%{http_code}\n' 'http://127.0.0.1:8081/api1/agent/download?type=deb&arch=amd64'
```

预期输出：`200`

### 4.3 执行一键安装

确认以上三个验证均通过后，执行一键安装：

```bash
curl -fsSL http://127.0.0.1:8081/install.sh | sudo bash
```

**预期输出流程：**

```
[INFO] 检测到包管理器类型: deb
[INFO] 检测到系统架构: amd64
[INFO] 正在下载安装包: http://127.0.0.1:8081/api1/agent/download?type=deb&arch=amd64
[INFO] 安装包下载完成
[INFO] 正在安装 Agent...
...（dpkg 安装输出）...
...（postinstall.sh 输出）...
[INFO] Agent 安装成功，服务已启动
[INFO] gRPC 服务器地址: 127.0.0.1:50051
● cloudsec-agent.service - ...
     Active: active (running) ...
```

### 4.4 验证安装结果

```bash
# 检查服务状态
sudo systemctl status cloudsec-agent

# 检查安装目录
ls -la /opt/cloudsec/agent/bin/
ls -la /opt/cloudsec/agent/plugins/

# 检查进程
ps aux | grep cloudsec-agent

# 检查 gRPC 连接（在 server Terminal A 中应看到 Agent 连接日志）
# INFO  [Transfer] Agent 连接: agent_id=xxx hostname=xxx
```

### 4.5 验证 Agent 与 server 通信

```bash
# 查询在线 Agent
curl -s http://127.0.0.1:8081/api/agents | python3 -m json.tool

# 查询数据库中的 Agent 信息
PGPASSWORD=root psql -h 127.0.0.1 -p 5432 -U postgres -d soc -c \
  "SELECT agent_id, host_name, host_ip, connection_status FROM agent_info;"
```

---

## 五、可能遇到的问题及解决方案

### 5.1 curl 返回错误

| 现象 | 原因 | 解决方案 |
|------|------|---------|
| `curl: (7) Failed to connect to 127.0.0.1 port 8081` | server 未启动 | 启动 server Server（步骤 4.1） |
| `curl: (22) The requested URL returned error: 500` | 模板渲染失败 | 查看 server 日志排查 |
| 输出 HTML 而非 shell 脚本 | 请求被其他服务拦截 | 确认 8081 端口是 server 在监听：`ss -tlnp \| grep 8081` |

### 5.2 安装脚本执行失败

| 现象 | 原因 | 解决方案 |
|------|------|---------|
| `请使用 root 权限运行此脚本` | 未加 sudo | 使用 `sudo bash` 或 `sudo su` 后执行 |
| `系统不支持 systemctl，无法安装` | 非 systemd 系统 | 仅支持 systemd 的 Linux |
| `未检测到 dpkg 或 rpm 包管理器` | 非标准发行版 | 仅支持 Debian/Ubuntu (dpkg) 或 CentOS/RHEL (rpm) |
| `不支持的架构: xxx` | 非 x86_64/aarch64 | 仅支持 amd64 和 arm64 |

### 5.3 安装包下载失败

| 现象 | 原因 | 解决方案 |
|------|------|---------|
| HTTP 404 | `package_dir` 下没有匹配的包 | 检查包名是否包含正确架构标识。执行 `curl -s http://127.0.0.1:8081/api1/agent/packages` 查看可用包 |
| HTTP 400 `参数 type 必须为 deb 或 rpm` | 包类型不匹配 | 确认系统有 dpkg 或 rpm |

### 5.4 dpkg 安装失败

| 现象 | 原因 | 解决方案 |
|------|------|---------|
| `dpkg: dependency problems` | 缺少依赖 | `sudo apt-get install -f` |
| `trying to overwrite ...` | 文件冲突 | 先卸载旧版本：`sudo dpkg --purge cloudsec-agent` |
| `package architecture (amd64) does not match system (arm64)` | 架构不匹配 | 编译对应架构的包 |

### 5.5 Agent 服务启动失败

```bash
# 查看服务状态
sudo systemctl status cloudsec-agent

# 查看详细日志
sudo journalctl -u cloudsec-agent -f

# 检查配置文件中的 server 地址
cat /opt/cloudsec/agent/agent.yaml
```

常见原因：
- `agent.yaml` 中 `server` 地址不是 `127.0.0.1:50051` → 通过 `cloudsecctl set --server=127.0.0.1:50051` 修改
- server 未启动或 gRPC 端口不可达 → 先启动 server

---

## 六、测试后清理

### 6.1 停止服务

```bash
# 停止 Agent
sudo systemctl stop cloudsec-agent

# 停止 server（Terminal A 中 Ctrl+C）
```

### 6.2 卸载 Agent（可选）

```bash
# 卸载 deb 包
sudo dpkg --purge cloudsec-agent

# 清理安装目录
sudo rm -rf /opt/cloudsec/agent
```

### 6.3 恢复 server 配置

```bash
# 恢复原始配置
sudo cp /opt/cloudsec/server/conf/server.yaml.bak /opt/cloudsec/server/conf/server.yaml
```

---

## 七、相关源码说明

| 文件 | 说明 |
|------|------|
| `server/internal/controller/install/install.go` | 安装控制器，实现三个 HTTP 端点 |
| `server/internal/controller/install/install.sh.tpl` | 安装脚本模板，使用 Go template 渲染 |
| `server/internal/router/router.go` | 路由注册，`install.enabled=true` 时注册安装路由 |
| `server/conf/server.yaml` | 服务端配置，包含 `install` 段 |
| `agent/deploy/scripts/postinstall.sh` | DEB/RPM 安装后脚本，创建目录、启用服务、设置环境变量 |
| `agent/deploy/nfpm.yaml` | NFPM 打包配置，定义包内容和安装脚本 |

### 关键配置项

```yaml
# server.yaml
install:
  enabled: true                                  # 启用一键安装功能
  package_dir: /opt/cloudsec/server/packages      # 安装包存放目录
  server_addr: "127.0.0.1:50051"                 # Agent gRPC 连接地址
```

- `enabled`：控制是否注册 `/install.sh` 等路由，`false` 时访问返回 404
- `package_dir`：server 从此目录查找 .deb/.rpm 文件，需手动放入编译产物
- `server_addr`：写入安装脚本的 `GRPC_ADDR` 变量；留空时从 HTTP 请求的 Host 头自动提取 IP，拼接 gRPC 端口
