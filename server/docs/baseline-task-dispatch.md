# Server 下发基线检测任务完整链路

## 概述

基线检测任务支持两种下发模式：

| 模式 | 参数 | 规则来源 | 适用场景 |
|------|------|----------|----------|
| **内置基线** | `baseline_id` | Agent 本地 YAML 配置文件 | 使用预定义的标准基线（如 1400-Ubuntu） |
| **自定义模板** | `template_id` | Server 端数据库模板 + 检查项 | 使用自定义的检查规则 |

## 数据流全景

```
┌─────────────────────────────────────────────────────────────────────────┐
│  HTTP Client                                                            │
│  POST /api/baseline/check                                               │
│  {"agent_ids":["xxx"], "baseline_id":1400}  或  {"template_id":2}       │
└──────────────────────────────┬──────────────────────────────────────────┘
                               ▼
┌──────────────────────────────────────────────────────────────────────────┐
│  Server: baseline_handler.SendBaselineCheck()                            │
│                                                                          │
│  ┌─ template_id > 0 ─────────────────────────────────┐                   │
│  │  1. DB baseline_template (id=N) → 模板元信息       │                   │
│  │  2. DB baseline_check_item (baseline_id=N) → 检查项│                   │
│  │  3. 构建 baselineInfoForAgent（完整规则）           │                   │
│  │  4. check_rules 字段通过 json.RawMessage 原样嵌入 │                   │
│  └────────────────────────────────────────────────────┘                   │
│  ┌─ baseline_id > 0 ─────────────────────────────────┐                   │
│  │  仅传递 baseline_id + check_id_list（无完整规则）   │                   │
│  └────────────────────────────────────────────────────┘                   │
│                                                                          │
│  json.Marshal(taskData) → Task.Data (JSON string)                        │
│  构建 proto.Command { Task: { DataType:8000, ObjectName:"baseline" } }   │
└──────────────────────────────┬───────────────────────────────────────────┘
                               ▼
┌──────────────────────────────────────────────────────────────────────────┐
│  gRPC 双向流传输 (Transfer service, port 50051)                          │
│  Server stream.Send(Command) → Agent stream.Recv()                       │
└──────────────────────────────┬───────────────────────────────────────────┘
                               ▼
┌──────────────────────────────────────────────────────────────────────────┐
│  Agent: transport.handleReceive()                                        │
│  → plugin.SendTask(cmd.Task) → 通过管道(fd3)发送给 baseline 插件进程      │
└──────────────────────────────┬───────────────────────────────────────────┘
                               ▼
┌──────────────────────────────────────────────────────────────────────────┐
│  Baseline Plugin: main.go → check.Analysis(task.Data)                    │
│                                                                          │
│  ┌─ BaselineInfo != nil ──────────────────────────────┐                  │
│  │  使用 Server 下发的完整规则（template_id 模式）     │                  │
│  └────────────────────────────────────────────────────┘                  │
│  ┌─ BaselineInfo == nil ──────────────────────────────┐                  │
│  │  加载本地 config/linux/{baseline_id}.yaml           │                  │
│  └────────────────────────────────────────────────────┘                  │
│                                                                          │
│  遍历 CheckList → AnalysisRule(check) → 返回检查结果                     │
└──────────────────────────────┬───────────────────────────────────────────┘
                               ▼
┌──────────────────────────────────────────────────────────────────────────┐
│  结果上报                                                                │
│  DataType 8000: 基线检查结果 (RetBaselineInfo JSON)                      │
│  DataType 8010: 任务状态 (succeed / failed)                              │
│  Plugin → 管道(fd4) → Agent Buffer → gRPC stream → Server               │
└──────────────────────────────┬───────────────────────────────────────────┘
                               ▼
┌──────────────────────────────────────────────────────────────────────────┐
│  Server: transfer.processBaselineResult()                                │
│  → mapper.MapBaselineResult() 解析结果                                   │
│  → baselineRepo.CreateCheckResult()       写入 baseline_check_result     │
│  → baselineRepo.BatchCreateCheckDetails() 写入 baseline_check_detail     │
└──────────────────────────────────────────────────────────────────────────┘
```

