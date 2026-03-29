# 告警触发测试脚本

`scripts/` 目录下提供了一组 shell 脚本，用于快速触发各插件的入侵检测告警，免去每次手动逐条输入命令。

---

## 快速开始

```bash
# 终端 A：启动 Agent（按需选择模式）
cd /opt/cloudsec
sudo ./bin/agent -standalone -plugins=ebpf_base_detector,detector,nids -output=stderr -test   # Standalone
sudo ./bin/agent -config agent.yaml -test                                                      # 集成测试

# 终端 B：运行全部告警触发
cd /home/work/goProject/src/BeeGuard/agent
sudo bash scripts/test-all-alerts.sh
```

---

## 脚本列表

### 统一入口

| 脚本 | 说明 |
|------|------|
| `test-all-alerts.sh` | 按插件分组串联调用下方所有脚本，支持按组运行 |

```bash
sudo bash scripts/test-all-alerts.sh              # 全部
sudo bash scripts/test-all-alerts.sh ebpf          # 仅 eBPF 类
sudo bash scripts/test-all-alerts.sh detector      # 仅 Detector 类
sudo bash scripts/test-all-alerts.sh nids          # 仅 NIDS
sudo bash scripts/test-all-alerts.sh scanner       # 仅 Scanner
```

### ebpf_base_detector 插件

| 脚本 | DataType | 触发内容 | 依赖 |
|------|----------|---------|------|
| `test-dangerous-commands.sh` | 6003 | 危险删除、敏感文件访问、危险权限修改、内核模块操作 | 无 |
| `test-privilege-escalation.sh` | 6006 | SUID 程序提权 + 白名单验证 | gcc, 非 root 用户 |
| `test-reverse-shell.sh` | 6004 | nc -e / Python dup2 / bash /dev/tcp 反弹 | netcat-traditional, python3 |
| `test-malicious-requests.sh` | 6008 | 矿池端口/矿池域名/C2域名/C2端点/钓鱼域名 | nc, dig |
| `test-file-integrity.sh` | 6009 | /etc/cron.d 文件创建/修改/删除、/etc/hosts 修改 | 无 |

### Detector 插件

| 脚本 | DataType | 触发内容 | 依赖 |
|------|----------|---------|------|
| `test-ssh-bruteforce.sh` | 6001 | 10 次 SSH 错误密码登录 | sshpass, sshd |
| `test-ftp-bruteforce.sh` | 6002 | 10 次 FTP 错误认证 | vsftpd |
| `test-ssh-anomaly-login.sh` | 6005 | 从非白名单 IP 成功 SSH 登录 | ssh_anomaly_login 规则配置 |

### NIDS 插件

| 脚本 | DataType | 触发内容 | 依赖 |
|------|----------|---------|------|
| `test-nids.sh` | 6007 | 12 条规则：Log4j/SQLi/CMDi/路径遍历/Struts2/Spring4Shell/Fastjson/扫描器 | nginx (80端口) |

### Scanner 插件

| 脚本 | DataType | 触发内容 | 依赖 |
|------|----------|---------|------|
| `test-scanner.sh` | 6061/6062 | 创建 EICAR 标准测试文件，等待 ClamAV 扫描检出 | ClamAV |

```bash
sudo bash scripts/test-scanner.sh prepare   # 创建测试文件（Agent 启动前执行）
sudo bash scripts/test-scanner.sh cleanup   # 清理测试文件
```

### 数据库清理

| 脚本 | 说明 |
|------|------|
| `clean-test-db.sh` | 集成测试前清空远程数据库所有表 |

```bash
DB_HOST=<REMOTE_IP> DB_USER=<DB_USER> DB_PASS=<DB_PASS> bash scripts/clean-test-db.sh
```

---

## 注意事项

1. **权限**：除 `test-nids.sh` 外，所有脚本需要 `sudo` 运行
2. **依赖缺失**：脚本会在前置检查阶段提示缺少的工具，`test-all-alerts.sh` 会跳过失败项继续执行
3. **白名单**：SSH 暴力破解默认 `127.0.0.1` 在白名单内，本地测试需先移除（参见脚本内提示）
4. **检测延迟**：eBPF 类告警实时触发；Detector 类告警有 1-2 分钟检测周期延迟
5. **Scanner 顺序**：需先执行 `test-scanner.sh prepare` 创建测试文件，再启动 Agent
