# DataType 详细说明

本文档描述 Agent 各插件使用的 DataType 定义和数据字段。

---

## 一、DataType 概览

| 范围 | 插件 | 用途 |
|------|------|------|
| 5050-5062 | collector | 资产采集数据 |
| 5100 | collector | 采集状态 |
| 59-65 | ebpf_base_detector | eBPF 事件（进程/网络/DNS/文件/挂载） |
| 6001-6009 | 多插件共用 | 安全告警（暴力破解/高危命令/反弹Shell/提权/NIDS/恶意请求/敏感文件） |
| 6010-6011 | detector | 检测器状态 |
| 6050-6061 | scanner | 病毒扫描（库更新/目录扫描/全盘扫描/检出结果） |
| 6007 | nids | 网络入侵检测告警 |
| 7001-7004 | ebpf_base_detector | 容器安全告警（危险命令/逃逸/反弹Shell/敏感文件） |
| 8000-8010 | baseline | 基线检查结果 |

---

## 二、Collector 插件 DataType

### 5050 - 进程信息

**Handler:** ProcessHandler

| 字段 | 类型 | 说明 |
|------|------|------|
| pid | string | 进程 ID |
| ppid | string | 父进程 ID |
| comm | string | 进程名（来自 /proc stat） |
| cmdline | string | 完整命令行 |
| path | string | 可执行文件路径 |
| checksum | string | 文件 MD5 |
| exe_hash | string | 可执行文件哈希 |
| cwd | string | 工作目录 |
| container_id | string | 容器 ID（如在容器内） |
| container_name | string | 容器名称 |
| package_seq | string | 采集批次序列号 |

> 注：通过 mapstructure 还会包含 /proc/[pid]/stat 和 status 中的所有字段（uid、pgid 等）。

### 5051 - 端口信息

**Handler:** PortHandler

| 字段 | 类型 | 说明 |
|------|------|------|
| family | string | 地址族 (ipv4/ipv6) |
| protocol | string | 协议 (tcp/udp) |
| state | string | 状态 (LISTEN/ESTABLISHED) |
| sport | string | 源端口 |
| dport | string | 目标端口 |
| sip | string | 源 IP |
| dip | string | 目标 IP |
| uid | string | 用户 ID |
| inode | string | Inode |
| username | string | 用户名 |
| pid | string | 进程 ID |
| exe | string | 可执行文件路径 |
| comm | string | 进程名 |
| cmdline | string | 命令行 |

### 5052 - 用户信息

**Handler:** UserHandler

| 字段 | 类型 | 说明 |
|------|------|------|
| username | string | 用户名 |
| uid | string | 用户 ID |
| gid | string | 组 ID |
| home | string | 家目录 |
| shell | string | 登录 Shell |
| password_expire_date | string | 密码过期日期 |
| password_remain_days | string | 密码剩余有效天数 |
| is_sudo | string | 是否有 sudo 权限 |
| sudoers | string | sudo 权限详情 |
| is_root | string | 是否 root 用户 |
| last_login_time | string | 最后登录时间 |
| last_login_ip | string | 最后登录 IP |
| package_seq | string | 采集批次序列号 |

### 5054 - 系统服务

**Handler:** ServiceHandler

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 服务名 |
| type | string | 服务类型 |
| command | string | 启动命令 |
| restart | string | 重启策略 |
| run_user | string | 运行用户 |
| status | string | 状态 (running/stopped) |
| version | string | 版本 |
| path | string | 服务文件路径 |
| working_dir | string | 工作目录 |
| checksum | string | 校验和 |
| package_seq | string | 采集批次序列号 |

### 5055 - 软件包信息

**Handler:** SoftwareHandler

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 软件名 |
| sversion | string | 版本 |
| source | string | 来源 |
| status | string | 状态 |
| vendor | string | 供应商 |
| package_seq | string | 采集批次序列号 |

### 5056 - 容器信息

**Handler:** ContainerHandler

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 容器 ID |
| name | string | 容器名 |
| state | string | 状态 (running/stopped) |
| image_id | string | 镜像 ID |
| image_name | string | 镜像名 |
| pid | string | 主进程 PID |
| pns | string | PID 命名空间 |
| runtime | string | 运行时 (docker/containerd) |
| create_time | string | 创建时间 |
| package_seq | string | 采集批次序列号 |

