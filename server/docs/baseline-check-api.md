# 基线检测接口使用说明

## 接口概览

| 项目 | 说明 |
|------|------|
| URL | `/api/baseline/check` |
| Method | `POST` |
| Content-Type | `application/json` |
| 认证 | 无 |
| 功能 | 向指定 Agent 下发基线检测任务 |

## 请求参数

### 请求体 (JSON)

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `agent_ids` | `string[]` | 是 | 目标 Agent ID 列表，至少包含一个 |
| `template_id` | `number` | 是 | 服务端基线模板 ID，必须大于 0 |
| `baseline_id` | `string` | 否 | 检测批次 ID，前端可自行生成用于任务追踪 |
| `check_id_list` | `number[]` | 否 | 指定检查项 ID 列表，不传则检查模板下所有项 |

### 参数说明

- **agent_ids**: 从 `GET /api/agents` 接口获取在线的 Agent 列表，选择需要执行基线检测的目标主机
- **template_id**: 从基线模板管理页面获取，对应 `baseline_template` 表的主键 ID。模板必须处于启用状态（`is_enabled = 1`）
- **baseline_id**: 前端生成的批次标识，用于后续查询该批次的检测结果。如不传，可在结果中通过 `agent_id` 关联
- **check_id_list**: 用于部分检测场景，仅执行模板中指定 ID 的检查项。留空表示执行模板下全部检查项

## 请求示例

### 基本调用 - 对单台主机执行全部检查

```json
{
  "agent_ids": ["agent-001"],
  "template_id": 1
}
```

### 批量调用 - 对多台主机执行检查并指定批次 ID

```json
{
  "agent_ids": ["agent-001", "agent-002", "agent-003"],
  "template_id": 2,
  "baseline_id": "task-20260306-001"
}
```

### 部分检查 - 仅执行指定检查项

```json
{
  "agent_ids": ["agent-001"],
  "template_id": 1,
  "check_id_list": [1, 3, 5],
  "baseline_id": "task-20260306-002"
}
```

## 响应格式

### 成功响应 (HTTP 200)

```json
{
  "success": 2,
  "failed": 1,
  "results": [
    {
      "agent_id": "agent-001",
      "success": true,
      "message": "Task sent"
    },
    {
      "agent_id": "agent-002",
      "success": true,
      "message": "Task sent"
    },
    {
      "agent_id": "agent-003",
      "success": false,
      "message": "Agent not found"
    }
  ]
}
```

### 响应字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | `number` | 成功发送任务的 Agent 数量 |
| `failed` | `number` | 发送失败的 Agent 数量 |
| `results` | `AgentSendResult[]` | 每个 Agent 的发送结果详情 |

**AgentSendResult:**

| 字段 | 类型 | 说明 |
|------|------|------|
| `agent_id` | `string` | Agent 标识 |
| `success` | `boolean` | 该 Agent 是否发送成功 |
| `message` | `string` | 结果描述 |

### 可能的 message 值

| message | 含义 | 前端处理建议 |
|---------|------|-------------|
| `Task sent` | 任务已成功下发 | 提示用户任务已下发，等待结果 |
| `Agent not found` | Agent 不在线或不存在 | 提示用户该主机离线 |
| `Command queue full` | Agent 命令队列已满 | 提示用户稍后重试 |

### 错误响应 (HTTP 400)

**参数缺失或格式错误:**

```json
{
  "success": false,
  "message": "Invalid request: Key: 'BaselineCheckRequest.AgentIDs' Error:Field validation for 'AgentIDs' failed on the 'required' tag"
}
```

**template_id 未传或无效:**

```json
{
  "success": false,
  "message": "template_id is required"
}
```

**模板不存在或已禁用:**

```json
{
  "success": false,
  "message": "Template not found: 99"
}
```

### 错误响应 (HTTP 500)

**加载检查项失败:**

```json
{
  "success": false,
  "message": "Failed to load check items: ..."
}
```

## 前端调用示例

### JavaScript / Fetch

```javascript
async function startBaselineCheck({ agentIds, templateId, baselineId, checkIdList }) {
  const response = await fetch('/api/baseline/check', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      agent_ids: agentIds,
      template_id: templateId,
      baseline_id: baselineId,
      check_id_list: checkIdList
    })
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message);
  }

  return await response.json();
}

// 调用示例
const result = await startBaselineCheck({
  agentIds: ['agent-001', 'agent-002'],
  templateId: 1,
  baselineId: `task-${Date.now()}`
});

console.log(`成功: ${result.success}, 失败: ${result.failed}`);
result.results.forEach(r => {
  console.log(`${r.agent_id}: ${r.success ? '已下发' : r.message}`);
});
```

### Axios

```javascript
import axios from 'axios';

async function startBaselineCheck({ agentIds, templateId, baselineId, checkIdList }) {
  const { data } = await axios.post('/api/baseline/check', {
    agent_ids: agentIds,
    template_id: templateId,
    baseline_id: baselineId,
    check_id_list: checkIdList
  });
  return data;
}
```

## 前端交互流程

```
用户选择主机和基线模板
        │
        ▼
前端调用 POST /api/baseline/check
        │
        ▼
   解析响应 results
        │
        ├── 全部成功 → 提示"任务已下发，请稍后查看结果"
        │
        ├── 部分失败 → 展示失败的 Agent 列表及原因
        │                (Agent not found / Command queue full)
        │
        └── 全部失败 → 提示错误信息，引导用户检查 Agent 状态
        │
        ▼
轮询或等待基线检测结果
  (结果存储在 baseline_check_result / baseline_check_detail 表)
```

## 注意事项

1. **任务异步执行**: 该接口仅负责下发任务，返回成功表示任务已发送至 Agent，不代表检测已完成。检测结果需通过其他接口查询
2. **Agent 在线要求**: 目标 Agent 必须在线并通过 gRPC 连接到 Server，否则会返回 `Agent not found`
3. **模板有效性**: `template_id` 对应的模板必须存在且处于启用状态（`is_enabled = 1`），否则返回 400 错误
4. **批量发送**: 支持同时向多个 Agent 下发任务，接口会逐个发送并汇总结果
5. **幂等性**: 接口不具备幂等性，重复调用会重复下发任务。前端应做防重复点击处理
6. **无认证**: 当前接口未配置认证中间件，生产环境部署时需注意网络隔离或增加认证