## 第 1 步：HTTP API 接口

### 请求

```
POST /api/baseline/check
Content-Type: application/json
```

### 请求体

```go
// internal/http/baseline_handler.go:18-24
type BaselineCheckRequest struct {
    AgentIDs    []string `json:"agent_ids" binding:"required"`    // 目标 agent 列表
    BaselineID  int      `json:"baseline_id"`                     // 内置基线 ID（与 template_id 二选一）
    CheckIDList []int    `json:"check_id_list"`                   // 可选，检查项过滤
    TemplateID  int64    `json:"template_id"`                     // 服务端模板 ID（与 baseline_id 二选一）
}
```

### 示例

```bash
# 模式一：内置基线
curl -X POST http://xxxxxx/api/baseline/check -H "Content-Type: application/json" -d '{"agent_ids":["agent-xxx"], "baseline_id": 1400}'

# 模式一（带检查项过滤）：
curl -X POST http://localhost:8080/api/baseline/check \
  -H "Content-Type: application/json" \
  -d '{"agent_ids":["agent-xxx"], "baseline_id": 1400, "check_id_list": [1, 2, 3]}'

# 模式二：自定义模板
curl -X POST http://localhost:8080/api/baseline/check \
  -H "Content-Type: application/json" \
  -d '{"agent_ids":["agent-xxx"], "template_id": 2}'
```

### 响应

```json
{
  "success": 1,
  "failed": 0,
  "results": [
    {"agent_id": "agent-xxx", "success": true, "message": "Task sent"}
  ]
}
```

## 第 2 步：Server 构建任务数据

### 关键数据结构

```go
// internal/http/baseline_handler.go:41-63

// 下发给 agent 的任务数据（与 agent 端 TaskData 对齐）
type baselineTaskData struct {
    BaselineId   int                    `json:"baseline_id"`
    CheckIdList  []int                  `json:"check_id_list,omitempty"`
    BaselineInfo *baselineInfoForAgent  `json:"baseline_info,omitempty"`
}

// 服务端构建的完整基线规则（与 agent 端 BaselineInfo 对齐）
type baselineInfoForAgent struct {
    BaselineId      int                 `json:"baseline_id"`
    BaselineVersion string              `json:"baseline_version"`
    CheckList       []checkInfoForAgent `json:"check_list"`
}

// 单个检查项（与 agent 端 CheckInfo 对齐）
type checkInfoForAgent struct {
    CheckId       int             `json:"check_id"`
    TitleCn       string          `json:"title_cn"`
    Security      string          `json:"security"`
    TypeCn        string          `json:"type_cn"`
    DescriptionCn string          `json:"description_cn"`
    SolutionCn    string          `json:"solution_cn"`
    Check         json.RawMessage `json:"check"`       // BaselineCheck JSON，原样嵌入
}
```

### template_id 模式的构建流程

```go
// baseline_handler.go:89-138

// 1. 从 DB 加载模板
template, _ := baselineRepo.GetTemplate(ctx, req.TemplateID)

// 2. 从 DB 加载检查项
items, _ := baselineRepo.ListCheckItemsByBaselineID(ctx, req.TemplateID)

// 3. 构建检查项列表
for _, item := range items {
    ci := checkInfoForAgent{
        CheckId:       int(item.ID),
        TitleCn:       item.ItemName,
        Security:      item.RiskLevel,
        TypeCn:        item.Category,
        DescriptionCn: item.FixSuggestion,
        SolutionCn:    item.FixSuggestion,
        Check:         json.RawMessage(item.CheckScript),  // ← 关键：DB 字段原样嵌入
    }
    checkList = append(checkList, ci)
}

// 4. 组装完整 taskData
taskData = baselineTaskData{
    BaselineId: int(req.TemplateID),
    BaselineInfo: &baselineInfoForAgent{
        BaselineId:      int(req.TemplateID),
        BaselineVersion: version,
        CheckList:       checkList,
    },
}
```

