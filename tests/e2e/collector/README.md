# Collector 插件 E2E 测试

这是一个独立的测试程序，用于测试 collector 插件的完整采集流程，不会影响主程序。

## 一、测试流程概述

1. **编译 collector 插件** - 在 agent 根目录执行 `make build` 或 `make build-plugins`
2. **运行测试程序** - 启动测试 agent（standalone 模式），自动执行：
   - 加载 collector 插件
   - 触发指定 Handler 的采集任务
   - 接收插件返回的采集结果
   - 格式化打印到终端，并可选写入 JSON 文件

> 需要 **root** 权限，因为采集系统信息（进程、端口、服务等）需要读取 `/proc` 等系统目录。

## 二、详细步骤

### 步骤 1: 编译插件

测试程序从 `../../../build/plugins` 目录加载插件（即 agent 根目录下的 `build/plugins/`），因此需要先编译。

```bash
cd /home/work/goProject/src/company/agent

# 编译主程序 + 所有插件
make build
```

确认存在：`build/plugins/collector/collector`。

### 步骤 2: 选择要测试的 Handler

编辑 `main.go` 第 69 行的 `HANDLER` 环境变量来选择要运行的 Handler：

```go
// 只运行单个 handler
os.Setenv("HANDLER", "web_service")

// 运行多个 handler（逗号分隔）
os.Setenv("HANDLER", "web_service,user,port")

// 运行所有 handler（删除或注释掉 Setenv 行）
```



### 步骤 3: 运行测试

进入到tests/e2e/collector目录下，先编译main.go为test_collector程序，然后运行test_collector程序

```bash
cd /home/work/goProject/src/company/agent/tests/e2e/collector

go build -o test_collector main.go

sudo ./test_collector
```

### 步骤 4: 观察输出

测试程序会格式化打印采集到的记录。以 Web 服务采集为例：

```
========== Web Service Record ==========
App Name: nginx
Server Type: nginx
Version: 1.18.0
Run User: root
Config Path: /etc/nginx/nginx.conf
Site Domain: 10.107.12.99
========================================
```

JSON 输出默认启用，采集记录会追加写入 `collector_records.json`。可在 `main.go` 中修改 `enableJSONOutput` 变量控制。



## 三、Handler 值参考表

| Handler              | HANDLER 值       | DataType | 采集间隔 | 说明                         |
| -------------------- | ---------------- | -------- | -------- | ---------------------------- |
| ProcessHandler       | `process`        | 5050     | 1h       | 进程采集                     |
| PortHandler          | `port`           | 5051     | 1h       | 端口采集                     |
| UserHandler          | `user`           | 5052     | 6h       | 用户账户采集                 |
| ServiceHandler       | `service`        | 5054     | 6h       | 系统服务采集                 |
| SoftwareHandler      | `software`       | 5055     | 6h       | 软件包采集                   |
| ContainerHandler     | `container`      | 5056     | 6h       | 容器资产采集                 |
| EnvSuspiciousHandler | `env_suspicious` | 5057     | 6h       | 可疑环境变量检测             |
| ImageHandler         | `image`          | 5058     | 6h       | 容器镜像采集                 |
| ImagePackageHandler  | `image_package`  | 5059     | 6h       | 镜像软件包采集               |
| WebServiceHandler    | `web_service`    | 5060     | 6h       | Web 服务采集（nginx/apache） |
| DatabaseHandler      | `database`       | 5061     | 6h       | 数据库服务采集               |
| KmodHandler          | `kmod`           | 5062     | 1h       | 内核模块采集                 |
