# Agent事件写入数据库的代码清单

## 概述
梳理 HCIDS Server 端接收 Agent 上报数据、解析为记录、写入 PostgreSQL 数据库的完整代码链路。

---

## 整体数据流

```
Agent (gRPC) → TransferServer.Transfer() → handlePackagedData() → processPayload()
    → Mapper.Map*()（字段映射） → Repository.*()（写入 PostgreSQL）
```

---

## 1. 入口层 — gRPC Handler

**文件**: `internal/grpc/handler/transfer.go`

| 函数 | 作用 |
|------|------|
| `Transfer()` | gRPC 双向流处理，接收 Agent 上报的 PackagedData |
| `handlePackagedData()` | 提取 Records 和 Agent 元信息 |
| `processPayload()` | 按 DataType 路由到各类处理逻辑（5050-8010） |

---

## 2. 数据映射层 — Mapper

### 安全告警映射: `internal/mapper/alert_mapper.go`

| 函数 | DataType | 说明 |
|------|----------|------|
| `MapBruteForceAlert()` | 6001/6002 | SSH/FTP 暴力破解 |
| `MapDangerousCommandAlert()` | 6003 | 高危命令执行 |
| `MapReverseShellAlert()` | 6004 | 反弹 Shell（eBPF 检测） |
| `MapAbnormalLoginAlert()` | 6005 | 异常登录 |
| `MapPrivilegeEscalationAlert()` | 6006 | 本地提权 |
| `MapMaliciousRequestAlert()` | 6008 | 恶意请求 |
| `MapNetworkAttackAlert()` | 6007 | 网络攻击（NIDS） |
| `MapMalwareScanAlert()` | 6061/6062 | 恶意文件/进程扫描 |

### 资产采集映射: `internal/mapper/asset_mapper.go`

| 函数 | DataType | 说明 |
|------|----------|------|
| `MapHost()` | — | 主机元信息 |
| `MapPort()` | 5051 | 监听端口 |
| `MapAccount()` | 5052 | 用户账户 |
| `MapProcess()` | 5050 | 进程信息 |
| `MapSystemService()` | 5054 | 系统服务 |
| `MapSoftware()` | 5055 | 软件包 |
| `MapContainer()` | 5056 | 容器 |
| `MapImage()` | 5058 | 容器镜像 |
| `MapImagePackage()` | 5059 | 镜像内软件包 |
| `MapWebService()` | 5060 | Web 服务 |
| `MapDatabase()` | 5061 | 数据库服务 |
| `MapKmod()` | 5062 | 内核模块 |
| `MapEnvSuspicious()` | 5057 | 可疑环境变量 |

### 其他映射

| 文件 | 函数 | 说明 |
|------|------|------|
| `internal/mapper/execve_mapper.go` | `MapExecve()` | eBPF execve 系统调用事件 (DataType 59) |
| `internal/mapper/connect_mapper.go` | `MapConnect()` | eBPF connect 出站连接事件 (DataType 60) |
| `internal/mapper/dns_mapper.go` | `MapDNS()` | eBPF DNS 查询事件 (DataType 63) |
| `internal/mapper/file_event_mapper.go` | `MapFileEvent()`, `MapFileIntegrityAlert()` | 文件操作事件 (DataType 64) / 文件完整性告警 (6009) |
| `internal/mapper/baseline_mapper.go` | — | 基线检查结果映射 |

---

## 3. 数据库模型层 — Model

### 告警模型: `internal/models/alert/*.go`

各告警类型拆分为独立文件：

