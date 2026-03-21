# Collector 插件测试流程

资产采集（collector）插件测试指南，使用 agent standalone 模式进行本地验证。

---

## 一、插件概述

Collector 插件负责采集主机资产信息，包含 12 个 Handler：

| Handler | DataType | 采集间隔 | 说明 |
|---------|----------|---------|------|
| ProcessHandler | 5050 | 1 小时 | 进程信息 |
| PortHandler | 5051 | 1 小时 | 网络端口 |
| KmodHandler | 5062 | 1 小时 | 内核模块 |
| ServiceHandler | 5054 | 6 小时 | 系统服务（systemd） |
| SoftwareHandler | 5055 | 6 小时 | 安装的软件包 |
| UserHandler | 5052 | 6 小时 | 用户账号 |
| EnvSuspiciousHandler | 5057 | 6 小时 | 可疑环境变量 |
| ContainerHandler | 5056 | 6 小时 | Docker/Containerd 容器 |
| ImageHandler | 5058 | 6 小时 | 容器镜像 |
| ImagePackageHandler | 5059 | 6 小时 | 镜像内软件包 |
| DatabaseHandler | 5061 | 6 小时 | 数据库服务（MySQL/PostgreSQL） |
| WebServiceHandler | 5060 | 6 小时 | Web 服务（Nginx/Apache） |

---

## 二、Standalone 模式测试

Standalone 模式无需连接 gRPC Server，插件启动后自动执行所有 Handler，采集结果输出到 stderr 或文件。

### 2.1 启动 agent（所有 Handler）

```bash
cd /home/work/goProject/src/company/agent

# 使用 build 目录测试（推荐日常开发）
sudo ./build/agent -standalone -plugins=collector -output=stderr -test
```

**参数说明：**

| 参数 | 说明 |
|------|------|
| `-standalone` | 启用 standalone 模式，不连接 gRPC Server |
| `-plugins=collector` | 仅加载 collector 插件 |
| `-output=stderr` | 采集结果输出到标准错误（也可指定文件路径，如 `/tmp/collector-output.json`） |
| `-test` | 测试模式，使用固定 agent ID |

### 2.2 输出保存到文件

```bash
# 方式一：output 参数指定文件
sudo ./build/agent -standalone -plugins=collector -output=/tmp/collector-output.json -test

# 方式二：重定向 stderr 到文件
sudo ./build/agent -standalone -plugins=collector -output=stderr -test 2>&1 | tee /tmp/collector-output.log
```

### 2.3 退出

- `Ctrl+C` 发送 SIGINT 信号，等待当前 Handler 执行完成后退出
- 如果 Handler 正在执行中，会等待完成（部分 Handler 采集耗时较长）

---

## 三、各 Handler 预期输出验证

启动 standalone 模式后，所有 Handler 会并发执行。以下是各 Handler 的预期输出和验证要点。

### 3.1 进程采集（DataType 5050）

**预期输出示例：**

```
========== Process Record ==========
PID: 1
Command: /sbin/init
Executable: /usr/lib/systemd/systemd
Working Directory: /
PPID: 0
State: S
User: root (UID: 0)
====================================
```

**验证要点：**
- 能采集到当前系统运行的进程列表
- 字段包含 pid、ppid、cmdline、exe、cwd、uid 等
- 进程数量应与 `ps aux | wc -l` 大致一致

### 3.2 端口采集（DataType 5051）

**预期输出示例：**

```
========== Port Record ==========
Protocol: 6 (TCP)
Family: 2 (IPv4)
Local:  0.0.0.0:22
Remote: 0.0.0.0:0
State: 10 (LISTEN)
UID: 0 (root)
Inode: 12345
=================================
```

**验证要点：**
- 能采集到 TCP/UDP 监听端口
- 与 `ss -tlnp` 输出对比，端口列表应一致
- Protocol 6=TCP, 17=UDP；Family 2=IPv4, 10=IPv6；State 10=LISTEN

### 3.3 用户采集（DataType 5052）

**预期输出示例：**

```
========== User Record ==========
Username: root
UID: 0
GID: 0 (root)
Home: /root
Shell: /bin/bash
Account Type: ROOT
Password Last Change: 19500
Password Max Days: 99999
=================================
```

**验证要点：**
- 能采集到 `/etc/passwd` 中的所有用户
- 包含密码过期信息（来自 `/etc/shadow`）
- 能识别 root 用户和 sudo 用户
- 弱密码检测字段（weak_password）

### 3.4 系统服务采集（DataType 5054）

**预期输出：** 采集 systemd 管理的服务列表

**验证要点：**
- 字段包含 name、type、command、status、run_user、version
- 与 `systemctl list-units --type=service` 对比
- version 字段通过智能提取获取简洁版本号（如 `curl 7.81.0`、`Python 3.11.2`），而非完整的 `--version` 输出

### 3.5 软件包采集（DataType 5055）

**预期输出示例：**

```
========== Software Record ==========
Name: openssh-server
Version: 1:8.9p1-3ubuntu0.6
Type: dpkg (Debian/Ubuntu)
Status: install ok installed
====================================
```

**验证要点：**
- Debian/Ubuntu 系统：采集 dpkg 包列表，与 `dpkg -l | wc -l` 对比
- RedHat/CentOS 系统：采集 rpm 包列表，与 `rpm -qa | wc -l` 对比

### 3.6 容器采集（DataType 5056）

**前提条件：** 需要安装 Docker 或 Containerd，且有运行中的容器

```bash
# 准备测试容器
docker run -d --name test-nginx nginx:latest
```

**预期输出示例：**

