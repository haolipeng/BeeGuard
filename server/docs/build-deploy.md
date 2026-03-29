# Server (server) 编译部署文档

本文档指导如何编译、部署和运行 Server (server)。

---

## 一、概述

Server (server) 是安全管理平台的服务端程序，负责：
- 提供 gRPC 服务，接收 Agent 上报的数据
- 提供 HTTP API 接口
- 将采集数据存储到 PostgreSQL 数据库
- 下发任务指令给 Agent

---

## 二、数据库准备

确保 PostgreSQL 已安装并创建数据库 `soc`，然后通过数据库迁移初始化表结构：

```bash
# 使用 migrate 工具执行数据库迁移
cd /home/work/goProject/src/BeeGuard/server
make migrate-up

# 验证
sudo -u postgres psql -d soc -c "\dt"
```

---

## 三、编译部署

```bash
cd /home/work/goProject/src/BeeGuard/server

# 编译源代码
make build

# 部署到 /opt/cloudsec/server/
make deploy
```



### 设置编译产物的版本

```bash
# 编译并指定版本号
make build VERSION=1.0.0
```



### 编译产物

```
build/
└── server                      # Server 主程序
```



### 部署后目录结构

```
/opt/cloudsec/server/
├── bin/
│   └── server                   # Server 主程序
├── conf/
│   └── server.yaml             # 配置文件
└── logs/
    └── server/                 # Server 日志目录
```

---

## 四、配置文件

### 配置文件位置

- 模板位置：`server/conf/server.yaml`
- 部署位置：`/opt/cloudsec/server/conf/server.yaml`

### 配置项说明

```yaml
# 服务器配置
server:
  port: 50051                   # gRPC 服务端口
  http_port: 8080               # HTTP API 端口
  max_recv_msg_size: 16         # 最大接收消息大小 (MB)
  max_send_msg_size: 16         # 最大发送消息大小 (MB)

# 数据库配置
database:
  host: localhost
  port: 5432
  user: postgres
  password: "happy"             # 修改为你的数据库密码
  database: soc
  pool_size: 10                 # 连接池大小

# 日志配置
log:
  level: info                   # debug, info, warn, error
```

---

## 五、运行

### 前提条件

1. PostgreSQL 服务已启动
2. 数据库 `soc` 已创建并导入表结构
3. 配置文件中数据库连接信息正确

### 启动方式

```bash
# 方式一：使用部署目录（推荐）
/opt/cloudsec/server/bin/server -config /opt/cloudsec/server/conf/server.yaml

# 方式二：使用源码目录（开发调试）
cd /home/work/goProject/src/BeeGuard/server
make run

# 方式三：后台运行
nohup /opt/cloudsec/server/bin/server -config /opt/cloudsec/server/conf/server.yaml \
    > /opt/cloudsec/server/logs/server/server.log 2>&1 &
```

### 启动日志

```
2026/01/27 10:00:00 配置加载成功: grpc_port=50051, http_port=8080, log_level=info
2026/01/27 10:00:00 数据库连接成功: host=localhost, database=soc
2026/01/27 10:00:00 gRPC Server 启动，监听端口 :50051
2026/01/27 10:00:00 HTTP Server 启动，监听端口 :8080
```

### 验证服务

```bash
# 检查端口
netstat -tlnp | grep -E "50051|8080"

# 检查进程
ps aux | grep server
```

---

## 六、清理

```bash
# 停止 Server
pkill -f "/opt/cloudsec/server/bin/server"

# 清理日志
sudo rm -rf /opt/cloudsec/server/logs/server/*
```
