# DataType 详细说明

本文档描述 Agent 各插件使用的 DataType 定义和数据字段。

---

## 一、DataType 概览

| 范围 | 插件 | 用途 |
|------|------|------|
| 5050-5062 | collector | 资产采集数据 |
| 5100 | collector | 采集状态 |
| 6001-6005 | detector | 威胁检测告警 |
| 6010-6011 | detector | 检测器状态 |
| 8000-8010 | baseline | 基线检查结果 |
| 59 | ebpf_base_detector | eBPF 进程事件 |

---

## 二、Collector 插件 DataType

### 5050 - 进程信息

**Handler:** ProcessHandler

| 字段 | 类型 | 说明 |
|------|------|------|
| pid | string | 进程 ID |
| ppid | string | 父进程 ID |
| name | string | 进程名 |
| cmdline | string | 完整命令行 |
| exe | string | 可执行文件路径 |
| checksum | string | 文件 MD5 |
| uid | string | 用户 ID |
| username | string | 用户名 |
| container_id | string | 容器 ID（如在容器内） |

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

### 5052 - 用户信息

**Handler:** UserHandler

| 字段 | 类型 | 说明 |
|------|------|------|
| username | string | 用户名 |
| uid | string | 用户 ID |
| gid | string | 组 ID |
| home | string | 家目录 |
| shell | string | 登录 Shell |
| password_expire | string | 密码过期时间 |
| sudo | string | 是否有 sudo 权限 |

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

### 5055 - 软件包信息

**Handler:** SoftwareHandler

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 软件名 |
| version | string | 版本 |
| arch | string | 架构 |
| type | string | 包管理器类型 (deb/rpm) |

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
| runtime | string | 运行时 (docker/containerd) |

### 5057 - 内核模块

**Handler:** KmodHandler

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 模块名 |
| size | string | 大小 |
| used_by | string | 依赖模块 |

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
| type | string | 服务类型 (nginx/apache) |
| port | string | 监听端口 |
| config_path | string | 配置文件路径 |
| version | string | 版本 |

### 5061 - 数据库服务

**Handler:** DatabaseHandler

| 字段 | 类型 | 说明 |
|------|------|------|
| type | string | 数据库类型 (mysql/postgresql) |
| port | string | 监听端口 |
| version | string | 版本 |

### 5062 - 环境变量可疑项

**Handler:** EnvSuspiciousHandler

| 字段 | 类型 | 说明 |
|------|------|------|
| key | string | 环境变量名 |
| value | string | 环境变量值 |
| pid | string | 进程 ID |
| reason | string | 可疑原因 |

### 5100 - 采集任务状态

| 字段 | 类型 | 说明 |
|------|------|------|
| token | string | 任务令牌 |
| status | string | 状态 (succeed/failed) |

---

## 三、Detector 插件 DataType

### 6001 - SSH 暴力破解告警

| 字段 | 类型 | 说明 |
|------|------|------|
| source_ip | string | 攻击源 IP |
| count | string | 失败次数 |
| rule | string | 触发规则名 |
| timeframe | string | 统计时间窗口 |
| level | string | 告警级别 |

### 6002 - FTP 暴力破解告警

| 字段 | 类型 | 说明 |
|------|------|------|
| source_ip | string | 攻击源 IP |
| count | string | 失败次数 |
| rule | string | 触发规则名 |
| timeframe | string | 统计时间窗口 |
| level | string | 告警级别 |

### 6005 - SSH 异常登录告警

| 字段 | 类型 | 说明 |
|------|------|------|
| user | string | 登录用户名 |
| source_ip | string | 登录源 IP |
| service | string | 服务类型 (ssh) |
| timestamp | string | 登录时间 |

### 6010 - 检测器配置更新

| 字段 | 类型 | 说明 |
|------|------|------|
| config_type | string | 配置类型 |
| status | string | 更新状态 |

### 6011 - 检测器任务状态

| 字段 | 类型 | 说明 |
|------|------|------|
| token | string | 任务令牌 |
| status | string | 状态 (succeed/failed) |

---

## 四、Baseline 插件 DataType

### 8000 - 基线检查结果

| 字段 | 类型 | 说明 |
|------|------|------|
| baseline_id | string | 基线 ID |
| check_id | string | 检查项 ID |
| result | string | 结果 (PASS/FAIL) |
| message | string | 结果描述 |
| token | string | 任务令牌 |

### 8010 - 基线任务状态

| 字段 | 类型 | 说明 |
|------|------|------|
| token | string | 任务令牌 |
| status | string | 状态 (succeed/failed) |
| baseline_id | string | 基线 ID |

---

## 五、ebpf_base_detector 插件 DataType

### 59 - eBPF 进程执行事件

| 字段 | 类型 | 说明 |
|------|------|------|
| pid | string | 进程 ID |
| ppid | string | 父进程 ID |
| uid | string | 用户 ID |
| comm | string | 命令名（最多 16 字符） |
| exe | string | 可执行文件路径 |
| args | string | 命令行参数 |
| rule_id | string | 匹配的规则 ID（如触发告警） |
| rule_name | string | 规则名称 |
| severity | string | 严重程度 |
| matched_pattern | string | 匹配的模式 |

**说明：** 当命令匹配高危规则时，会附带 `rule_id`、`rule_name`、`severity` 等字段。

---

## 六、DataType 分配规范

| 范围 | 用途 | 分配状态 |
|------|------|----------|
| 1-999 | 系统保留 | - |
| 1000-4999 | Agent 内部 | - |
| 5000-5999 | Collector | 部分已用 |
| 6000-6999 | Detector | 部分已用 |
| 7000-7999 | 预留 | 未分配 |
| 8000-8999 | Baseline | 部分已用 |

新插件开发时，请向项目负责人申请 DataType 范围。
