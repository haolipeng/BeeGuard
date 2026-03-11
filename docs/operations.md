# 运维指南

本文档描述 Agent 的日常运维操作。

---

## 一、启停管理

### 启动 Agent

```bash
# 前台启动
cd /opt/cloudsec/agent
sudo ./bin/agent

# 后台启动
cd /opt/cloudsec/agent
sudo nohup ./bin/agent > /opt/cloudsec/agent/logs/agent/agent.log 2>&1 &
```

### 停止 Agent

```bash
# 优雅停止
sudo pkill -SIGTERM -f "/opt/cloudsec/agent/bin/agent"

# 强制停止（包括所有插件）
sudo pkill -9 -f "/opt/cloudsec"
```

### 重启 Agent

```bash
sudo pkill -SIGTERM -f "/opt/cloudsec/agent/bin/agent" && sleep 2 && \
cd /opt/cloudsec/agent && \
sudo nohup ./bin/agent > /opt/cloudsec/agent/logs/agent/agent.log 2>&1 &
```

---

## 二、状态检查

```bash
# 检查 Agent 进程
ps aux | grep -E "agent|collector|baseline|detector|ebpf_base_detector|nids|scanner" | grep -v grep

# 检查连接状态
netstat -anp | grep agent

# 查看最近日志
tail -50 /opt/cloudsec/agent/logs/agent/agent.log
```

---

## 三、日志管理

### 日志位置

| 组件 | 路径 |
|------|------|
| Agent | `/opt/cloudsec/agent/logs/agent/` |
| 插件 | `/opt/cloudsec/agent/logs/plugins/<plugin>/` |

### 日志清理

```bash
# 清理所有日志
sudo rm -rf /opt/cloudsec/agent/logs/agent/*
sudo rm -rf /opt/cloudsec/agent/logs/plugins/*

# 清理 7 天前的日志
find /opt/cloudsec/agent/logs -name "*.log" -mtime +7 -delete
```

---

## 四、配置更新

```bash
# 1. 编辑配置
sudo vim /opt/cloudsec/agent/agent.yaml

# 2. 重启生效
sudo pkill -SIGTERM -f "/opt/cloudsec/agent/bin/agent"
cd /opt/cloudsec/agent
sudo ./bin/agent &
```

---

## 五、升级部署

```bash
# 1. 停止 Agent
sudo pkill -SIGTERM -f "/opt/cloudsec/agent/bin/agent"

# 2. 备份（可选）
sudo cp -r /opt/cloudsec /opt/cloudsec.bak

# 3. 部署新版本
cd /path/to/agent
make build && make deploy

# 4. 启动
cd /opt/cloudsec/agent
sudo ./bin/agent &
```

---

## 六、目录结构

```
/opt/cloudsec/agent/
├── bin/agent              # 主程序
├── agent.yaml             # 配置文件
├── plugins/               # 插件目录
├── data/                  # 运行时数据
└── logs/                  # 日志目录
```
