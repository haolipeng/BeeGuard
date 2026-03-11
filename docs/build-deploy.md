# Agent 编译部署文档

本文档指导如何编译、部署和运行 Agent 及其插件。

---

## 一、概述

Agent 是安装在目标主机上的安全探针程序，负责：
- 与 Server (hcids) 建立 gRPC 连接
- 加载并管理插件（collector、baseline、detector、ebpf_base_detector、nids、scanner）
- 采集主机资产信息、执行基线检查、检测安全威胁并上报 Server

### 插件说明

| 插件 | 功能 | 说明 |
|------|------|------|
| collector | 资产采集 | 采集主机资产信息（进程、端口、用户等） |
| baseline | 基线检查 | 安全基线合规检查 |
| detector | 威胁检测 | SSH/FTP 暴力破解检测、异常登录检测 |
| ebpf_base_detector | eBPF 驱动 | 基于 eBPF 的进程监控、高危命令检测 |
| nids | 网络入侵检测 | 基于网络流量分析的攻击检测 |
| scanner | 病毒扫描 | 恶意文件扫描和检出 |

---

## 二、编译环境要求

### 基础环境

- Go 1.25+
- Make

### eBPF 编译环境 (ebpf_base_detector 插件)

ebpf_base_detector 插件使用 eBPF 技术，编译需要以下工具：

```bash
# Ubuntu/Debian
apt install clang llvm libbpf-dev linux-headers-$(uname -r)

# 验证安装
clang --version          # >= 15.0
llvm-strip --version
ls /usr/include/bpf/     # libbpf 头文件
ls /sys/kernel/btf/vmlinux  # BTF 支持
```

---

## 三、编译部署

### 3.1 编译部署 Agent + 所有插件

```bash
cd /home/work/goProject/src/company/agent

# 编译 Agent + 所有插件 (自动生成 eBPF 代码)
make build

# 部署到 /opt/cloudsec/agent/
make deploy
```

### 3.2 仅编译部署 Agent

```bash
# 仅编译 Agent
make build-agent

# 仅部署 Agent
make deploy-agent
```

### 3.3 仅编译部署所有插件

```bash
# 仅编译所有插件
make build-plugins

# 仅部署插件
make deploy-plugins
```

### 3.4 仅编译部署 ebpf_base_detector 插件

```bash
# 仅编译 ebpf_base_detector 插件 (自动生成 eBPF 代码)
make build-driver

# 仅部署 ebpf_base_detector 插件
make deploy-driver
```

### 3.5 手动生成 eBPF 代码

如需单独重新生成 eBPF 代码：

```bash
make generate-ebpf
```

### 编译产物

```
build/
├── agent                       # Agent 主程序
└── plugins/
    ├── collector               # 采集插件
    ├── baseline                # 基线检查插件
    ├── detector                # 威胁检测插件
    ├── ebpf_base_detector      # eBPF 驱动插件
    ├── nids                    # 网络入侵检测插件
    └── scanner                 # 病毒扫描插件
```

### 部署后目录结构

```
/opt/cloudsec/agent/
├── bin/
│   └── agent                   # Agent 主程序
├── agent.yaml                  # 配置文件
├── plugins/
│   ├── collector/
│   │   └── collector           # 采集插件
│   ├── baseline/
│   │   └── baseline            # 基线检查插件
│   ├── detector/
│   │   └── detector            # 威胁检测插件
│   ├── ebpf_base_detector/
│   │   ├── ebpf_base_detector  # eBPF 驱动插件
│   │   └── config/
│   │       └── dangerous_commands.yaml  # 高危命令规则
│   ├── nids/
│   │   ├── nids                # 网络入侵检测插件
│   │   └── config/
│   │       ├── nids.yaml       # NIDS 配置
│   │       └── nids.rules      # 检测规则
│   └── scanner/
│       ├── scanner             # 病毒扫描插件
│       └── config/
│           └── scanner.yaml    # 扫描配置
├── data/
│   ├── agent/                  # Agent 运行时数据
│   └── plugins/
│       ├── collector/          # 采集插件数据
│       ├── baseline/           # 基线插件数据
│       ├── detector/           # 检测插件数据
│       ├── ebpf_base_detector/ # 驱动插件数据
│       ├── nids/               # NIDS 插件数据
│       └── scanner/            # 扫描插件数据
└── logs/
    ├── agent/                  # Agent 日志
    └── plugins/
        ├── collector/          # 采集插件日志
        ├── baseline/           # 基线插件日志
        ├── detector/           # 检测插件日志
        ├── ebpf_base_detector/ # 驱动插件日志
        ├── nids/               # NIDS 插件日志
        └── scanner/            # 扫描插件日志
```

