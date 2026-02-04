# Agent 编译部署文档

本文档指导如何编译、部署和运行 Agent 及其插件。

---

## 一、概述

Agent 是安装在目标主机上的安全探针程序，负责：
- 与 Server (hcids) 建立 gRPC 连接
- 加载并管理插件（collector、baseline、detector、driver）
- 采集主机资产信息、执行基线检查、检测安全威胁并上报 Server

### 插件说明

| 插件 | 功能 | 说明 |
|------|------|------|
| collector | 资产采集 | 采集主机资产信息（进程、端口、用户等） |
| baseline | 基线检查 | 安全基线合规检查 |
| detector | 威胁检测 | SSH/FTP 暴力破解检测、异常登录检测 |
| driver | eBPF 驱动 | 基于 eBPF 的进程监控、高危命令检测 |

---

## 二、编译环境要求

### 基础环境

- Go 1.25+
- Make

### eBPF 编译环境 (driver 插件)

driver 插件使用 eBPF 技术，编译需要以下工具：

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

# 部署到 /opt/cloudsec/
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

### 3.4 仅编译部署 driver 插件

```bash
# 仅编译 driver 插件 (自动生成 eBPF 代码)
make build-driver

# 仅部署 driver 插件
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
    └── driver                  # eBPF 驱动插件
```

### 部署后目录结构

```
/opt/cloudsec/
├── bin/
│   └── agent                   # Agent 主程序
├── plugins/
│   ├── collector/
│   │   └── collector           # 采集插件
│   ├── baseline/
│   │   └── baseline            # 基线检查插件
│   ├── detector/
│   │   └── detector            # 威胁检测插件
│   └── driver/
│       ├── driver              # eBPF 驱动插件
│       └── config/
│           └── dangerous_commands.yaml  # 高危命令规则
├── conf/
│   └── agent.yaml              # 配置文件
├── data/
│   ├── agent/                  # Agent 运行时数据
│   └── plugins/
│       ├── collector/          # 采集插件数据
│       ├── baseline/           # 基线插件数据
│       ├── detector/           # 检测插件数据
│       └── driver/             # 驱动插件数据
└── logs/
    ├── agent/                  # Agent 日志
    └── plugins/
        ├── collector/          # 采集插件日志
        ├── baseline/           # 基线插件日志
        ├── detector/           # 检测插件日志
        └── driver/             # 驱动插件日志
```

---

## 四、配置文件

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
working_directory: "/opt/cloudsec/data/agent"

# 插件目录
plugins_directory: "/opt/cloudsec/plugins"

# 连接失败最大重试次数
retry_max_count: 10

# 重试间隔（秒）
retry_interval: 5
```

---

## 五、运行

### 前提条件

1. Server (hcids) 已启动并监听 50051 端口
2. 配置文件中 `server` 地址正确
3. 使用 root 权限运行（采集系统信息、eBPF 需要）

### 启动方式

```bash
# 方式一：使用部署目录（推荐）
sudo /opt/cloudsec/bin/agent -config /opt/cloudsec/conf/agent.yaml

# 方式二：使用源码目录（开发调试）
cd /home/work/goProject/src/company/agent
sudo ./build/agent -config agent.yaml

# 方式三：后台运行
sudo nohup /opt/cloudsec/bin/agent -config /opt/cloudsec/conf/agent.yaml \
    > /opt/cloudsec/logs/agent/agent.log 2>&1 &
```

### 启动日志

```
2026-01-27T10:00:00.000+0800    INFO    config initialized successfully
2026-01-27T10:00:00.010+0800    INFO    connected to server     {"server": "127.0.0.1:50051"}
2026-01-27T10:00:00.020+0800    INFO    collector plugin loaded successfully
2026-01-27T10:00:00.030+0800    INFO    baseline plugin loaded successfully
2026-01-27T10:00:00.040+0800    INFO    detector plugin loaded successfully
2026-01-27T10:00:00.050+0800    INFO    driver plugin loaded successfully
```

---

## 六、清理

```bash
# 停止 Agent
sudo pkill -f "/opt/cloudsec/bin/agent"

# 清理日志
sudo rm -rf /opt/cloudsec/logs/agent/*
sudo rm -rf /opt/cloudsec/logs/plugins/*

# 清理运行时数据
sudo rm -rf /opt/cloudsec/data/*

# 清理编译产物
cd /home/work/goProject/src/company/agent
make clean
```

