# Collector 插件资产采集汇总

## 一、采集器总览

| 采集器 | DataType | 运行周期 | 数据来源 |
|--------|----------|----------|----------|
| Process（进程） | 5050 | 每小时 | /proc 目录 |
| Port（监听端口） | 5051 | 每小时 | /proc/net/tcp, udp |
| Kmod（内核模块） | 5062 | 每小时 | /proc/modules |
| Service（系统服务） | 5054 | 每6小时 | systemd 配置文件 |
| Software（软件包） | 5055 | 每6小时 | dpkg, rpm, PyPI, JAR |
| User（用户账号） | 5052 | 每6小时 | /etc/passwd, shadow, wtmp |
| EnvSuspicious（可疑环境变量） | 5056 | 每6小时 | /etc/environment, profile |
| App（应用服务） | 5060 | - | 运行进程 |

---

## 二、各资产字段详情

### 1. 进程（Process）- DataType: 5050

| 字段名 | 类型 | 来源 | 含义 |
|--------|------|------|------|
| `pid` | int | /proc | 进程ID |
| `ppid` | int | /proc/[pid]/stat | 父进程ID |
| `pgid` | int | /proc/[pid]/stat | 进程组ID |
| `sid` | int | /proc/[pid]/stat | 会话ID |
| `comm` | string | /proc/[pid]/stat | 进程名称 |
| `state` | string | /proc/[pid]/stat | 进程状态（R/S/D等） |
| `cmdline` | string | /proc/[pid]/cmdline | 完整命令行参数 |
| `cwd` | string | /proc/[pid]/cwd | 当前工作目录 |
| `exe` | string | /proc/[pid]/exe | 可执行文件路径 |
| `exe_hash` | string | /proc/[pid]/exe | 可执行文件SHA256哈希 |
| `checksum` | string | /proc/[pid]/exe | 可执行文件MD5校验和 |
| `start_time` | int64 | /proc/[pid]/stat | 进程启动时间（Unix时间戳） |
| `ruid` | int | /proc/[pid]/status | 真实用户ID |
| `euid` | int | /proc/[pid]/status | 有效用户ID |
| `suid` | int | /proc/[pid]/status | 保存的用户ID |
| `fsuid` | int | /proc/[pid]/status | 文件系统用户ID |
| `rgid` | int | /proc/[pid]/status | 真实组ID |
| `egid` | int | /proc/[pid]/status | 有效组ID |
| `sgid` | int | /proc/[pid]/status | 保存的组ID |
| `fsgid` | int | /proc/[pid]/status | 文件系统组ID |
| `rusername` | string | 计算 | 真实用户名 |
| `eusername` | string | 计算 | 有效用户名 |
| `susername` | string | 计算 | 保存的用户名 |
| `fsusername` | string | 计算 | 文件系统用户名 |
| `umask` | string | /proc/[pid]/status | 文件掩码 |
| `nspid` | string | /proc/[pid]/status | 命名空间PID |
| `nspgid` | string | /proc/[pid]/status | 命名空间进程组ID |
| `nssid` | string | /proc/[pid]/status | 命名空间会话ID |
| `cns` | uint64 | /proc/[pid]/ns | Cgroup命名空间ID |
| `ins` | uint64 | /proc/[pid]/ns | IPC命名空间ID |
| `mns` | uint64 | /proc/[pid]/ns | Mount命名空间ID |
| `nns` | uint64 | /proc/[pid]/ns | Network命名空间ID |
| `pns` | uint64 | /proc/[pid]/ns | PID命名空间ID |
| `uns` | uint64 | /proc/[pid]/ns | User命名空间ID |
| `utns` | uint64 | /proc/[pid]/ns | UTS命名空间ID |
| `container_id` | string | 缓存 | 容器ID |
| `container_name` | string | 缓存 | 容器名称 |
| `integrity` | string | 计算 | 文件完整性状态 |
| `package_seq` | int64 | 自动 | 采集批次序列号 |

---

### 2. 监听端口（Port）- DataType: 5051

| 字段名 | 类型 | 含义 |
|--------|------|------|
| `family` | int | IP地址族（2=IPv4, 10=IPv6） |
| `protocol` | int | 协议（6=TCP, 17=UDP） |
| `state` | int | 端口状态（TCP: 10=LISTEN, UDP: 7） |
| `sport` | int | 源端口（本地监听端口） |
| `dport` | int | 目标端口（通常为0） |
| `sip` | string | 源IP地址（本地IP） |
| `dip` | string | 目标IP地址 |
| `uid` | int | 用户ID |
| `inode` | uint64 | Socket inode号 |
| `username` | string | 用户名 |
| `pid` | int | 绑定该端口的进程ID |
| `exe` | string | 进程可执行文件路径 |
| `comm` | string | 进程名称 |
| `cmdline` | string | 进程命令行 |
| `psm` | string | PSM标识（K8s环境变量） |
| `pod_name` | string | Pod名称 |
| `package_seq` | int64 | 采集批次序列号 |

---

### 3. 用户账号（User）- DataType: 5052