### 5057 - 环境变量可疑项

**Handler:** EnvSuspiciousHandler

| 字段 | 类型 | 说明 |
|------|------|------|
| var_name | string | 环境变量名 |
| var_value | string | 环境变量值 |
| source | string | 来源进程 |
| suspicious_reasons | string | 可疑原因 |
| total_envs | string | 总环境变量数 |
| suspicious_count | string | 可疑数量 |
| package_seq | string | 采集批次序列号 |

### 5058 - 镜像资产

**Handler:** ImageHandler

| 字段 | 类型 | 说明 |
|------|------|------|
| image_id | string | 镜像 ID |
| image_name | string | 镜像名称 |
| image_version | string | 镜像版本/标签 |
| image_size | string | 镜像大小（如 134MB） |
| container_count | string | 关联容器数 |
| image_build_time | string | 镜像构建时间 |
| runtime | string | 运行时 (docker/containerd) |

### 5059 - 镜像软件包

**Handler:** ImagePackageHandler

| 字段 | 类型 | 说明 |
|------|------|------|
| image_id | string | 镜像 ID |
| image_name | string | 镜像名称 |
| container_id | string | 采集时使用的容器 ID |
| package_name | string | 软件包名称 |
| package_version | string | 软件包版本 |
| package_type | string | 包管理器类型 (dpkg/rpm/apk) |
| os_version | string | 容器内 OS 版本 |
| package_seq | string | 采集批次序列号 |

### 5060 - Web 服务

**Handler:** WebServiceHandler

| 字段 | 类型 | 说明 |
|------|------|------|
| app_name | string | 应用名 |
| server_type | string | 服务类型 (nginx/apache) |
| path | string | 配置文件路径 |
| version | string | 版本 |
| run_user | string | 运行用户 |
| site_domain | string | 站点域名 |
| package_seq | string | 采集批次序列号 |

### 5061 - 数据库服务

**Handler:** DatabaseHandler

| 字段 | 类型 | 说明 |
|------|------|------|
| db_type | string | 数据库类型 (mysql/postgresql) |
| port | string | 监听端口 |
| db_version | string | 版本 |
| run_user | string | 运行用户 |
| package_seq | string | 采集批次序列号 |

### 5062 - 内核模块

**Handler:** KmodHandler

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 模块名 |
| size | string | 大小 |
| refcount | string | 引用计数 |
| used_by | string | 依赖模块 |
| state | string | 状态 |
| addr | string | 地址 |
| package_seq | string | 采集批次序列号 |

### 5100 - 采集任务状态

| 字段 | 类型 | 说明 |
|------|------|------|
| token | string | 任务令牌 |
| status | string | 状态 (succeed/failed) |
| msg | string | 状态消息 |

---

## 三、Detector 插件 DataType

### 6001 - SSH 暴力破解告警

| 字段 | 类型 | 说明 |
|------|------|------|
| alert_type | string | 告警类型 |
| service | string | 服务名 (ssh) |
| rule_name | string | 触发规则名 |
| description | string | 规则描述 |
| source_ip | string | 攻击源 IP |
| target_user | string | 目标用户 |
| count | string | 失败次数 |
| timeframe | string | 统计时间窗口(秒) |
| first_seen | string | 首次检测时间 |
| last_seen | string | 最后检测时间 |
| level | string | 告警级别 |
| result | string | 检测结果 |
| abnormal_type | string | 异常类型 |

### 6002 - FTP 暴力破解告警

字段同 6001，`service` 字段值为 `ftp`。

### 6005 - SSH 异常登录告警

字段同 6001，通过同一 `sendAlert()` 函数发送。

### 6010 - 检测器配置更新（仅入站任务 DataType）

> 注：6010 不是插件发送的数据类型，而是 Server 下发配置更新任务时使用的 DataType。

### 6011 - 检测器任务状态

| 字段 | 类型 | 说明 |
|------|------|------|
| token | string | 任务令牌 |
| status | string | 状态 (succeed/failed) |
| msg | string | 状态消息 |

---

## 四、安全告警 DataType（多插件共用）

以下告警类型在 `business_plugins/lib/alerttype.go` 中统一定义，由不同插件产生。

### 6003 - 高危命令告警

