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
| `internal/mapper/baseline_mapper.go` | — | 基线检查结果映射 |

---

## 3. 数据库模型层 — Model

### 告警模型: `internal/model/alert.go`

| Model | 数据库表 | 关键字段 |
|-------|---------|---------|
| `AlertBruteForce` | `alert_brute_force` | AgentID, SourceIP, AttackType, Username, AttemptCount |
| `AlertDangerousCommand` | `alert_dangerous_command` | AgentID, Command, CommandType, User, PrivilegeLevel |
| `AlertReverseShell` | `alert_reverse_shell` | AgentID, CommandLine, ShellType, TargetHost, TargetPort |
| `AlertAbnormalLogin` | `alert_abnormal_login` | AgentID, SourceIP, LoginUser, LoginTime, RiskLevel |
| `AlertPrivilegeEscalation` | `alert_privilege_escalation` | AgentID, EscalatedUser, ParentProcess, ProcessID |
| `AlertMaliciousRequest` | `alert_malicious_request` | AgentID, MaliciousDomain, MaliciousIP, RequestCount |
| `AlertNetworkAttack` | `alert_network_attack` | AgentID, AttackerIP, VulnerabilityName, AttackCount |
| `AlertMalwareScan` | `alert_malware_scan` | AgentID, FilePath, ThreatType, MalwareFamily |

### 资产模型: `internal/model/asset.go`

| Model | 数据库表 |
|-------|---------|
| `AssetHost` | `asset_host` |
| `AssetPort` | `asset_port` |
| `AssetAccount` | `asset_account` |
| `AssetProcess` | `asset_process` |
| `AssetDatabase` | `asset_database` |
| `AssetWebService` | `asset_web_service` |
| `AssetSystemService` | `asset_system_service` |
| `AssetContainer` | `asset_container` |
| `AssetImage` | `asset_image` |
| `AssetKmod` | `asset_kmod` |
| `AssetSoftware` | `asset_software` |

### 事件模型: `internal/model/execve.go`

| Model | 数据库表 |
|-------|---------|
| `AssetExecve` | `asset_execve` |

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
| **5050-5062** | 资产采集 | Process, Port, User, Service, Software, Container, EnvSuspicious, Image, ImagePackage, WebService, Database, Kmod |
| **6001-6010** | 安全告警 | SSH暴破, FTP暴破, 高危命令, 反弹Shell, 异常登录, 本地提权, eBPF反弹Shell, 恶意请求, 网络攻击 |
| **6060-6062** | 恶意扫描 | 扫描状态, 文件检测, 进程检测 |
| **8000-8010** | 基线检查 | 检查结果, 任务状态 |

---

## 关键文件汇总

| 层级 | 文件路径 |
|------|---------|
| gRPC 入口 | `internal/grpc/handler/transfer.go` |
| 告警映射 | `internal/mapper/alert_mapper.go` |
| 资产映射 | `internal/mapper/asset_mapper.go` |
| 事件映射 | `internal/mapper/execve_mapper.go` |
| 基线映射 | `internal/mapper/baseline_mapper.go` |
| 告警模型 | `internal/model/alert.go` |
| 资产模型 | `internal/model/asset.go` |
| 事件模型 | `internal/model/execve.go` |
| 基线模型 | `internal/model/baseline.go` |
| 告警写库 | `internal/db/repository/alert_repository.go` |
| 资产写库 | `internal/db/repository/asset_repository.go` |
| 事件写库 | `internal/db/repository/execve_repository.go` |
| 基线写库 | `internal/db/repository/baseline_repository.go` |
| DB 连接 | `internal/db/postgres.go` |
| Protobuf | `proto/grpc.pb.go` |