### baseline_id 模式的构建（简单）

```go
// baseline_handler.go:139-145
taskData = baselineTaskData{
    BaselineId:  req.BaselineID,      // 仅传 ID
    CheckIdList: req.CheckIDList,     // 可选的过滤列表
}
// BaselineInfo 为 nil，agent 会加载本地 YAML
```

## 第 3 步：序列化并通过 gRPC 下发

```go
// baseline_handler.go:148-165

// 序列化为 JSON 字符串
taskDataJSON, _ := json.Marshal(taskData)

// 构建 gRPC Command
cmd := &proto.Command{
    Task: &proto.Task{
        DataType:   8000,                                          // DataTypeBaselineCheck
        ObjectName: "baseline",                                    // 路由到 baseline 插件
        Data:       string(taskDataJSON),                          // 任务数据 JSON
        Token:      fmt.Sprintf("baseline-%d", time.Now().UnixNano()), // 任务追踪令牌
    },
}

// 逐个发送给目标 agent
transferServer.SendCommandWithError(agentID, cmd)
```

### template_id 模式下 Task.Data 的完整 JSON 示例

```json
{
  "baseline_id": 2,
  "baseline_info": {
    "baseline_id": 2,
    "baseline_version": "1.0",
    "check_list": [
      {
        "check_id": 16,
        "title_cn": "密码最小长度检查",
        "security": "high",
        "type_cn": "密码策略",
        "description_cn": "...",
        "solution_cn": "...",
        "check": {
          "rules": [
            {
              "type": "file_line_check",
              "param": ["/etc/login.defs"],
              "filter": "\\s*\\t*PASS_MIN_LEN\\s*\\t*(\\d+)",
              "result": "$(>=)8"
            }
          ]
        }
      },
      {
        "check_id": 21,
        "title_cn": "/etc/passwd文件权限检查",
        "security": "high",
        "type_cn": "文件权限",
        "description_cn": "...",
        "solution_cn": "...",
        "check": {
          "rules": [
            {
              "type": "file_permission",
              "param": ["/etc/passwd", "644"]
            }
          ]
        }
      }
    ]
  }
}
```

### baseline_id 模式下 Task.Data 的 JSON 示例

```json
{
  "baseline_id": 1400,
  "check_id_list": [1, 2, 3]
}
```

> `baseline_info` 为空，agent 会加载本地 `config/linux/1400.yaml`。

## 第 4 步：Agent 端接收与执行

### 传输层路由

```go
// agent/transport/transfer.go:172-204
// handleReceive() 收到 Command 后，根据 Task.ObjectName 路由到对应插件
plg.SendTask(*cmd.Task)   // ObjectName="baseline" → baseline 插件
```

### 插件接收任务

```go
// agent/business_plugins/baseline/main.go:74-102
task := pluginClient.ReceiveTask()        // 通过管道 fd3 接收
retBaselineInfo, err := check.Analysis(task.Data)  // task.Data 就是上面的 JSON 字符串
```

### 反序列化与执行

```go
// agent/business_plugins/baseline/check/analysis.go:134-158
func Analysis(data interface{}) (RetBaselineInfo, error) {
    var taskData TaskData
    json.Unmarshal([]byte(data.(string)), &taskData)
    return AnalysisBaseline(taskData)
}

// analysis.go:62-131
func AnalysisBaseline(taskData TaskData) (RetBaselineInfo, error) {
    if taskData.BaselineInfo != nil {
        // template_id 模式：使用 Server 下发的完整规则
        baselineInfo = *taskData.BaselineInfo
    } else {
        // baseline_id 模式：加载本地 YAML 配置
        baselineInfo = getBaselineConfigData(taskData.BaselineId)
    }

    // 遍历检查项，逐个执行规则
    for _, checkInfo := range baselineInfo.CheckList {
        ifPass, err := AnalysisRule(checkInfo.Check)  // Check 即 BaselineCheck 结构
        // ... 记录结果
    }
}
```

## check_rules 字段格式要求

