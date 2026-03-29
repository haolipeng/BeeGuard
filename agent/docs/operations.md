# 运维指南

本文档描述 Agent 的日常运维操作，包括服务管理、状态检查、配置变更、升级和卸载。

---

## 一、服务管理（cloudsecctl）

Agent 安装后通过 systemd 管理，推荐使用 `cloudsecctl` 控制工具操作。

### 1.1 cloudsecctl 命令

```bash
CTL=/opt/cloudsec/agent/bin/cloudsecctl

sudo $CTL status           # 查看服务状态
sudo $CTL start            # 启动
sudo $CTL stop             # 停止
sudo $CTL restart          # 重启
sudo $CTL enable           # 开机自启（复制 service 文件到 systemd 并 enable）
sudo $CTL disable          # 取消开机自启
sudo $CTL service-reload   # 重载 systemd 配置（修改 service 文件后执行）
```

### 1.2 配置覆盖（set/unset）

`cloudsecctl set` 将配置写入 `/opt/cloudsec/agent/specified_env`，通过 systemd `EnvironmentFile` 注入，**不修改 agent.yaml**。修改后需重启生效。

```bash
# 设置服务端地址
sudo $CTL set --server="10.0.0.1:50051"

# 设置 Agent ID
sudo $CTL set --id="custom-agent-id"

# 清除覆盖配置（恢复使用 agent.yaml 中的值）
sudo $CTL unset --server
sudo $CTL unset --id

# 重启使配置生效
sudo $CTL restart
```

### 1.3 直接使用 systemctl

cloudsecctl 本质上是对 systemctl 的封装，也可以直接使用：

```bash
sudo systemctl status cloudsec-agent
sudo systemctl start cloudsec-agent
sudo systemctl stop cloudsec-agent
sudo systemctl restart cloudsec-agent
```

### 1.4 手动启停（无 systemd 场景）

```bash
# 前台启动
cd /opt/cloudsec/agent
sudo ./bin/agent

# 后台启动
cd /opt/cloudsec/agent
sudo nohup ./bin/agent > /opt/cloudsec/agent/logs/agent/agent.log 2>&1 &

# 优雅停止
sudo pkill -SIGTERM -f "/opt/cloudsec/agent/bin/agent"

# 强制停止（包括所有插件子进程）
sudo pkill -9 -f "/opt/cloudsec"
```

---

## 二、状态检查

### 2.1 服务状态

```bash
sudo /opt/cloudsec/agent/bin/cloudsecctl status
```

### 2.2 进程检查

```bash
# 检查 Agent 及插件进程
ps aux | grep -E "agent|collector|baseline|detector|ebpf_base_detector|nids|scanner" | grep -v grep

# 检查 gRPC 连接
ss -anp | grep agent
```

### 2.3 日志检查

```bash
# Agent 运行日志
tail -50 /opt/cloudsec/agent/logs/agent/agent.log

# 指定插件日志
tail -50 /opt/cloudsec/agent/logs/plugins/ebpf_base_detector/ebpf_base_detector.log
tail -50 /opt/cloudsec/agent/logs/plugins/detector/detector.log
```

---

## 三、日志管理

### 日志位置

| 组件 | 路径 |
|------|------|
| Agent | `/opt/cloudsec/agent/logs/agent/` |
| collector | `/opt/cloudsec/agent/logs/plugins/collector/` |
| baseline | `/opt/cloudsec/agent/logs/plugins/baseline/` |
| detector | `/opt/cloudsec/agent/logs/plugins/detector/` |
| ebpf_base_detector | `/opt/cloudsec/agent/logs/plugins/ebpf_base_detector/` |
| nids | `/opt/cloudsec/agent/logs/plugins/nids/` |
| scanner | `/opt/cloudsec/agent/logs/plugins/scanner/` |

### 日志轮转

Agent 使用 zap + lumberjack 日志轮转，通过 `agent.yaml` 配置：

```yaml
log:
  level: "info"        # debug/info/warn/error
  max_size: 10         # 单文件最大 MB
  max_backups: 5       # 保留旧文件数
  compress: false      # 是否压缩旧文件
```

### 手动清理

```bash
# 清理所有日志
sudo rm -rf /opt/cloudsec/agent/logs/agent/*
sudo rm -rf /opt/cloudsec/agent/logs/plugins/*

# 清理 7 天前的日志
sudo find /opt/cloudsec/agent/logs -name "*.log" -mtime +7 -delete
```

---

## 四、配置管理

### 4.1 配置文件

| 文件 | 说明 |
|------|------|
| `/opt/cloudsec/agent/agent.yaml` | 主配置文件 |
| `/opt/cloudsec/agent/specified_env` | 运行时覆盖配置（cloudsecctl set 写入） |
| `/opt/cloudsec/agent/plugins/*/config/` | 各插件配置目录 |