**产生插件：** ebpf_base_detector

| 字段 | 类型 | 说明 |
|------|------|------|
| rule_id | string | 匹配的规则 ID |
| rule_name | string | 规则名称 |
| severity | string | 严重级别 (critical/high/medium/low) |
| command | string | 触发的完整命令 |
| matched_pattern | string | 匹配的模式 |
| pid | string | 进程 ID |
| uid | string | 用户 ID |
| exe_path | string | 可执行文件路径 |

### 6004 - 反弹 Shell 告警

**产生插件：** ebpf_base_detector

| 字段 | 类型 | 说明 |
|------|------|------|
| pid | string | 进程 ID |
| tgid | string | 线程组 ID |
| ppid | string | 父进程 ID |
| uid | string | 用户 ID |
| comm | string | 进程名 |
| exe_path | string | 可执行文件路径 |
| args | string | 命令行参数 |
| fd_type | string | 文件描述符类型 |
| stdin_path | string | stdin 路径 |
| stdout_path | string | stdout 路径 |
| remote_ip | string | 远程 IP |
| remote_port | string | 远程端口 |
| rule_name | string | 规则名称 |
| confidence | string | 置信度 |
| description | string | 描述 |

### 6006 - 本地提权告警

**产生插件：** ebpf_base_detector

| 字段 | 类型 | 说明 |
|------|------|------|
| pid | string | 进程 ID |
| tgid | string | 线程组 ID |
| ppid | string | 父进程 ID |
| uid | string | 用户 ID |
| comm | string | 进程名 |
| exe_path | string | 可执行文件路径 |
| old_uid | string | 提权前 UID |
| old_euid | string | 提权前 EUID |
| new_uid | string | 提权后 UID |
| new_euid | string | 提权后 EUID |

### 6007 - NIDS 网络攻击告警

**产生插件：** nids

| 字段 | 类型 | 说明 |
|------|------|------|
| src_ip | string | 攻击来源 IP |
| dst_ip | string | 目标 IP |
| src_port | string | 来源端口 |
| dst_port | string | 目标端口 |
| vulnerability_name | string | 漏洞/攻击名称 |
| attack_status | string | 攻击分类 |
| severity | string | 严重级别 |
| sid | string | 规则 SID |
| reference | string | 参考链接 |
| attack_count | string | 攻击��数 |
| last_attack_time | string | 最后攻击时间 |
| first_attack_time | string | 首次攻击时间 |
| matched_payload | string | 匹配的载荷片段 |
| http_method | string | HTTP 方法 |
| http_uri | string | HTTP URI |

### 6008 - 恶意请求告警

**产生插件：** ebpf_base_detector（威胁情报匹配）

| 字段 | 类型 | 说明 |
|------|------|------|
| event_type | string | 事件类型 (connect/dns) |
| rule_id | string | 规则 ID |
| rule_name | string | 规则名称 |
| severity | string | 严重级别 |
| threat_type | string | 威胁类型 |
| indicator_type | string | 指标类型 (ip/domain) |
| matched_value | string | 匹配的值 |
| pid | string | 进程 ID |
| comm | string | 进程名 |
| exe_path | string | 可执行文件路径 |
| remote_ip | string | 远程 IP |
| remote_port | string | 远程端口 |
| domain | string | 域名 |

### 6009 - 敏感文件监控告警

**产生插件：** ebpf_base_detector（文件完整性监控）

| 字段 | 类型 | 说明 |
|------|------|------|
| action | string | 操作类型 (create/rename/delete) |
| new_path | string | 文件路径 |
| old_path | string | 原路径（仅 rename） |
| pid | string | 进程 ID |
| uid | string | 用户 ID |
| comm | string | 进程名 |
| exe_path | string | 可执行文件路径 |

---

## 五、Baseline 插件 DataType

### 8000 - 基线检查结果

| 字段 | 类型 | 说明 |
|------|------|------|
| data | string | 检查结果 JSON |
| token | string | 任务令牌 |
| template_name | string | 模板名称 |
| template_id | string | 模板 ID |
| baseline_id | string | 基线 ID |

### 8010 - 基线任务状态

| 字段 | 类型 | 说明 |
|------|------|------|
| token | string | 任务令牌 |
| status | string | 状态 (succeed/failed) |
| msg | string | 状态消息 |
| baseline_id | string | 基线 ID |