`baseline_check_item.check_rules` 字段必须是合法的 `BaselineCheck` JSON，因为 Server 通过 `json.RawMessage` 原样嵌入到下发数据中，Agent 端直接反序列化为 `BaselineCheck` 结构体。

### BaselineCheck 结构

```go
// agent/business_plugins/baseline/check/rule_engine.go:13-24
type RuleStruct struct {
    Type    string      `json:"type"`      // 规则类型
    Param   []string    `json:"param"`     // 参数列表
    Filter  string      `json:"filter"`    // 可选，正则提取
    Require string      `json:"require"`   // 可选，前置条件
    Result  interface{} `json:"result"`    // 期望值
}

type BaselineCheck struct {
    Condition string       `json:"condition"`  // "all" | "any" | "none"，默认 "all"
    Rules     []RuleStruct `json:"rules"`      // 规则列表
}
```

### 支持的规则类型

| type | 作用 | param | result |
|------|------|-------|--------|
| `file_line_check` | 逐行匹配文件内容 | `["文件路径"]` | 正则或比较运算符 |
| `command_check` | 执行命令匹配输出 | `["shell命令"]` | 正则或字符串 |
| `file_permission` | 检查文件权限 | `["文件路径", "权限值"]` | 不需要 |
| `file_user_group` | 检查文件属主 | `["文件路径", "UID:GID"]` | 不需要 |
| `if_file_exist` | 检查文件是否存在 | `["文件路径"]` | `true` / `false` |
| `file_md5_check` | 校验文件 MD5 | `["文件路径", "md5值"]` | 不需要 |
| `func_check` | 调用内置函数 | `["函数名"]` | `true` |

### result 支持的运算符

| 运算符 | 含义 | 示例 |
|--------|------|------|
| `$(<=)` | 小于等于 | `$(<=)90` |
| `$(>=)` | 大于等于 | `$(>=)8` |
| `$(<)` | 小于 | `$(<)5` |
| `$(>)` | 大于 | `$(>)0` |
| `$(&&)` | 逻辑与 | `$(not)^#$(&&)password.*pam_cracklib` |
| `$(not)` | 逻辑非 | `$(not)^root:$(&&)^\w+:\w+:0:` |

### check_rules JSON 示例

```json
// 单规则，无 condition（默认 all）
{"rules":[{"type":"file_line_check","param":["/etc/login.defs"],"filter":"\\s*\\t*PASS_MAX_DAYS\\s*\\t*(\\d+)","result":"$(<=)90"}]}

// 多规则 + condition
{"condition":"all","rules":[{"type":"file_permission","param":["/etc/passwd","644"]},{"type":"file_permission","param":["/etc/shadow","400"]}]}

// 反向检查（不应存在空密码）
{"condition":"none","rules":[{"type":"file_line_check","param":["/etc/shadow"],"result":"^\\w+::"}]}
```

## 相关源文件索引

| 组件 | 文件 | 关键行 |
|------|------|--------|
| HTTP Handler | `internal/http/baseline_handler.go` | 65-199 |
| 数据类型常量 | `internal/http/datatype.go` | 30-32 |
| gRPC 下发 | `internal/grpc/handler/transfer.go` | SendCommandWithError |
| gRPC 结果接收 | `internal/grpc/handler/transfer.go` | 703-718 |
| 结果解析 | `internal/mapper/baseline_mapper.go` | 39-120 |
| DB 操作 | `internal/db/repository/baseline_repository.go` | 20-82 |
| Proto 定义 | `proto/grpc.proto` | Command, Task, PackagedData |
| Agent 传输层 | `agent/transport/transfer.go` | 156-240 |
| Agent 插件入口 | `agent/business_plugins/baseline/main.go` | 70-102 |
| Agent 规则引擎 | `agent/business_plugins/baseline/check/rule_engine.go` | 13-341 |
| Agent 规则执行 | `agent/business_plugins/baseline/check/rules.go` | 27-288 |
| Agent 分析入口 | `agent/business_plugins/baseline/check/analysis.go` | 62-158 |