### 4.2 修改主配置

```bash
# 1. 编辑配置
sudo vim /opt/cloudsec/agent/agent.yaml

# 2. 重启生效
sudo /opt/cloudsec/agent/bin/cloudsecctl restart
```

### 4.3 修改插件配置

插件配置修改后同样需要重启 Agent 生效：

```bash
# 示例：修改高危命令规则
sudo vim /opt/cloudsec/agent/plugins/ebpf_base_detector/config/dangerous_commands.yaml

# 重启
sudo /opt/cloudsec/agent/bin/cloudsecctl restart
```

---

## 五、资源限制

systemd service 文件中配置了资源限制：

| 参数 | 值 | 说明 |
|------|------|------|
| MemoryMax | 500M | Agent 及所有插件最大内存 |
| CPUQuota | 20% | CPU 使用上限 |
| Restart | always | 异常退出自动重启 |
| RestartSec | 45 | 重启间隔（秒） |
| KillMode | control-group | 停止时杀死整个 cgroup |

如需调整，修改 service 文件后重载：

```bash
sudo vim /opt/cloudsec/agent/cloudsec-agent.service
sudo /opt/cloudsec/agent/bin/cloudsecctl service-reload
sudo /opt/cloudsec/agent/bin/cloudsecctl restart
```

---

## 六、升级

### 6.1 包升级（推荐）

```bash
# Debian/Ubuntu
sudo dpkg -i cloudsec-agent_<new_version>_<arch>.deb

# RHEL/CentOS
sudo rpm -U cloudsec-agent-<new_version>.<arch>.rpm
```

升级时 `agent.yaml` 及插件配置文件标记为 `noreplace`，不会被覆盖。

### 6.2 手动升级

```bash
# 1. 停止
sudo /opt/cloudsec/agent/bin/cloudsecctl stop

# 2. 备份（可选）
sudo cp -r /opt/cloudsec /opt/cloudsec.bak

# 3. 编译部署新版本
cd /home/work/goProject/src/BeeGuard/agent
make build && make deploy

# 4. 启动
sudo /opt/cloudsec/agent/bin/cloudsecctl start
```

---

## 七、卸载

### 7.1 包卸载

```bash
# Debian/Ubuntu
sudo dpkg -r cloudsec-agent

# RHEL/CentOS
sudo rpm -e cloudsec-agent
```

卸载过程自动完成：
1. 禁用 systemd 服务
2. 停止 Agent 进程
3. 清理运行时数据（`data/`、`logs/`、`specified_env`、`plugin.sock`）
4. 移除 systemd service 文件

配置文件保留在 `/opt/cloudsec/agent/`，如需完全清理：

```bash
# DEB 完全清除（含配置文件）
sudo dpkg -P cloudsec-agent

# RPM 卸载后手动清理
sudo rm -rf /opt/cloudsec
```

### 7.2 远程卸载

Server 端可通过 gRPC 下发卸载指令（DataType 1061），Agent 收到后：
1. 生成临时卸载脚本，通过 `systemd-run --scope` 在独立 cgroup 中启动
2. Agent 自行退出
3. 卸载脚本等待 Agent 进程退出后，自动检测包管理器（dpkg/rpm）执行卸载
4. 清理 `/opt/cloudsec/agent` 目录

详见 [Agent远程卸载测试](integration-test/agent-uninstall-testing.md)。

---

## 八、目录结构

```
/opt/cloudsec/agent/
├── bin/
│   ├── agent                    # Agent 主程序
│   └── cloudsecctl              # 控制工具
├── cloudsec-agent.service       # systemd service 文件
├── agent.yaml                   # 主配置文件
├── specified_env                # 运行时覆盖配置
├── btf/                         # BTF 文件（内核兼容）
├── plugins/
│   ├── collector/               # 资产采集
│   ├── baseline/                # 基线检查
│   ├── detector/                # 威胁检测
│   ├── ebpf_base_detector/      # eBPF 进程监控
│   ├── nids/                    # 网络入侵检测
│   └── scanner/                 # 病毒扫描
├── data/                        # 运行时数据
│   ├── agent/
│   └── plugins/
└── logs/                        # 日志
    ├── agent/
    └── plugins/
```

---

## 相关文档

- [编译部署](build-deploy.md) — 编译、打包和部署流程
- [配置详解](configuration.md) — 配置项完整说明
- [故障排查](troubleshooting.md) — 常见问题诊断
- [Agent远程卸载测试](integration-test/agent-uninstall-testing.md) — 远程卸载验证流程