---

## 四、打包

项目使用 [nfpm](https://github.com/goreleaser/nfpm) 生成 DEB/RPM 安装包，打包配置位于 `deploy/nfpm.yaml`。

### 4.1 安装 nfpm

```bash
# 方式一：go install
go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest

# 方式二：下载预编译二进制
# https://github.com/goreleaser/nfpm/releases
```

### 4.2 执行打包

```bash
# 生成 DEB 包（自动先执行 build-all）
make package-deb

# 生成 RPM 包
make package-rpm

# 同时生成 DEB + RPM
make package
```

打包产物输出到 `build/` 目录，文件名格式为 `cloudsec-agent_<version>_<arch>.deb` 或 `cloudsec-agent-<version>.<arch>.rpm`。

### 安装包内容

安装包包含以下文件：

| 类别 | 文件 | 安装路径 |
|------|------|----------|
| 二进制 | agent | `/opt/cloudsec/agent/bin/agent` |
| 二进制 | cloudsecctl | `/opt/cloudsec/agent/bin/cloudsecctl` |
| 服务 | cloudsec-agent.service | `/opt/cloudsec/agent/cloudsec-agent.service` |
| 配置 | agent.yaml | `/opt/cloudsec/agent/agent.yaml` |
| 插件 | collector, baseline, detector, ebpf_base_detector, nids, scanner | `/opt/cloudsec/agent/plugins/<name>/<name>` |
| 插件配置 | detector rules, ebpf configs, baseline configs, nids configs, scanner config | `/opt/cloudsec/agent/plugins/<name>/config/` |

---

## 五、客户部署

### 5.1 系统要求

- Linux（支持 systemd）
- root 权限
- 内核支持 BTF（ebpf_base_detector 插件需要，内核 >= 5.8 推荐）

### 5.2 安装

```bash
# Debian/Ubuntu
sudo dpkg -i cloudsec-agent_<version>_<arch>.deb

# RHEL/CentOS
sudo rpm -i cloudsec-agent-<version>.<arch>.rpm
```

安装过程自动完成：
1. 检查 systemd 环境（`preinstall.sh`）
2. 创建运行时目录（`/opt/cloudsec/agent/data/`、`/opt/cloudsec/agent/logs/`）
3. 注册并启用 systemd 服务
4. 启动 agent

### 5.3 安装时指定服务端地址

通过环境变量在安装时指定 Server 地址和 Agent ID：

```bash
# DEB
sudo SPECIFIED_SERVER="10.0.0.1:50051" dpkg -i cloudsec-agent_*.deb

# RPM
sudo SPECIFIED_SERVER="10.0.0.1:50051" rpm -i cloudsec-agent-*.rpm

# 同时指定 Agent ID
sudo SPECIFIED_SERVER="10.0.0.1:50051" SPECIFIED_AGENT_ID="agent-001" dpkg -i cloudsec-agent_*.deb
```

### 5.4 安装后管理

使用 `cloudsecctl` 工具管理服务：

```bash
sudo /opt/cloudsec/agent/bin/cloudsecctl status     # 查看状态
sudo /opt/cloudsec/agent/bin/cloudsecctl start      # 启动
sudo /opt/cloudsec/agent/bin/cloudsecctl stop       # 停止
sudo /opt/cloudsec/agent/bin/cloudsecctl restart    # 重启
sudo /opt/cloudsec/agent/bin/cloudsecctl set --server="<addr>"  # 修改服务端地址
sudo /opt/cloudsec/agent/bin/cloudsecctl set --id="<id>"        # 修改 Agent ID
```

### 5.5 升级

直接安装新版本包即可，升级过程会自动重新加载服务：

```bash
# DEB
sudo dpkg -i cloudsec-agent_<new_version>_<arch>.deb

# RPM
sudo rpm -U cloudsec-agent-<new_version>.<arch>.rpm
```

升级时配置文件（`agent.yaml` 等标记为 `config|noreplace` 的文件）不会被覆盖。

### 5.6 卸载

```bash
# Debian/Ubuntu
sudo dpkg -r cloudsec-agent

# RHEL/CentOS
sudo rpm -e cloudsec-agent
```

卸载时自动完成：
1. 停止 agent 服务
2. 移除 systemd 服务注册
3. 清理运行时数据和日志（`/opt/cloudsec/agent/data/`、`/opt/cloudsec/agent/logs/`）

配置文件在卸载后保留，如需完全清理：

```bash
# DEB 完全清除（含配置文件）
sudo dpkg -P cloudsec-agent

# RPM 卸载后手动清理
sudo rm -rf /opt/cloudsec
```

---

## 六、配置文件（部署后修改）

### 配置文件查找优先级

1. 命令行参数 `-config` 指定的路径
2. `/etc/cloudsec-agent/agent.yaml`
3. 当前目录 `agent.yaml`

### 配置项说明

```yaml
# Server 连接地址
server: "127.0.0.1:50051"

# 连接超时时间（秒）
connect_timeout: 30

# Agent 工作目录（存放运行时数据）
working_directory: "/var/run/cloudsec-agent"

# 插件目录
plugins_directory: "/opt/cloudsec/agent/plugins"

# 日志目录
log_directory: "/opt/cloudsec/agent/logs"

# 连接失败最大重试次数
retry_max_count: 10

# 重试间隔（秒）
retry_interval: 5
```

---

## 七、运行

### 前提条件

1. Server (hcids) 已启动并监听 50051 端口
2. 配置文件中 `server` 地址正确
3. 使用 root 权限运行（采集系统信息、eBPF 需要）

### 启动方式

```bash
# 方式一：使用部署目录（推荐）
cd /opt/cloudsec/agent
sudo ./bin/agent

# 方式二：使用源码目录（开发调试）
# 日志输出到当前目录下的 logs/agent.log
cd /home/work/goProject/src/company/agent
make build
make deploy
cd /opt/cloudsec/agent
sudo ./bin/agent

# 方式三：后台运行
cd /opt/cloudsec/agent
sudo nohup ./bin/agent > /opt/cloudsec/agent/logs/agent/agent.log 2>&1 &
```

### 启动日志

```
2026-01-27T10:00:00.000+0800    INFO    config initialized successfully
2026-01-27T10:00:00.010+0800    INFO    connected to server     {"server": "127.0.0.1:50051"}
2026-01-27T10:00:00.020+0800    INFO    collector plugin loaded successfully
2026-01-27T10:00:00.030+0800    INFO    baseline plugin loaded successfully
2026-01-27T10:00:00.040+0800    INFO    detector plugin loaded successfully
2026-01-27T10:00:00.050+0800    INFO    ebpf_base_detector plugin loaded successfully
```

---

## 八、清理

```bash
# 停止 Agent
sudo pkill -f "/opt/cloudsec/agent/bin/agent"

# 清理日志
sudo rm -rf /opt/cloudsec/agent/logs/agent/*
sudo rm -rf /opt/cloudsec/agent/logs/plugins/*

# 清理运行时数据
sudo rm -rf /opt/cloudsec/agent/data/*

# 清理编译产物
cd /home/work/goProject/src/company/agent
make clean
```

