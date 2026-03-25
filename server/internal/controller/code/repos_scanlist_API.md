# 仓库扫描列表 API 文档

本文档介绍了仓库扫描列表相关的 API 接口，包括创建、查询、更新和删除操作。

## API 端点

所有 API 都通过 `/api1/code_scan_list` 前缀访问。

### 创建仓库扫描记录

- **POST** `/api1/code_scan_list/scan_lists`
- **描述**: 创建一个新的仓库扫描记录
- **请求体**:
  ```json
  {
    "repo_id": 1,
    "scan_type": "full",
    "status": "pending"
  }
  ```
- **响应**:
  ```json
  {
    "message": "创建成功",
    "data": {
      "id": 1,
      "repo_id": 1,
      "scan_type": "full",
      "status": "pending",
      "created_at": "2026-01-23T10:00:00Z",
      "updated_at": "2026-01-23T10:00:00Z"
    }
  }
  ```

### 获取单个仓库扫描记录

- **GET** `/api1/code_scan_list/scan_lists/:id`
- **描述**: 根据ID获取特定的仓库扫描记录
- **参数**: `id` - 仓库扫描记录的ID
- **响应**:
  ```json
  {
    "message": "查询成功",
    "data": {
      "id": 1,
      "repo_id": 1,
      "scan_type": "full",
      "status": "pending",
      "created_at": "2026-01-23T10:00:00Z",
      "updated_at": "2026-01-23T10:00:00Z"
    }
  }
  ```

### 获取所有仓库扫描记录

- **GET** `/api1/code_scan_list/scan_lists`
- **描述**: 获取所有仓库扫描记录（支持分页）
- **查询参数**:
  - `page`: 页码，默认为 1
  - `pageSize`: 每页大小，默认为 10
- **响应**:
  ```json
  {
    "message": "查询成功",
    "data": [
      {
        "id": 1,
        "repo_id": 1,
        "scan_type": "full",
        "status": "pending",
        "created_at": "2026-01-23T10:00:00Z",
        "updated_at": "2026-01-23T10:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 10
  }
  ```

### 更新仓库扫描记录

- **PUT** `/api1/code_scan_list/scan_lists/:id`
- **描述**: 根据ID更新特定的仓库扫描记录
- **参数**: `id` - 仓库扫描记录的ID
- **请求体**:
  ```json
  {
    "status": "completed",
    "result": "Scan completed successfully"
  }
  ```
- **响应**:
  ```json
  {
    "message": "更新成功",
    "data": {
      "id": 1,
      "repo_id": 1,
      "scan_type": "full",
      "status": "completed",
      "result": "Scan completed successfully",
      "created_at": "2026-01-23T10:00:00Z",
      "updated_at": "2026-01-23T10:30:00Z"
    }
  }
  ```

### 删除仓库扫描记录

- **DELETE** `/api1/code_scan_list/scan_lists/:id`
- **描述**: 根据ID删除特定的仓库扫描记录
- **参数**: `id` - 仓库扫描记录的ID
- **响应**:
  ```json
  {
    "message": "删除成功"
  }
  ```

## 字段说明

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| id | integer | 否 | 记录唯一标识符，自动生成 |
| repo_id | integer | 是 | 关联的仓库ID |
| scan_type | string | 是 | 扫描类型，如 "full", "incremental" |
| status | string | 否 | 扫描状态，默认为 "pending" |
| created_at | datetime | 否 | 记录创建时间，自动生成 |
| updated_at | datetime | 否 | 记录更新时间，自动生成 |
| started_at | datetime | 否 | 扫描开始时间 |
| finished_at | datetime | 否 | 扫描结束时间 |
| result | text | 否 | 扫描结果详情 |
| error_msg | text | 否 | 错误信息 |

## 示例请求

### 使用 curl 创建仓库扫描记录

```bash
curl -X POST http://localhost:8080/api1/code_scan_list/scan_lists \
  -H "Content-Type: application/json" \
  -d '{
    "repo_id": 1,
    "scan_type": "full",
    "status": "pending"
  }'
```

### 使用 PowerShell 创建仓库扫描记录

```powershell
$Body = @{
    repo_id = 1
    scan_type = "full"
    status = "pending"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api1/code_scan_list/scan_lists" -Method Post -Body $Body -ContentType "application/json"
```

### 使用 curl 获取仓库扫描记录

```bash
curl -X GET http://localhost:8080/api1/code_scan_list/scan_lists/1
```

### 使用 curl 更新仓库扫描记录

```bash
curl -X PUT http://localhost:8080/api1/code_scan_list/scan_lists/1 \
  -H "Content-Type: application/json" \
  -d '{
    "status": "completed",
    "result": "Scan completed successfully"
  }'
```

### 使用 curl 删除仓库扫描记录

```bash
curl -X DELETE http://localhost:8080/api1/code_scan_list/scan_lists/1
```