| 字段名 | 类型 | 来源 | 含义 |
|--------|------|------|------|
| `username` | string | /etc/passwd | 用户名 |
| `password` | string | /etc/passwd | 密码字段（通常为x） |
| `uid` | int | /etc/passwd | 用户ID |
| `gid` | int | /etc/passwd | 组ID |
| `groupname` | string | 计算 | 组名 |
| `info` | string | /etc/passwd | 用户备注信息 |
| `home` | string | /etc/passwd | 用户主目录 |
| `shell` | string | /etc/passwd | 登录shell |
| `last_login_time` | int64 | /var/log/wtmp | 最后登录时间（Unix时间戳） |
| `last_login_ip` | string | /var/log/wtmp | 最后登录IP地址 |
| `is_root` | bool | 计算 | 是否为root账号（uid==0） |
| `is_sudo` | bool | sudo -l | 是否有sudo权限 |
| `is_expired` | bool | /etc/shadow | 密码是否已过期 |
| `is_expiring_soon` | bool | /etc/shadow | 密码是否即将过期 |
| `password_last_change` | int64 | /etc/shadow | 密码最后修改日期 |
| `password_max_days` | int | /etc/shadow | 密码最大使用天数 |
| `password_warn_days` | int | /etc/shadow | 密码过期前警告天数 |
| `password_expire_date` | int64 | /etc/shadow | 密码过期日期 |
| `password_remain_days` | int | /etc/shadow | 密码剩余有效天数 |
| `sudoers` | string | sudo -l | sudo权限内容 |
| `package_seq` | int64 | 自动 | 采集批次序列号 |

---

### 4. 系统服务（Service）- DataType: 5054

| 字段名 | 类型 | 来源 | 含义 |
|--------|------|------|------|
| `name` | string | 文件名 | 服务名称（如nginx.service） |
| `type` | string | Type字段 | 服务类型（simple/oneshot/dbus等） |
| `command` | string | ExecStart字段 | 启动命令 |
| `restart` | string | Restart字段 | 是否自动重启（true/false） |
| `working_dir` | string | WorkingDirectory | 工作目录 |
| `checksum` | string | 文件MD5 | 服务文件MD5校验和 |
| `bus_name` | string | D-Bus | D-Bus总线名称 |
| `package_seq` | int64 | 自动 | 采集批次序列号 |

---

### 5. 软件包（Software）- DataType: 5055

#### dpkg包（Debian/Ubuntu）

| 字段名 | 类型 | 含义 |
|--------|------|------|
| `type` | string | 固定值 "dpkg" |
| `name` | string | 包名称 |
| `sversion` | string | 版本号 |
| `source` | string | 源包名称 |
| `status` | string | 包状态 |
| `package_seq` | int64 | 采集批次序列号 |

#### rpm包（RedHat/CentOS）

| 字段名 | 类型 | 含义 |
|--------|------|------|
| `type` | string | 固定值 "rpm" |
| `name` | string | 包名称 |
| `sversion` | string | 版本号 |
| `source_rpm` | string | 源RPM包 |
| `vendor` | string | 厂商 |
| `package_seq` | int64 | 采集批次序列号 |

#### PyPI包（Python）

| 字段名 | 类型 | 含义 |
|--------|------|------|
| `type` | string | 固定值 "pypi" |
| `name` | string | 包名称 |
| `sversion` | string | 版本号 |
| `component_version` | string | 组件版本 |
| `package_seq` | int64 | 采集批次序列号 |

#### JAR包（Java）

| 字段名 | 类型 | 含义 |
|--------|------|------|
| `type` | string | 固定值 "jar" |
| `name` | string | JAR包名称 |
| `sversion` | string | 版本号 |
| `path` | string | JAR文件路径 |
| `pid` | int | Java进程ID |
| `cmdline` | string | 进程命令行 |
| `psm` | string | PSM标识 |
| `pod_name` | string | Pod名称 |
| `container_id` | string | 容器ID |
| `container_name` | string | 容器名称 |
| `package_seq` | int64 | 采集批次序列号 |

---

### 6. 应用服务（App）- DataType: 5060

#### 支持的应用

| 应用名 | 进程名 | 类型 |
|--------|--------|------|
| Apache | apache2, httpd | web_service |
| Nginx | nginx | web_service |
| Redis | redis-server | database |
| MySQL | mysqld | database |
| PostgreSQL | postgres | database |
| MongoDB | mongod | database |

#### 字段列表

| 字段名 | 类型 | 含义 |
|--------|------|------|
| `name` | string | 应用名称 |
| `type` | string | 应用类型（web_service/database） |
| `sversion` | string | 版本号 |
| `conf` | string | 配置文件路径 |
| `pid` | int | 进程ID |
| `exe` | string | 可执行文件路径 |
| `start_time` | int64 | 启动时间（Unix时间戳） |
| `package_seq` | int64 | 采集批次序列号 |

---

### 7. 可疑环境变量（EnvSuspicious）- DataType: 5056

| 字段名 | 类型 | 含义 |
|--------|------|------|
| `var_name` | string | 环境变量名 |
| `var_value` | string | 环境变量值 |
| `suspicious_reasons` | string | 可疑原因（分号分隔） |
| `source` | string | 来源（system） |
| `package_seq` | int64 | 采集批次序列号 |

#### 检测的可疑变量名

LD_PRELOAD, LD_LIBRARY_PATH, PROMPT_COMMAND, PS1, PATH, HISTFILE, HISTCONTROL, HTTP_PROXY, HTTPS_PROXY, NO_PROXY, TMPDIR, TMP, TEMP

---

### 8. 内核模块（Kmod）- DataType: 5062

| 字段名 | 类型 | 来源 | 含义 |
|--------|------|------|------|
| `name` | string | 字段1 | 模块名称 |
| `size` | string | 字段2 | 模块大小（字节） |
| `refcount` | string | 字段3 | 引用计数 |
| `used_by` | string | 字段4 | 依赖该模块的模块列表 |
| `state` | string | 字段5 | 状态（Live/Loading/Unloading） |
| `addr` | string | 字段6 | 模块内存地址 |
| `package_seq` | int64 | 自动 | 采集批次序列号 |
