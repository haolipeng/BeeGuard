# E2E 测试目录

本目录包含端到端（End-to-End）测试，用于在本地验证各插件的完整流程。

## 目录结构

```
tests/e2e/
├── baseline/          # Baseline 插件 E2E 测试
│   ├── main.go        # 测试主程序
│   ├── test.sh        # 一键编译 + 运行脚本
│   ├── go.mod
│   └── README.md      # 详细说明文档
└── collector/         # Collector 插件 E2E 测试
    ├── main.go        # 测试主程序（standalone 模式）
    ├── go.mod
    └── README.md      # 详细说明（可参考 docs/standalone-test/collector-testing.md）
```

## 环境要求

- **Go**：1.25+（以项目 go.mod 为准）
- **工作目录**：建议在 **agent 仓库根目录** 下执行编译与脚本
- **权限**：部分测试（如 collector）需要 **root** 权限读取系统信息

---

## 编译

E2E 测试会启动“测试用 agent”并加载已编译好的插件，因此需要先编译插件（及可选的主程序）。

### 在 agent 根目录编译

```bash
# 进入 agent 仓库根目录
cd /path/to/agent

# 编译主程序 + 所有插件（推荐）
make build
```

编译成功后，插件二进制位于：

- `build/plugins/baseline/baseline`
- `build/plugins/collector/collector`
- 其他插件见 `build/plugins/` 下对应子目录

---

## 运行指南

### 1. Baseline 插件测试（推荐用脚本）

Baseline 测试脚本会：编译 baseline 插件 → 拷贝到 `/tmp/plugin/baseline/` → 运行测试程序。

```bash
cd tests/e2e/baseline
chmod +x test.sh
./test.sh
```

脚本会先到 `business_plugins/baseline` 执行 `go build -o baseline main.go`，再回到 `tests/e2e/baseline` 执行 `go run main.go`。无需事先执行 `make build`。

**手动运行（不用脚本）：**

```bash
# 1）编译插件并放到 agent 可发现的位置
make build-plugins   # 或在 business_plugins/baseline 下 go build
mkdir -p /tmp/plugin/baseline
cp business_plugins/baseline/baseline /tmp/plugin/baseline/
chmod +x /tmp/plugin/baseline/baseline

# 2）运行测试程序
cd tests/e2e/baseline
go mod tidy
go run main.go
```

按 `Ctrl+C` 停止测试。

### 2. Collector 插件测试

Collector E2E 使用 **standalone 模式**，测试程序会从 **agent 根目录下的 `build/plugins`** 加载插件（代码中写死为 `../../../build/plugins`，即相对 `tests/e2e/collector` 的 agent 根目录）。

**步骤一：先编译插件**

在 **agent 根目录** 执行：

```bash
make build-plugins
# 或
make build
```

确认存在：`build/plugins/collector/collector`。

**步骤二：运行 E2E 测试程序**

```bash
cd tests/e2e/collector
go mod tidy
go run main.go
```

默认会通过环境变量 `HANDLER` 只跑部分 Handler（如 `user`）；可在 `main.go` 中修改 `os.Setenv("HANDLER", "user")` 来切换或跑全量。
按 `Ctrl+C` 停止；采集结果会打印到终端，并可选写入 `collector_records.json`。

**Collector Handler 列表：**

| Handler 名称 | DataType | 采集间隔 | HANDLER 值 | 说明 |
|-------------|----------|---------|-----------|------|
| ProcessHandler | 5050 | 1h | `process` | 进程采集 |
| PortHandler | 5051 | 1h | `port` | 端口采集 |
| UserHandler | 5052 | 6h | `user` | 用户账户采集 |
| ServiceHandler | 5054 | 6h | `service` | 系统服务采集 |
| SoftwareHandler | 5055 | 6h | `software` | 软件包采集 |
| ContainerHandler | 5056 | 6h | `container` | 容器资产采集 |
| EnvSuspiciousHandler | 5057 | 6h | `env_suspicious` | 可疑环境变量检测 |
| ImageHandler | 5058 | 6h | `image` | 容器镜像采集 |
| ImagePackageHandler | 5059 | 6h | `image_package` | 镜像软件包采集 |
| WebServiceHandler | 5060 | 6h | `web_service` | Web 服务采集（nginx/apache） |
| DatabaseHandler | 5061 | 6h | `database` | 数据库服务采集 |
| KmodHandler | 5062 | 1h | `kmod` | 内核模块采集 |

**Web 服务采集字段（DataType 5060）：**

| 字段 | 说明 |
|------|------|
| `app_name` | 应用名称（nginx / apache） |
| `server_type` | 服务器类型 |
| `version` | 版本号 |
| `run_user` | 运行用户 |
| `path` | 配置文件路径 |
| `site_domain` | 站点域名（从配置文件中解析，逗号分隔，过滤 `_`/`localhost`/`*`） |

> `site_domain` 解析逻辑：读取 nginx `server_name` 或 apache `ServerName`/`ServerAlias` 指令，展开一级 `include` 文件，去重后以逗号拼接，超过 255 字符在最后一个逗号处截断。

---

## 测试说明

E2E 测试会：

1. 编译对应插件（或使用已有 `build/plugins/` 产物）
2. 将插件放到约定目录（Baseline：`/tmp/plugin/{插件名}/`；Collector：`build/plugins/`）
3. 启动测试 agent（plugin daemon + 任务下发）
4. 加载插件并发送测试任务
5. 接收并打印（或写入 JSON）插件返回的结果

## 注意事项

- 测试可能需要 **root** 权限（如 collector 读系统服务、端口等）。
- 测试会创建临时目录与文件（如 `/tmp/plugin/`、`collector_records.json`）。
- 程序会持续运行直到手动 `Ctrl+C` 退出。
- 更多 Collector 说明（Handler 列表、HANDLER 环境变量、standalone 详解）见：`docs/standalone-test/collector-testing.md`。
