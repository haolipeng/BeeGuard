# 高危命令检测测试方案 - Standalone 模式

## 目标
为 agent 添加 standalone 模式，用于本地测试高危命令检测功能：
- 可选择加载哪些插件
- 检测结果不上报 server，而是写日志或文件

---

## 方案概述

**架构变化**:
```
正常模式: 插件 → buffer → transport → gRPC Server
Standalone: 插件 → buffer → standalone.OutputHandler → 日志/文件
```

---

## 需要修改/新增的文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `config/config.go` | 修改 | 添加 StandaloneConfig 结构 |
| `main.go` | 修改 | 添加命令行参数，条件启动 |
| `standalone/output.go` | 新增 | standalone 输出处理器 |
| `plugin/plugin.go` | 修改 | 添加本地插件自动加载 |

---

## Step 1: 扩展配置结构

**文件**: `config/config.go`

添加 StandaloneConfig:
```go
type StandaloneConfig struct {
    Enabled       bool     `yaml:"enabled"`
    Output        string   `yaml:"output"`        // "log" 或 "file"
    OutputPath    string   `yaml:"output_path"`   // 输出文件路径
    Plugins       []string `yaml:"plugins"`       // 指定加载的插件
    FlushInterval int      `yaml:"flush_interval"` // 刷新间隔（秒）
}

type Config struct {
    // ... 现有字段 ...
    Standalone *StandaloneConfig `yaml:"standalone,omitempty"`
}
```

修改 `ValidateAndSetDefaults`:
- standalone 模式下 server 不是必须的
- 设置 standalone 默认值

---

## Step 2: 添加命令行参数

**文件**: `main.go`

新增参数:
```go
standalone := flag.Bool("standalone", false, "Enable standalone mode")
outputMode := flag.String("output", "log", "Output mode: log or file")
outputPath := flag.String("output-path", "", "Output file path")
plugins := flag.String("plugins", "", "Comma-separated plugin list")
```

条件启动:
```go
if cfg.Standalone != nil && cfg.Standalone.Enabled {
    zap.S().Info("running in standalone mode")
    go standalone.StartOutputHandler(Context, wg)
} else {
    go transport.StartTransfer(Context, wg)
}
```

---

## Step 3: 创建 Standalone 输出处理器

**文件**: `standalone/output.go` (新增)

核心功能:
1. 定期从 buffer 读取 EncodedRecord
2. 解析 Payload，提取检测结果字段
3. **仅输出检测到高危命令的记录** (有 rule_id 字段)
4. 输出到日志 (zap) 或 JSON 文件

```go
func StartOutputHandler(ctx context.Context, wg *sync.WaitGroup) {
    ticker := time.NewTicker(interval)
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            recs := buffer.ReadEncodedRecords()
            for _, rec := range recs {
                // 解析 Payload
                payload := parsePayload(rec.Data)

                // 仅输出高危命令检测结果（有 rule_id 字段）
                ruleID, ok := payload.Fields["rule_id"]
                if !ok || ruleID == "" {
                    continue  // 跳过未触发规则的事件
                }

                output := buildOutput(rec, payload)
                if cfg.Output == "log" {
                    zap.S().Infow("dangerous command detected",
                        "rule_id", payload.Fields["rule_id"],
                        "rule_name", payload.Fields["rule_name"],
                        "severity", payload.Fields["severity"],
                        "command", payload.Fields["command"],
                    )
                } else {
                    file.Write(json.Marshal(output))
                }
            }
        }
    }
}
```

---

## Step 4: 添加本地插件自动加载

**文件**: `plugin/plugin.go`

在 `Startup` 函数中添加 standalone 模式逻辑:
```go
if cfg.Standalone != nil && cfg.Standalone.Enabled {
    go loadLocalPlugins(ctx, cfg.Standalone.Plugins)
}
```

`loadLocalPlugins` 函数:
1. 扫描 `plugins_directory` 目录
2. 过滤只加载指定的插件
3. 调用 `Load()` 加载插件

---

## 使用示例

### 命令行方式

```bash
# 仅加载 driver 插件，输出到日志
./agent -standalone -plugins=driver -output=log

# 输出到 JSON 文件
./agent -standalone -plugins=driver -output=file -output-path=/tmp/results.json

# 测试模式 + standalone
./agent -test -standalone -plugins=driver
```

### 配置文件方式

**agent-standalone.yaml**:
```yaml
working_directory: "/tmp/cloudsec-agent"
plugins_directory: "/opt/cloudsec/plugins"

standalone:
  enabled: true
  output: "file"
  output_path: "/tmp/detection-results.json"
  flush_interval: 5
  plugins:
    - driver
```

```bash
./agent -config=agent-standalone.yaml
```

---

## 输出格式

### 日志输出 (zap)
```
INFO    dangerous command detected    {"timestamp": 1705288245, "data_type": 59,
        "rule_id": "DC001", "rule_name": "危险删除操作",
        "severity": "critical", "command": "rm -rf /"}
```

### JSON 文件输出
```json
{"timestamp":1705288245,"data_type":59,"data":{"rule_id":"DC001","rule_name":"危险删除操作","severity":"critical","command":"rm -rf /","matched_pattern":"rm\\s+.*-rf\\s+/"}}
```

---

## 测试流程

1. 编译 agent: `go build -o agent main.go`
2. 编译 driver 插件: `cd business_plugins/driver && go build -o driver`
3. 准备插件目录: `mkdir -p /opt/cloudsec/plugins/driver && cp driver /opt/cloudsec/plugins/driver/`
4. 启动 standalone 模式: `sudo ./agent -standalone -plugins=driver`
5. 执行测试命令: `rm -rf /tmp/test_nonexistent`
6. 查看检测结果: 日志或 JSON 文件

---

## 实现步骤

1. **修改 `config/config.go`**
   - 添加 StandaloneConfig 结构
   - 修改验证逻辑

2. **创建 `standalone/output.go`**
   - 实现 StartOutputHandler
   - 支持日志和文件两种输出

3. **修改 `main.go`**
   - 添加命令行参数
   - 条件启动 transport 或 standalone 输出

4. **修改 `plugin/plugin.go`**
   - 添加 loadLocalPlugins 函数
   - Startup 中调用本地加载

5. **测试验证**
   - 单元测试
   - 手动触发高危命令测试

---

## 关键文件路径

| 文件 | 路径 |
|------|------|
| 主入口 | `main.go` |
| 配置 | `config/config.go` |
| 插件管理 | `plugin/plugin.go` |
| 传输层 | `transport/transfer.go` (参考) |
| 缓冲区 | `buffer/buffer.go` (参考) |
| Standalone 输出 | `standalone/output.go` (新增) |

---

## 注意事项

1. **需要 root 权限**: driver 插件的 eBPF 需要 root
2. **最小化改动**: 复用 buffer 机制，仅替换传输层
3. **向后兼容**: 不影响现有 gRPC 模式
4. **数据解析**: 需要在 output.go 中解析 bridge.Payload