---

## 六、ebpf_base_detector 插件 DataType

### 59 - eBPF 进程执行事件 (Execve)

| 字段 | 类型 | 说明 |
|------|------|------|
| pid | string | 进程 ID（线程 ID） |
| tgid | string | 线程组 ID（进程 ID） |
| ppid | string | 父进程 ID |
| pgid | string | 进程组 ID |
| uid | string | 用户 ID |
| comm | string | 命令名（最多 16 字符） |
| exe_path | string | 可执行文件完整路径 |
| args | string | ���令行参数 |
| stdin_path | string | FD 0 的文件路径 |
| stdout_path | string | FD 1 的文件路径 |
| tty_name | string | 控制终端名称 |
| socket_pid | string | 持有 socket 的进程 PID |
| fd_type | string | 内核预过滤标记 (0=无, 1=stdin是socket, 2=stdout, 3=both) |
| remote_ip | string | socket 远程 IP（仅当存在 socket 连接时） |
| remote_port | string | socket 远程端口 |
| local_ip | string | socket 本地 IP |
| local_port | string | socket 本地端口 |

**说明：** 当命令匹配高危规则时，告警以 DataType 6003 发送，附带 `rule_id`、`rule_name`、`severity` 等字段。

### 60 - Connect 出站连接事件

| 字段 | 类型 | 说明 |
|------|------|------|
| pid | string | 进程 ID |
| tgid | string | 线程组 ID |
| ppid | string | 父进程 ID |
| uid | string | 用户 ID |
| comm | string | 进程名 |
| exe_path | string | 可执行文件路径 |
| protocol | string | 协议 (tcp/udp) |
| remote_ip | string | 目标 IP |
| remote_port | string | 目标端口 |
| local_ip | string | 本地 IP |
| local_port | string | 本地端口 |
| retval | string | 系统调用返回值 |

### 61 - Bind 端口绑定事件

| 字段 | 类型 | 说明 |
|------|------|------|
| pid | string | 进程 ID |
| tgid | string | 线程组 ID |
| ppid | string | 父进程 ID |
| uid | string | 用户 ID |
| comm | string | 进程名 |
| exe_path | string | 可执行文件路径 |
| protocol | string | 协议 (tcp/udp) |
| bind_ip | string | 绑定 IP |
| bind_port | string | 绑定端口 |
| retval | string | 系统调用返回值 |

### 62 - Accept 入站连接事件

| 字段 | 类型 | 说明 |
|------|------|------|
| pid | string | 进程 ID |
| tgid | string | 线程组 ID |
| ppid | string | 父进程 ID |
| uid | string | 用户 ID |
| comm | string | 进程名 |
| exe_path | string | 可执行文件路径 |
| protocol | string | 协议 (tcp/udp) |
| remote_ip | string | 连接来源 IP |
| remote_port | string | 连接来源端口 |
| local_ip | string | 本地 IP |
| local_port | string | 本地监听端口 |
| retval | string | 系统调用返回值 |

### 63 - DNS 查询事件

| 字段 | 类型 | 说明 |
|------|------|------|
| pid | string | 进程 ID |
| tgid | string | 线程组 ID |
| ppid | string | 父进程 ID |
| uid | string | 用户 ID |
| comm | string | 进程名 |
| exe_path | string | 可执行文件路径 |
| domain | string | 查询域名 |
| query_type | string | 查询类型 (A/AAAA/CNAME/MX/TXT 或数字) |
| dns_server_ip | string | DNS 服务器 IP |
| dns_server_port | string | DNS 服务器端口 |
| opcode | string | DNS opcode |
| rcode | string | DNS rcode |

### 64 - 文件操作事件

| 字段 | 类型 | 说明 |
|------|------|------|
| pid | string | 进程 ID |
| tgid | string | 线程组 ID |
| ppid | string | 父进程 ID |
| uid | string | 用户 ID |
| comm | string | 进程名 |
| exe_path | string | 可执行文件路径 |
| action | string | 操作类型 (create/rename/delete) |
| new_path | string | 文件路径（创建/重命名后） |
| old_path | string | 原路径（仅 rename） |
| s_id | string | 文件系统 ID |
| socket_pid | string | 持有 socket 的进程 PID（仅当存在时） |
| remote_ip | string | socket 远程 IP（仅当存在 socket 时） |
| remote_port | string | socket 远程端口 |
| local_ip | string | socket 本地 IP |
| local_port | string | socket 本地端口 |

