# Agent 编译部署文档

本文档指导如何编译、部署和运行 Agent 及其插件。

---

## 一、概述

Agent 是安装在目标主机上的指标数据采集探针程序，负责：
- 与 Server (hcids) 建立 gRPC 连接
- 加载并管理插件（collector、baseline）
- 采集主机资产信息并上报 Server

---

## 二、编译部署

## 2、1 编译部署Agent + 所有插件

```bash
cd /home/work/goProject/src/company/agent

# 编译 Agent + 所有插件
make build-all

# 部署到 /opt/cloudsec/
make deploy
```

### 2、2 仅编译部署Agent

```bash
# 仅编译所有插件
make build-plugins

# 仅部署 Agent
make deploy-agent
```



## 2、3 仅编译部署所有插件

```
# 仅编译所有插件
make build-plugins

# 仅部署插件
make deploy-plugins
```



### 编译产物

```
build/
├── agent                       # Agent 主程序
└── plugins/
    ├── collector               # 采集插件
    └── baseline                # 基线检查插件
```

### 部署后目录结构

```
/opt/cloudsec/
├── bin/
│   └── agent                   # Agent 主程序
├── plugins/
│   ├── collector               # 采集插件
│   └── baseline                # 基线检查插件
├── conf/
│   └── agent.yaml              # 配置文件
├── data/
│   ├── agent/                  # Agent 运行时数据
│   └── plugins/
│       └── collector/          # 插件运行时数据
└── logs/
    ├── agent/                  # Agent 日志
    └── plugins/
        ├── collector/          # 采集插件日志
        └── baseline/           # 基线插件日志
```

---

## 三、配置文件

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

## 四、运行

### 前提条件

1. Server (hcids) 已启动并监听 50051 端口
2. 配置文件中 `server` 地址正确
3. 使用 root 权限运行（采集系统信息需要）

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
```

---

## 五、清理

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
