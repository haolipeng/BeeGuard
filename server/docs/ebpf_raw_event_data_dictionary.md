# eBPF 原始事件数据字典

本文档记录服务端（server）接收并存储的 eBPF 原始事件的表结构、字段映射和精简策略。

---

## 1. Connect 出站连接事件（DataType 60）

### 概述

记录 eBPF 捕获的 `connect` 系统调用事件，用于追踪主机上所有出站网络连接行为。

### 表结构 `event_connect`

| 列名 | 类型 | 约束 | 索引 | 说明 |
|------|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | PK | 主键ID |
| agent_id | VARCHAR(64) | NOT NULL | YES | Agent唯一标识 |
| host_name | VARCHAR(255) | | | 主机名 |
| host_ip | VARCHAR(64) | | | 主机IP地址 |
| pid | INTEGER | NOT NULL | | 进程ID（线程ID） |
| tgid | INTEGER | | | 线程组ID（进程ID） |
| ppid | INTEGER | | | 父进程ID |
| uid | INTEGER | | | 用户ID |
| comm | VARCHAR(16) | | YES | 进程名（最多16字节） |
| exe_path | VARCHAR(256) | | YES | 可执行文件完整路径 |
| protocol | VARCHAR(8) | | | 协议类型（tcp/udp） |
| remote_ip | VARCHAR(64) | | YES | 远端IP地址 |
| remote_port | INTEGER | | | 远端端口 |
| pid_tree | TEXT | | | 进程树（预留） |
| event_time | TIMESTAMP | NOT NULL | YES | 事件发生时间 |
| created_at | TIMESTAMP | DEFAULT NOW() | | 记录创建时间 |

### Agent 上报字段 vs 入库字段

| Agent 字段 | 入库 | 对应列 | 精简原因 |
|------------|------|--------|----------|
| pid | YES | pid | 进程标识 |
| tgid | YES | tgid | 进程标识 |
| ppid | YES | ppid | 进程标识 |
| uid | YES | uid | 进程标识 |
| comm | YES | comm | 进程信息 |
| exe_path | YES | exe_path | 进程信息 |
| protocol | YES | protocol | 区分 TCP/UDP |
| remote_ip | YES | remote_ip | 连接目标（核心字段） |
| remote_port | YES | remote_port | 连接目标（核心字段） |
| local_ip | NO | - | AgentContext 已有 host_ip |
| local_port | NO | - | 临时源端口，无分析价值 |
| retval | NO | - | 只捕获成功连接，恒为 0 |
| pid_tree | 预留 | pid_tree | Agent 当前未发送该字段 |

### 入库验证条件

- `PID > 0`：进程ID必须有效
- `RemoteIP != ""`：必须有连接目标IP

---

## 2. DNS 查询事件（DataType 63）

### 概述

记录 eBPF 捕获的 DNS 查询事件，用于追踪主机上所有域名解析行为。

### 表结构 `event_dns`

| 列名 | 类型 | 约束 | 索引 | 说明 |
|------|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | PK | 主键ID |
| agent_id | VARCHAR(64) | NOT NULL | YES | Agent唯一标识 |
| host_name | VARCHAR(255) | | | 主机名 |
| host_ip | VARCHAR(64) | | | 主机IP地址 |
| pid | INTEGER | NOT NULL | | 进程ID（线程ID） |
| tgid | INTEGER | | | 线程组ID（进程ID） |
| ppid | INTEGER | | | 父进程ID |
| uid | INTEGER | | | 用户ID |
| comm | VARCHAR(16) | | YES | 进程名（最多16字节） |
| exe_path | VARCHAR(256) | | YES | 可执行文件完整路径 |
| domain | VARCHAR(255) | | YES | 查询域名 |
| query_type | VARCHAR(16) | | | 查询类型（A/AAAA/CNAME/MX/TXT等） |
| pid_tree | TEXT | | | 进程树（预留） |
| event_time | TIMESTAMP | NOT NULL | YES | 事件发生时间 |
| created_at | TIMESTAMP | DEFAULT NOW() | | 记录创建时间 |

### Agent 上报字段 vs 入库字段

| Agent 字段 | 入库 | 对应列 | 精简原因 |
|------------|------|--------|----------|
| pid | YES | pid | 进程标识 |
| tgid | YES | tgid | 进程标识 |
| ppid | YES | ppid | 进程标识 |
| uid | YES | uid | 进程标识 |
| comm | YES | comm | 进程信息 |
| exe_path | YES | exe_path | 进程信息 |
| domain | YES | domain | 查询域名（核心字段） |
| query_type | YES | query_type | 查询类型 A/AAAA/CNAME 等 |
| dns_server_ip | NO | - | 通常为本地 resolver |
| dns_server_port | NO | - | 固定为 53 |
| opcode | NO | - | 标准查询固定为 0 |
| rcode | NO | - | 响应码，价值低 |
| pid_tree | 预留 | pid_tree | Agent 当前未发送该字段 |

### 入库验证条件

- `PID > 0`：进程ID必须有效
- `Domain != ""`：必须有查询域名

---

## 3. 数据流路径

```
Agent (eBPF) --> gRPC PackagedData --> TransferServer.processPayload()
  --> mapper.MapConnect() / mapper.MapDNS()     字段映射
  --> 验证必填字段
  --> connectRepo.Create() / dnsRepo.Create()   INSERT 写入
```

## 4. 典型查询场景

### 查询某主机最近的出站连接

```sql
SELECT event_time, comm, exe_path, protocol, remote_ip, remote_port
FROM event_connect
WHERE agent_id = 'xxx'
ORDER BY event_time DESC
LIMIT 50;
```

### 查询连接到特定IP的进程

```sql
SELECT agent_id, host_ip, comm, exe_path, pid, event_time
FROM event_connect
WHERE remote_ip = '1.2.3.4'
ORDER BY event_time DESC;
```

### 查询某主机最近的DNS查询

```sql
SELECT event_time, comm, exe_path, domain, query_type
FROM event_dns
WHERE agent_id = 'xxx'
ORDER BY event_time DESC
LIMIT 50;
```

### 查询解析特定域名的进程

```sql
SELECT agent_id, host_ip, comm, exe_path, pid, event_time
FROM event_dns
WHERE domain = 'example.com'
ORDER BY event_time DESC;
```