### 65 - Mount 挂载事件

| 字段 | 类型 | 说明 |
|------|------|------|
| pid | string | 进程 ID（线程 ID） |
| tgid | string | 线程组 ID（进程 ID） |
| ppid | string | 父进程 ID |
| uid | string | 用户 ID |
| comm | string | 进程名 |
| exe_path | string | 可执行文件路径 |
| dev_name | string | 挂载源设备 |
| dir_name | string | 挂载目标路径 |
| fs_type | string | 文件系统类型 |
| flags | string | 挂载标志位 |
| retval | string | 系统调用返回值 |
| mntns_id | string | 挂载命名空间 ID |
| root_mntns_id | string | 根挂载命名空间 ID |
| is_container | string | 是否在容器内（mntns_id != root_mntns_id） |

**说明：** 当容器内进程挂载宿主机块设备时，触发容器逃逸告警（DataType 7002）。

### 66 - Perf 事件丢失报告

| 字段 | 类型 | 说明 |
|------|------|------|
| lost_count | string | 上报周期内丢失的 perf 事件总数 |
| report_interval | string | 上报周期（秒），固定值 "30" |

**说明：** 当 eBPF perf buffer 溢出导致事件丢失时，Agent 每 30 秒汇总丢失计数并上报。Server 端以 WARN 日志记录，用于监控高负载场景下的事件丢失情况。

---

## 七、容器安全告警 DataType

所有容器告警（7001-7004）均包含以下公共容器元数据字段（可选，仅当能识别容器时填充）：

| 字段 | 类型 | 说明 |
|------|------|------|
| container_id | string | 完整容器 ID |
| container_id_short | string | 容器 ID 短格式（前 12 位） |
| container_name | string | 容器名称 |
| image_name | string | 容器镜像名称 |

### 7001 - 容器高危命令告警

**检测类型：** `container_dangerous_command`
**数据来源：** ExecveEvent（进程执行事件）

| 字段 | 类型 | 说明 |
|------|------|------|
| pid | string | 进程 ID |
| tgid | string | 线程组 ID |
| ppid | string | 父进程 ID |
| pgid | string | 进程组 ID |
| uid | string | 用户 ID |
| comm | string | 进程名 |
| exe_path | string | 可执行文件路径 |
| args | string | 命令行参数 |
| command | string | 完整命令行 |
| pid_tree | string | 进程树 |
| detection_type | string | 固定值 `container_dangerous_command` |
| rule_id | string | 匹配的规则 ID |
| rule_name | string | 规则名称 |
| severity | string | 严重级别 |
| rule_description | string | 规则描述 |
| matched_pattern | string | 匹配的模式 |
| timestamp | string | 时间戳 |

### 7002 - 容器逃逸告警

**检测类型：** `container_escape`
**数据来源：** MountEvent（挂载事件）
**触发条件：** 容器内进程挂载宿主机块设备（/dev/sd*, /dev/vd*, /dev/nvme*, /dev/xvd*, /dev/hd*）

| 字段 | 类型 | 说明 |
|------|------|------|
| pid | string | 进程 ID |
| tgid | string | 线程组 ID |
| ppid | string | 父进程 ID |
| uid | string | 用户 ID |
| comm | string | 进程名 |
| exe_path | string | 可执行文件路径 |
| dev_name | string | 挂载源设备 |
| dir_name | string | 挂载目标路径 |
| fs_type | string | 文件系统类型 |
| flags | string | 挂载标志位 |
| retval | string | 系统调用返回值 |
| mntns_id | string | 挂载命名空间 ID |
| root_mntns_id | string | 根挂载命名空间 ID |
| is_container | string | 是否在容器内 |
| detection_type | string | 固定值 `container_escape` |
| rule_name | string | 规则名称（如 `container_escape_mount_device`） |
| severity | string | 严重级别（如 `critical`） |
| rule_description | string | 规则描述 |
| pid_tree | string | 进程树 |
| timestamp | string | 时间戳 |

