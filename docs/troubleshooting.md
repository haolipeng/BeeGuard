# 故障排查指南

本文档描述 Agent 常见问题及排查方法。

---

## 一、连接问题

### Agent 无法连接 Server

**现象：** 日志中出现连接失败错误

**排查步骤：**
```bash
# 1. 检查 Server 是否启动
netstat -tlnp | grep 50051

# 2. 检查网络连通性
telnet <server_ip> 50051

# 3. 检查配置文件中 server 地址
cat /opt/cloudsec/agent/agent.yaml | grep server

# 4. 查看 Agent 日志
tail -f /opt/cloudsec/agent/logs/agent/agent.log
```

---

## 二、插件问题

### 插件加载失败

**现象：** 日志中出现 `plugin load failed`

**排查步骤：**
```bash
# 1. 检查插件文件是否存在
ls -la /opt/cloudsec/agent/plugins/

# 2. 检查文件权限
chmod +x /opt/cloudsec/agent/plugins/*/

# 3. 检查依赖（ebpf_base_detector 插件需要 eBPF 环境）
ls /sys/kernel/btf/vmlinux
```

### 插件无数据上报

**可能原因：**
- 未调用 `client.Flush()`
- 配置文件错误
- 白名单配置导致过滤

**排查：** 使用 Standalone 模式本地测试
```bash
cd /opt/cloudsec/agent
sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=/opt/cloudsec/agent/logs/agent.log -test
```

---

## 三、eBPF/ebpf_base_detector 问题

### eBPF 加载失败

**现象：** `failed to load eBPF program`

**排查步骤：**
```bash
# 1. 检查内核版本 (需要 >= 5.x)
uname -r

# 2. 检查 BTF 支持
ls /sys/kernel/btf/vmlinux

# 3. 检查 root 权限
whoami
```

### 未检测到高危命令

**排查步骤：**
```bash
# 1. 检查规则配置
cat /opt/cloudsec/agent/plugins/ebpf_base_detector/config/dangerous_commands.yaml

# 2. 确认 ebpf_base_detector 已启动
ps aux | grep ebpf_base_detector
```

---

## 四、日志位置

| 组件 | 日志路径 |
|------|----------|
| Agent | `/opt/cloudsec/agent/logs/agent/agent.log` |
| Collector | `/opt/cloudsec/agent/logs/plugins/collector/collector.log` |
| Baseline | `/opt/cloudsec/agent/logs/plugins/baseline/baseline.log` |
| Detector | `/opt/cloudsec/agent/logs/plugins/detector/detector.log` |
| ebpf_base_detector | `/opt/cloudsec/agent/logs/plugins/ebpf_base_detector/ebpf_base_detector.log` |

---

## 五、快速诊断命令

```bash
# 检查 Agent 进程
ps aux | grep agent

# 检查插件进程
ps aux | grep -E "collector|baseline|detector|ebpf_base_detector|nids|scanner"

# 查看最近错误
grep -i error /opt/cloudsec/agent/logs/agent/agent.log | tail -20

# 检查端口占用
netstat -tlnp | grep agent
```