| Model | 文件 | 数据库表 | 关键字段 |
|-------|------|---------|---------|
| `BruteForce` | `brute_force.go` | `alert_brute_force` | AgentID, SourceIP, AttackType, Username, AttemptCount |
| `DangerousCommand` | `dangerous_command.go` | `alert_dangerous_command` | AgentID, Command, CommandType, User, PrivilegeLevel |
| `ReverseShell` | `reverse_shell.go` | `alert_reverse_shell` | AgentID, CommandLine, ShellType, TargetHost, TargetPort |
| `AbnormalLogin` | `abnormal_login.go` | `alert_abnormal_login` | AgentID, SourceIP, LoginUser, LoginTime, RiskLevel |
| `PrivilegeEscalation` | `privilege_escalation.go` | `alert_privilege_escalation` | AgentID, EscalatedUser, ParentProcess, ProcessID |
| `MaliciousRequest` | `malicious_request.go` | `alert_malicious_request` | AgentID, MaliciousDomain, MaliciousIP, RequestCount |
| `NetworkAttack` | `network_attack.go` | `alert_network_attack` | AgentID, AttackerIP, VulnerabilityName, AttackCount |
| `MalwareScan` | `malware_scan.go` | `alert_malware_scan` | AgentID, FilePath, ThreatType, MalwareFamily |
| `FileIntegrity` | `file_integrity.go` | `alert_file_integrity` | AgentID, FilePath, Action, ProcessName |

### 资产模型: `internal/models/assets/host/*.go` 和 `internal/models/assets/container/*.go`

| Model | 包路径 | 数据库表 |
|-------|--------|---------|
| `host.Host` | `assets/host` | `asset_host` |
| `host.Port` | `assets/host` | `asset_port` |
| `host.Account` | `assets/host` | `asset_account` |
| `host.Process` | `assets/host` | `asset_process` |
| `host.Database` | `assets/host` | `asset_database` |
| `host.WebService` | `assets/host` | `asset_web_service` |
| `host.SystemService` | `assets/host` | `asset_system_service` |
| `host.Kmod` | `assets/host` | `asset_kmod` |
| `host.Software` | `assets/host` | `asset_software` |
| `host.EnvSuspicious` | `assets/host` | `asset_env_suspicious` |
| `container.Container` | `assets/container` | `asset_container` |
| `container.Image` | `assets/container` | `asset_image` |

### 事件模型: `internal/models/assets/host/execve.go`

| Model | 数据库表 |
|-------|---------|
| `host.Execve` | `event_execve` |

### 基线模型: `internal/model/baseline.go`

| Model | 说明 |
|-------|------|
| `BaselineCheckResult` | 基线检查结果 |

---

## 4. 持久化层 — Repository（写入数据库的核心代码）

### 告警写入: `internal/db/repository/alert_repository.go`

| 方法 | 操作 | 说明 |
|------|------|------|
| `CreateBruteForceAlert()` | INSERT | 暴力破解告警 |
| `CreateDangerousCommandAlert()` | INSERT | 高危命令告警 |
| `CreateReverseShellAlert()` | INSERT | 反弹 Shell 告警 |
| `CreateAbnormalLoginAlert()` | INSERT | 异常登录告警 |
| `CreatePrivilegeEscalationAlert()` | INSERT | 本地提权告警 |
| `CreateOrUpdateMaliciousRequestAlert()` | UPSERT | 恶意请求告警（按 agent_id+domain/IP 聚合） |
| `CreateMalwareScanAlert()` | INSERT | 恶意文件扫描告警 |
| `CreateNetworkAttackAlert()` | INSERT | 网络攻击告警 |
| `CreateFileIntegrityAlert()` | INSERT | 敏感文件监控告警 |

### 资产写入: `internal/db/repository/asset_repository.go`

| 方法 | 操作 | 说明 |
|------|------|------|
| `CreateOrUpdateHost()` | UPSERT (on agent_id) | 主机信息 |
| `CreateOrUpdatePort()` | UPSERT (on agent_id+port+protocol) | 端口信息 |
| `CreateOrUpdateAccount()` | UPSERT (on agent_id+name) | 账户信息 |
| `CreateOrUpdateProcess()` | UPSERT (on agent_id+path) | 进程信息 |
| `CreateOrUpdateDatabase()` | UPSERT (on agent_id+db_type) | 数据库服务 |
| `CreateOrUpdateWebService()` | UPSERT | Web 服务 |
| `CreateOrUpdateSystemService()` | UPSERT (on agent_id+name) | 系统服务 |
| `CreateOrUpdateContainer()` | UPSERT | 容器 |
| `CreateOrUpdateImage()` | UPSERT | 容器镜像 |
| `CreateOrUpdateImagePackage()` | UPSERT | 镜像软件包 |
| `CreateOrUpdateKmod()` | UPSERT | 内核模块 |
| `CreateOrUpdateSoftware()` | UPSERT | 软件包 |
| `CreateOrUpdateEnvSuspicious()` | UPSERT | 可疑环境变量 |