```
========== Container Record ==========
Container ID: a1b2c3d4e5f6...
Container Name: test-nginx
State: running
Image Name: nginx:latest
Runtime: docker
=====================================
```

**验证要点：**
- 与 `docker ps -a` 对比，容器列表应一致
- 未安装 Docker/Containerd 时，此 Handler 不会输出记录（正常行为）

**清理：**

```bash
docker rm -f test-nginx
```

### 3.7 可疑环境变量检测（DataType 5057）

**预期输出示例（无可疑项时）：**

```
========== Environment Suspicious Detection Summary ==========
Total Environment Variables: 25
Suspicious Count: 0
No suspicious environment variables found.
==============================================================
```

**验证要点：**
- 扫描当前系统所有进程的环境变量
- 未检测到可疑项时，输出汇总信息（suspicious_count=0）
- 检测到可疑项时，输出具体的变量名、值和可疑原因

### 3.8 数据库服务采集（DataType 5061）

**前提条件：** 需要安装 MySQL 或 PostgreSQL

**验证要点：**
- 检测本机运行的数据库进程
- 字段包含 type（mysql/postgresql）、port、version
- 未安装数据库时，此 Handler 不会输出记录

### 3.9 Web 服务采集（DataType 5060）

**前提条件：** 需要安装 Nginx 或 Apache

**验证要点：**
- 检测本机运行的 Web 服务进程
- 字段包含 type（nginx/apache）、port、config_path、version
- 未安装 Web 服务时，此 Handler 不会输出记录

### 3.10 内核模块采集（DataType 5062）

**预期输出示例：**

```
========== Kernel Module Record ==========
Name: ip_tables
Size: 32768 bytes
RefCount: 3
Used By: iptable_filter,iptable_nat
State: Live
Address: 0xffffffffc0a00000
==========================================
```

**验证要点：**
- 与 `lsmod | wc -l` 对比，模块数量应一致
- 字段包含 name、size、refcount、used_by、state、addr

### 3.11 容器镜像采集（DataType 5058）

**前提条件：** 需要安装 Docker 或 Containerd

**验证要点：**
- 与 `docker images` 对比，镜像列表应一致
- 字段包含 image_id、image_name、image_version、image_size

### 3.12 镜像软件包采集（DataType 5059）

**前提条件：** 需要有**运行中的容器**（Handler 通过 `docker exec` 进入容器执行包管理命令，仅处理 `State=running` 的容器）

```bash
# 准备测试容器
docker run -d --name test-nginx nginx:alpine
```

**验证要点：**
- 进入运行中容器，通过 dpkg/rpm/apk 采集已安装软件包
- 按 ImageID 去重，每个镜像只采集一次
- 字段包含 image_id、image_name、package_name、package_version、package_type、os_version
- 无运行中容器时，此 Handler 不会输出记录（正常行为）

---

## 四、E2E 自动化测试

E2E 测试通过测试程序模拟 Server 下发任务，验证 collector 插件的完整采集流程。

### 4.1 执行 E2E 测试

```bash
cd /home/work/goProject/src/company/agent

# 方式一：使用 Makefile（推荐）
make test-e2e-collector

# 方式二：直接执行脚本
cd tests/e2e/collector && ./test.sh
```

### 4.2 E2E 测试流程说明

测试程序会自动执行以下步骤：

1. 编译 collector 插件
2. 启动 plugin daemon 和 transport daemon
3. 加载 collector 插件
4. 按顺序发送各类采集任务（每个任务间隔 2 秒）：
   - 5050（进程） → 5051（端口） → 5052（用户） → 5054（服务）
   - → 5055（软件） → 5056（容器） → 5057（环境变量）
   - → 5061（数据库） → 5060（Web 服务）
5. 持续读取并打印采集结果（每 500ms 轮询一次）
6. 180 秒后自动退出

### 4.3 预期结果

测试程序会格式化输出每条采集记录，同时输出任务状态响应（DataType 5100）：

```
========== Task Status ==========
Status: succeed
Token: test-process-token-1234567890
Message:
================================
```

每个 Handler 执行完成后都会返回一条 `status: succeed` 的任务状态。

### 4.4 可选：启用 JSON 文件输出

E2E 测试程序支持将采集结果写入 JSON 文件，需要修改测试代码中的开关：

```go
// tests/e2e/collector/main.go
enableJSONOutput = true                       // 改为 true
jsonOutputFile = "collector_records.json"      // 输出文件路径
```

修改后重新执行测试，结果会同时写入 `collector_records.json` 文件。

---

## 五、常见问题排查

### 5.1 权限不足

```
Error: operation not permitted
```

**解决：** 使用 `sudo` 运行，collector 需要 root 权限读取 `/proc`、`/etc/shadow` 等文件。

### 5.2 插件未找到

```
plugin not found: collector
```

**解决：** 检查 plugins 目录是否正确：

```bash
# build 目录模式
ls -la build/plugins/collector/collector

# deploy 目录模式
ls -la /opt/cloudsec/agent/plugins/collector/collector
```

### 5.3 容器/镜像相关 Handler 无输出

这是正常行为。未安装 Docker/Containerd 时，ContainerHandler、ImageHandler、ImagePackageHandler 不会产生采集记录。

### 5.4 数据库/Web 服务 Handler 无输出

同上，未安装对应服务（MySQL/PostgreSQL/Nginx/Apache）时不会输出记录。

### 5.5 查看 collector 插件日志

```bash
# standalone 模式下，插件日志在 agent 工作目录下
# 默认位置取决于启动方式：

# build 目录启动
ls -la /tmp/cloudsec-agent/plugins/collector/

# deploy 目录启动
cat /opt/cloudsec/agent/logs/plugins/collector/collector.log
```
