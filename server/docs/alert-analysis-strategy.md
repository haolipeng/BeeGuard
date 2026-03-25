# 告警高级分析策略

## 概述

告警高级分析模块使用 AI（Ollama + qwen3.5:0.8b）对安全告警进行智能分析，生成分析报告，包括风险等级、攻击模式、攻击阶段、处置建议和 IOC 指标。

## 执行频率

| 配置项 | 默认值 | 配置位置 | 说明 |
|-------|-------|---------|------|
| 调度间隔 | 30 分钟 | `Config.ScheduleInterval` | 自动扫描间隔 |
| 启动行为 | 立即执行一次 | `scheduler.go:76` | 服务启动后立即执行首次分析 |

## 分析触发方式

```go
// 1. 自动触发（定时任务）
scheduler.Start()  // 启动调度器，每 30 分钟自动执行

// 2. 手动触发
POST /api1/analysis/trigger  // API 接口触发
```

## 分析维度

### 1. 主机维度分析 (`AnalyzeByHost`)

按主机 IP 聚合告警进行分析。

**触发条件：**
- 最近 2 小时内的告警
- 该主机告警数 >= 2 条
- 每次最多分析 20 条告警

**SQL 查询：**
```sql
SELECT host_ip
FROM v_alert_unified
WHERE alert_time >= (CURRENT_TIMESTAMP AT TIME ZONE 'UTC') - INTERVAL '2 hours'
GROUP BY host_ip
HAVING COUNT(*) >= 2
```

### 2. 攻击源分析 (`AnalyzeBySourceIP`)

按攻击源 IP 聚合告警进行分析。

**触发条件：**
- 最近 1 小时内的告警
- 来源 IP 匹配 `details->>'source_ip'` 或 `details->>'attacker_ip'`
- 每次最多分析 20 条告警

**注意：** 此分析维度目前仅在手动调用时执行，定时任务暂未自动触发。

### 3. 高危告警分析 (`AnalyzeCriticalAlert`)

对特定高危告警类型进行单条分析。

**触发条件：**
- 最近 2 小时内的告警
- 告警类型为：`reverse_shell`（反弹Shell）、`privilege_escalation`（本地提权）、`malware_scan`（文件查杀）
- 每次最多分析 10 条

**SQL 查询：**
```sql
SELECT alert_type, id, ...
FROM v_alert_unified
WHERE alert_type IN ('reverse_shell', 'privilege_escalation', 'malware_scan')
  AND alert_time >= (CURRENT_TIMESTAMP AT TIME ZONE 'UTC') - INTERVAL '2 hours'
LIMIT 10
```

## 防重复分析机制

### 缓存配置

| 配置项 | 默认值 | 说明 |
|-------|-------|------|
| 缓存 TTL | 24 小时 | 已分析告警的缓存有效期 |
| 缓存目录 | `/tmp/server/analysis_cache` | 磁盘缓存存储位置 |
| 缓存文件 | `analysis_cache.json` | 缓存数据文件 |

### 工作流程

```
1. 扫描告警
      ↓
2. FilterAnalyzed() 过滤已分析的告警
      ↓
3. 调用 Ollama AI 分析
      ↓
4. MarkBatch() 标记告警为已分析
      ↓
5. 保存分析报告
```

### 缓存键格式

```
{alert_type}:{alert_id}

示例：
- reverse_shell:123
- privilege_escalation:456
- malware_scan:789
```

### 缓存过期处理

- 每小时清理一次过期缓存
- 过期后的告警可以被重新分析
- 服务重启后从磁盘加载未过期的缓存

## 时间窗口限制

| 分析类型 | 时间窗口 | 最大告警数 |
|---------|---------|-----------|
| 主机分析 | 最近 2 小时 | 20 条 |
| 攻击源分析 | 最近 1 小时 | 20 条 |
| 高危告警分析 | 最近 2 小时 | 10 条 |

## 是否会反复分析同一告警？

**答案：不会（在 24 小时缓存有效期内）**

- 分析完成后，告警会被标记到磁盘缓存
- 下次扫描时，`FilterAnalyzed()` 会过滤掉已分析的告警
- 缓存 24 小时后过期，过期后同一告警可能会被重新分析

## 报告存储

| 存储位置 | 说明 |
|---------|------|
| 磁盘 | `/tmp/server/analysis_reports/` - JSON 文件 |
| 数据库 | `analysis_reports` 表 |

**报告保存条件：**
- 主机分析：所有报告都保存
- 攻击源分析：仅保存 `medium`/`high`/`critical` 风险等级
- 单告警分析：仅保存 `medium`/`high`/`critical` 风险等级

## 关键代码位置

| 文件 | 功能 |
|-----|------|
| `internal/analysis/scheduler.go` | 定时调度器 |
| `internal/analysis/engine.go` | 分析引擎核心逻辑 |
| `internal/analysis/cache.go` | 防重复分析缓存 |
| `internal/analysis/ollama.go` | AI 分析客户端 |
| `internal/analysis/model.go` | 数据模型定义 |
| `internal/controller/analysis/analysis.go` | API 控制器 |

## 配置示例

```go
// 在 main.go 或初始化代码中
analysis.Init(analysis.Config{
    OllamaURL:       "http://localhost:11434",
    OllamaModel:     "qwen3.5:0.8b",
    CacheDir:        "/tmp/server/analysis_cache",
    CacheTTL:        24 * time.Hour,
    ReportDir:       "/tmp/server/analysis_reports",
    ScheduleInterval: 30 * time.Minute,
    AutoStart:       true,
})
```

## 调优建议

### 1. 调整调度间隔

```go
// 更频繁的分析（适合高安全要求场景）
ScheduleInterval: 15 * time.Minute

// 较低频率（节省 AI 资源）
ScheduleInterval: 1 * time.Hour
```

### 2. 调整告警阈值

修改 `scanHostsWithAlerts` 中的 SQL：

```go
// 降低阈值，分析更多主机
HAVING COUNT(*) >= 1

// 提高阈值，只分析告警多的主机
HAVING COUNT(*) >= 5
```

### 3. 调整缓存 TTL

```go
// 更长的缓存时间（减少重复分析）
CacheTTL: 48 * time.Hour

// 更短的缓存时间（适合需要重新分析的场景）
CacheTTL: 12 * time.Hour
```

### 4. 扩展高危告警类型

修改 `scanCriticalAlerts` 中的 SQL：

```go
WHERE alert_type IN (
    'reverse_shell',
    'privilege_escalation',
    'malware_scan',
    'dangerous_command',      // 新增：高危命令
    'container_reverse_shell' // 新增：容器反弹Shell
)
```

## API 接口

| 接口 | 方法 | 说明 |
|-----|------|------|
| `/api1/analysis/trigger` | POST | 手动触发分析 |
| `/api1/analysis/db_reports` | GET | 获取报告列表 |
| `/api1/analysis/db_reports/:id` | GET | 获取报告详情 |
| `/api1/analysis/db_reports/:id` | DELETE | 删除报告 |
| `/api1/analysis/db_reports/stats` | GET | 获取统计数据 |