### 事件写入: `internal/db/repository/execve_repository.go`

| 方法 | 操作 | 说明 |
|------|------|------|
| `Create()` | INSERT（追加写入） | execve 系统调用事件 |
| `DeleteOldRecords()` | DELETE | 清理旧事件记录 |

### 其他事件写入

| 文件 | 方法 | 说明 |
|------|------|------|
| `connect_repository.go` | `Create()` | connect 出站连接事件 (DataType 60) |
| `dns_repository.go` | `Create()` | DNS 查询事件 (DataType 63) |
| `file_event_repository.go` | `Create()` | 文件操作事件 (DataType 64) |
| `agent_info_repository.go` | `CreateOrUpdate()` | Agent 元信息 |
| `vuln_repository.go` | `Create()` | 漏洞扫描结果 |

### 基线写入: `internal/db/repository/baseline_repository.go`

| 方法 | 操作 | 说明 |
|------|------|------|
| `CreateCheckResult()` | INSERT | 基线检查结果 |
| `BatchCreateCheckDetails()` | BATCH INSERT | 基线检查明细 |

---

## 5. 数据库连接: `internal/db/postgres.go`

- 使用 GORM 连接 PostgreSQL
- `Init()` 初始化连接，`GetDB()` 获取全局 DB 实例
- 资产类操作使用 `clause.OnConflict` 实现 UPSERT
- 告警类多数使用 INSERT 直接追加

---

## 6. DataType 速查表

| 范围 | 类别 | 具体类型 |
|------|------|---------|
| **59** | 事件流 | execve 系统调用 |
| **60-64** | eBPF 事件 | Connect, Bind, Accept, DNS, File |
| **5050-5062** | 资产采集 | Process, Port, User, Service, Software, Container, EnvSuspicious, Image, ImagePackage, WebService, Database, Kmod |
| **6001-6009** | 安全告警 | SSH暴破, FTP暴破, 高危命令, 反弹Shell, 异常登录, 本地提权, NIDS, 恶意请求, 敏感文件 |
| **6050-6061** | 恶意扫描 | 库更新, 目录扫描, 全盘扫描, 扫描状态, 文件检出 |
| **8000-8010** | 基线检查 | 检查结果, 任务状态 |

---

## 关键文件汇总

| 层级 | 文件路径 |
|------|---------|
| gRPC 入口 | `internal/grpc/handler/transfer.go` |
| 告警映射 | `internal/mapper/alert_mapper.go` |
| 资产映射 | `internal/mapper/asset_mapper.go` |
| 事件映射 | `internal/mapper/execve_mapper.go`, `connect_mapper.go`, `dns_mapper.go`, `file_event_mapper.go` |
| 基线映射 | `internal/mapper/baseline_mapper.go` |
| 告警模型 | `internal/models/alert/*.go` |
| 资产模型 | `internal/models/assets/host/*.go`, `internal/models/assets/container/*.go` |
| 事件模型 | `internal/models/assets/host/execve.go` |
| 基线模型 | `internal/model/baseline.go` |
| 告警写库 | `internal/db/repository/alert_repository.go` |
| 资产写库 | `internal/db/repository/asset_repository.go` |
| 事件写库 | `internal/db/repository/execve_repository.go`, `connect_repository.go`, `dns_repository.go`, `file_event_repository.go` |
| 基线写库 | `internal/db/repository/baseline_repository.go` |
| Agent 信息 | `internal/db/repository/agent_info_repository.go` |
| 漏洞 | `internal/db/repository/vuln_repository.go` |
| DB 连接 | `internal/db/postgres.go` |
| Protobuf | `proto/grpc.pb.go` |