### 7003 - 容器反弹 Shell 告警

**检测类型：** `container_reverse_shell`
**数据来源：** ExecveEvent（进程执行事件）
**触发条件：** 容器内进程的 stdin 或 stdout 指向网络套接字

| 字段 | 类型 | 说明 |
|------|------|------|
| pid | string | 进程 ID |
| tgid | string | 线程组 ID |
| ppid | string | 父进程 ID |
| pgid | string | 进程组 ID |
| uid | string | 用户 ID |
| comm | string | 进程名 |
| exe_path | string | 可执行文件路径 |
| args | string | 命令行参数 |
| fd_type | string | FD 类型标记 (1=stdin是socket, 2=stdout, 3=both) |
| stdin_path | string | FD 0 的路径 |
| stdout_path | string | FD 1 的路径 |
| tty_name | string | 控制终端名称 |
| socket_pid | string | 持有 socket 的进程 PID |
| pid_tree | string | 进程树 |
| detection_type | string | 固定值 `container_reverse_shell` |
| rule_name | string | 规则名称（`container_stdin_socket` 或 `container_stdout_socket`） |
| confidence | string | 置信度（如 `high`） |
| description | string | 检测描述 |
| remote_ip | string | 远程 IP（如有） |
| remote_port | string | 远程端口（如有） |
| local_ip | string | 本地 IP（如有） |
| local_port | string | 本地端口（如有） |
| timestamp | string | 时间戳 |

### 7004 - 容器敏感文件告警

**检测类型：** `container_sensitive_file`
**数据来源：** FileEvent（文件操作事件）

| 字段 | 类型 | 说明 |
|------|------|------|
| pid | string | 进程 ID |
| tgid | string | 线程组 ID |
| ppid | string | 父进程 ID |
| uid | string | 用户 ID |
| comm | string | 进程名 |
| exe_path | string | 可执行文件路径 |
| action | string | 操作类型 (create/rename/delete) |
| new_path | string | 文件路径 |
| old_path | string | 原路径（仅 rename） |
| s_id | string | 文件系统 ID |
| detection_type | string | 固定值 `container_sensitive_file` |
| rule_id | string | 匹配的规则 ID |
| rule_name | string | 规则名称 |
| severity | string | 严重级别 |
| rule_description | string | 规则描述 |
| matched_pattern | string | 匹配的模式 |
| pid_tree | string | 进程树 |
| operator_user | string | 操作用户名 |
| operator_process | string | 操作进程名 |
| timestamp | string | 时间戳 |
| socket_pid | string | 持有 socket 的进程 PID（如有） |
| remote_ip | string | 远程 IP（如有） |
| remote_port | string | 远程端口（如有） |
| local_ip | string | 本地 IP（如有） |
| local_port | string | 本地端口（如有） |

---

## 八、Scanner 插件 DataType

| DataType | 用途 | 说明 |
|----------|------|------|
| 6050 | 病毒库更新 | 病毒库更新状态 |
| 6053 | 指定目录扫描 | 指定目录的恶意文件扫描结果 |
| 6057 | 全盘扫描 | 全盘恶意文件扫描结果 |
| 6060 | 扫描任务状态 | 扫描任务执行状态 |
| 6061 | 静态文件检出 | 静态文件恶意检出结果 |

---

## 九、DataType 分配规范

| 范围 | 用途 | 分配状态 |
|------|------|----------|
| 1-999 | 系统保留 | - |
| 59-65 | eBPF 事件 | 已用 |
| 1000-4999 | Agent 内部 | - |
| 5000-5999 | Collector | 部分已用 (5050-5062, 5100) |
| 6000-6099 | 安全告警/检测 | 部分已用 (6001-6011, 6050-6061) |
| 7000-7999 | 容器安全告警 | 部分已用 (7001-7004) |
| 8000-8999 | Baseline | 部分已用 (8000, 8010) |

新插件开发时，请向项目负责人申请 DataType 范围。

---

## 相关文档

- [架构设计文档](architecture.md) — 系统架构概览和插件架构
- [配置文件详解](configuration.md) — 各插件配置项说明
- [插件开发指南](plugin-development.md) — 新插件开发时的 DataType 注册方法
- [gRPC 协议说明](grpc-protocol.md) — DataType 在传输协议中的使用方式
