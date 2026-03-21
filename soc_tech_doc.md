### **2、4、1 通用功能数据库设计**

（预留，暂无独立的通用功能表）


### 	**2、4、2 资产管理数据库设计**

#### **1、主机列表数据库设计：**

| 字段            | 类型            | 必填 | 说明                                           |
| --------------- | --------------- | ---- | ---------------------------------------------- |
| id              | BIGINT UNSIGNED | 是   | 主键ID                                         |
| agent\_id       | VARCHAR(64)     | 是   | Agent唯一标识(安装时生成，不随IP/hostname变化) |
| host\_name      | VARCHAR(128)    | 是   | 主机名称(可变，Agent上报更新)                  |
| host\_ip        | VARCHAR(45)     | 是   | 主机IP地址(可变，Agent上报更新，支持IPv6)      |
| mac\_addr       | VARCHAR(45)     | 否   | MAC地址                                        |
| os\_type        | VARCHAR(32)     | 否   | 操作系统类型: linux/windows                    |
| os\_version     | VARCHAR(64)     | 否   | 操作系统版本                                   |
| agent\_status   | TINYINT         | 否   | Agent状态: 0=离线，1=在线（Agent上报数据时设置为1，心跳超时或断连时更新为0） |
| agent\_version  | VARCHAR(32)     | 否   | Agent版本号                                    |
| last\_heartbeat | DATETIME        | 否   | 最后心跳时间                                   |
| created\_at     | DATETIME        | 是   | 创建时间                                       |
| updated\_at     | DATETIME        | 是   | 更新时间                                       |

#### **2、端口列表数据库设计：**

#### 

| 字段            | 类型            | 必填 | 说明                                           |
| --------------- | --------------- | ---- | ---------------------------------------------- |
| id              | BIGINT UNSIGNED | 是   | 主键ID                                         |
| agent\_id       | VARCHAR(64)     | 是   | Agent唯一标识(安装时生成，不随IP/hostname变化) |
| host\_name      | VARCHAR(128)    | 是   | 主机名称(可变，Agent上报更新)                  |
| host\_ip        | VARCHAR(45)     | 是   | 主机IP地址(可变，Agent上报更新，支持IPv6)      |
| os\_type        | VARCHAR(32)     | 否   | 操作系统类型: linux/windows                    |
| port            | INT             | 是   | 端口                                           |
| protocol        | SMALLINT        | 是   | 端口协议: 6=TCP, 17=UDP                        |
| listen\_ip      | VARCHAR(45)     | 是   | 监听IP                                         |
| listen\_process | VARCHAR(45)     | 是   | 监听进程                                       |
| run\_user       | VARCHAR(64)     | 否   | 运行用户                                       |
| os\_version     | VARCHAR(64)     | 否   | 操作系统版本                                   |
| agent\_status   | TINYINT         | 否   | Agent状态: 0=离线，1=在线（Agent上报数据时设置为1，心跳超时或断连时更新为0） |
| agent\_version  | VARCHAR(32)     | 否   | Agent版本号                                    |
| process\_time   | DATETIME        | 否   | 进程启动时间                                   |
| created\_at     | DATETIME        | 是   | 创建时间                                       |
| updated\_at     | DATETIME        | 是   | 更新时间                                       |

#### **3、账号列表数据库设计：**

#### 

| 字段              | 类型            | 必填 | 说明                                           |
| ----------------- | --------------- | ---- | ---------------------------------------------- |
| id                | BIGINT UNSIGNED | 是   | 主键ID                                         |
| agent\_id         | VARCHAR(64)     | 是   | Agent唯一标识(安装时生成，不随IP/hostname变化) |
| host\_name        | VARCHAR(128)    | 是   | 主机名称(可变，Agent上报更新)                  |
| host\_ip          | VARCHAR(45)     | 是   | 主机IP地址(可变，Agent上报更新，支持IPv6)      |
| os\_type          | VARCHAR(32)     | 否   | 操作系统类型: linux/windows                    |
| name              | VARCHAR(128)    | 是   | 账号名称                                       |
| uid               | INT             | 是   | UID                                            |
| status            | TINYINT         | 是   | 账号状态: 0=正常，1=即将过期，2=已过期         |
| permission        | VARCHAR(64)     | 是   | 权限: normal(普通用户)、root、sudo、root,sudo  |
| login\_type       | VARCHAR(128)    | 否   | 登录Shell: /bin/bash、/bin/sh、/sbin/nologin等 |
| last\_login\_time | DATETIME        | 否   | 最后登录时间                                   |
| created\_at       | DATETIME        | 是   | 创建时间                                       |
| updated\_at       | DATETIME        | 是   | 更新时间                                       |

#### **4、进程列表数据库设计：**

| 字段        | 类型            | 必填 | 说明                                           |
| ----------- | --------------- | ---- | ---------------------------------------------- |
| id          | BIGINT UNSIGNED | 是   | 主键ID                                         |
| agent\_id   | VARCHAR(64)     | 是   | Agent唯一标识(安装时生成，不随IP/hostname变化) |
| host\_name  | VARCHAR(128)    | 是   | 主机名称(可变，Agent上报更新)                  |
| host\_ip    | VARCHAR(45)     | 是   | 主机IP地址(可变，Agent上报更新，支持IPv6)      |
| os\_type    | VARCHAR(32)     | 否   | 操作系统类型: linux/windows                    |
| name        | VARCHAR(128)    | 是   | 进程名称                                       |
| status      | VARCHAR(64)     |      | 进程状态                                       |
| version     | VARCHAR(64)     |      | 进程版本                                       |
| path        | VARCHAR(512)    | 是   | 进程路径                                       |
| run\_name   | VARCHAR(128)    | 是   | 运行用户                                       |
| start\_time | DATETIME        | 否   | 进程启动时间                                   |
| created\_at | DATETIME        | 是   | 创建时间                                       |
| updated\_at | DATETIME        | 是   | 更新时间                                       |

#### 

#### **5、数据库列表设计：**

| 字段        | 类型            | 必填 | 说明                                           |
| ----------- | --------------- | ---- | ---------------------------------------------- |
| id          | BIGINT UNSIGNED | 是   | 主键ID                                         |
| agent\_id   | VARCHAR(64)     | 是   | Agent唯一标识(安装时生成，不随IP/hostname变化) |
| host\_name  | VARCHAR(128)    | 是   | 主机名称(可变，Agent上报更新)                  |
| host\_ip    | VARCHAR(45)     | 是   | 主机IP地址(可变，Agent上报更新，支持IPv6)      |
| os\_type    | VARCHAR(32)     | 否   | 操作系统类型: linux/windows                    |
| db\_type    | VARCHAR(45)     | 是   | 数据库类型                                     |
| db\_version | VARCHAR(45)     | 是   | 数据库版本                                     |
| port        | INT             | 是   | 监听端口                                       |
| run\_user   | VARCHAR(64)     | 否   | 运行用户                                       |
| created\_at | DATETIME        | 是   | 创建时间                                       |
| updated\_at | DATETIME        | 是   | 更新时间                                       |

#### **6、Web服务数据库设计：**

| 字段         | 类型            | 必填 | 说明                                           |
| ------------ | --------------- | ---- | ---------------------------------------------- |
| id           | BIGINT UNSIGNED | 是   | 主键ID                                         |
| agent\_id    | VARCHAR(64)     | 是   | Agent唯一标识(安装时生成，不随IP/hostname变化) |
| host\_name   | VARCHAR(128)    | 是   | 主机名称(可变，Agent上报更新)                  |
| host\_ip     | VARCHAR(45)     | 是   | 主机IP地址(可变，Agent上报更新，支持IPv6)      |
| os\_type     | VARCHAR(32)     | 否   | 操作系统类型: linux/windows                    |
| name         | VARCHAR(128)    | 是   | 应用名                                         |
| version      | VARCHAR(64)     | 是   | 版本                                           |
| server\_type | VARCHAR(64)     | 是   | 服务器类型                                     |
| site\_domain | VARCHAR(255)    | 否   | 站点域名                                       |
| path         | VARCHAR(512)    | 否   | 根路径                                         |
| created\_at  | DATETIME        | 是   | 创建时间                                       |
| updated\_at  | DATETIME        | 是   | 更新时间                                       |

#### **7、系统服务数据库设计：**

| 字段        | 类型            | 必填 | 说明                                           |
| ----------- | --------------- | ---- | ---------------------------------------------- |
| id          | BIGINT UNSIGNED | 是   | 主键ID                                         |
| agent\_id   | VARCHAR(64)     | 是   | Agent唯一标识(安装时生成，不随IP/hostname变化) |
| host\_name  | VARCHAR(128)    | 是   | 主机名称(可变，Agent上报更新)                  |
| host\_ip    | VARCHAR(45)     | 是   | 主机IP地址(可变，Agent上报更新，支持IPv6)      |
| name        | VARCHAR(255)    | 是   | 服务名称                                       |
| version     | VARCHAR(64)     | 否   | 版本                                           |
| status      | VARCHAR(64)     | 是   | 状态                                           |
| run\_user   | VARCHAR(255)    | 是   | 运行用户                                       |
| path        | VARCHAR(512)    | 是   | 根路径                                         |
| describe    | VARCHAR(512)    | 否   | 描述                                           |
| created\_at | DATETIME        | 是   | 创建时间                                       |
| updated\_at | DATETIME        | 是   | 更新时间                                       |

#### **8、软件列表数据库设计：**

| 字段        | 类型            | 必填 | 说明                                           |
| ----------- | --------------- | ---- | ---------------------------------------------- |
| id          | BIGINT UNSIGNED | 是   | 主键ID                                         |
| agent\_id   | VARCHAR(64)     | 是   | Agent唯一标识(安装时生成，不随IP/hostname变化) |
| host\_name  | VARCHAR(128)    | 是   | 主机名称(可变，Agent上报更新)                  |
| host\_ip    | VARCHAR(45)     | 是   | 主机IP地址(可变，Agent上报更新，支持IPv6)      |
| name        | VARCHAR(255)    | 是   | 软件名称                                       |
| version     | VARCHAR(128)    | 否   | 软件版本                                       |
| type        | VARCHAR(32)     | 是   | 软件类型: dpkg, rpm, pypi, jar                 |
| source      | VARCHAR(255)    | 否   | 来源                                           |
| status      | VARCHAR(64)     | 否   | 状态                                           |
| vendor      | VARCHAR(255)    | 否   | 厂商                                           |
| path        | VARCHAR(512)    | 否   | 路径(jar类型)                                  |
| created\_at | DATETIME        | 是   | 创建时间                                       |
| updated\_at | DATETIME        | 是   | 更新时间                                       |

#### **9、容器列表数据库设计：**

| 字段         | 类型            | 必填 | 说明                                           |
| ------------ | --------------- | ---- | ---------------------------------------------- |
| id           | BIGINT UNSIGNED | 是   | 主键ID                                         |
| agent\_id    | VARCHAR(64)     | 是   | Agent唯一标识(安装时生成，不随IP/hostname变化) |
| host\_name   | VARCHAR(128)    | 是   | 主机名称(可变，Agent上报更新)                  |
| host\_ip     | VARCHAR(45)     | 是   | 主机IP地址(可变，Agent上报更新，支持IPv6)      |
| container\_id | VARCHAR(128)    | 是   | 容器ID                                         |
| name         | VARCHAR(255)    | 是   | 容器名称                                       |
| state        | VARCHAR(32)     | 是   | 容器状态                                       |
| image\_id    | VARCHAR(128)    | 否   | 镜像ID                                         |
| image\_name  | VARCHAR(255)    | 否   | 镜像名称                                       |
| runtime      | VARCHAR(32)     | 否   | 运行时(docker/containerd)                      |
| pid          | VARCHAR(16)     | 否   | 容器主进程PID                                  |
| create\_time | VARCHAR(32)     | 否   | 容器创建时间                                   |
| created\_at  | DATETIME        | 是   | 创建时间                                       |
| updated\_at  | DATETIME        | 是   | 更新时间                                       |

#### **10、可疑环境变量数据库设计：**

| 字段                  | 类型            | 必填 | 说明                                           |
| --------------------- | --------------- | ---- | ---------------------------------------------- |
| id                    | BIGINT UNSIGNED | 是   | 主键ID                                         |
| agent\_id             | VARCHAR(64)     | 是   | Agent唯一标识(安装时生成，不随IP/hostname变化) |
| host\_name            | VARCHAR(128)    | 是   | 主机名称(可变，Agent上报更新)                  |
| host\_ip              | VARCHAR(45)     | 是   | 主机IP地址(可变，Agent上报更新，支持IPv6)      |
| var\_name             | VARCHAR(255)    | 是   | 环境变量名                                     |
| var\_value            | TEXT            | 否   | 环境变量值                                     |
| suspicious\_reasons   | TEXT            | 否   | 可疑原因                                       |
| source                | VARCHAR(128)    | 否   | 来源                                           |
| created\_at           | DATETIME        | 是   | 创建时间                                       |
| updated\_at           | DATETIME        | 是   | 更新时间                                       |

#### **11、内核模块数据库设计：**

| 字段        | 类型            | 必填 | 说明                                           |
| ----------- | --------------- | ---- | ---------------------------------------------- |
| id          | BIGINT UNSIGNED | 是   | 主键ID                                         |
| agent\_id   | VARCHAR(64)     | 是   | Agent唯一标识(安装时生成，不随IP/hostname变化) |
| host\_name  | VARCHAR(128)    | 是   | 主机名称(可变，Agent上报更新)                  |
| host\_ip    | VARCHAR(45)     | 是   | 主机IP地址(可变，Agent上报更新，支持IPv6)      |
| name        | VARCHAR(128)    | 是   | 模块名称                                       |
| size        | VARCHAR(32)     | 否   | 模块大小                                       |
| refcount    | VARCHAR(16)     | 否   | 引用计数                                       |
| used\_by    | VARCHAR(512)    | 否   | 使用该模块的模块列表                           |
| state       | VARCHAR(32)     | 否   | 模块状态                                       |
| addr        | VARCHAR(32)     | 否   | 内存地址                                       |
| created\_at | DATETIME        | 是   | 创建时间                                       |
| updated\_at | DATETIME        | 是   | 更新时间                                       |


### **2、4、3 入侵检测数据库设计**

#### **1、高危命令告警表 (alert\_dangerous\_command)**

记录检测到的危险命令执行行为。

| 字段 | 类型 | 必填 | 说明 |
| ----- | ----- | ----- | ----- |
| id | BIGINT UNSIGNED | 是 | 主键ID |
| agent\_id | VARCHAR(64) | 是 | Agent唯一标识 |
| host\_id | BIGINT UNSIGNED | 否 | 关联主机ID |
| host\_name | VARCHAR(128) | 是 | 主机名 |
| host\_ip | VARCHAR(45) | 是 | 主机IP |
| command | TEXT | 是 | 执行的命令内容 |
| command\_type | VARCHAR(32) | 是 | 命令类型 |
| user | VARCHAR(64) | 是 | 执行用户 |
| privilege\_level | VARCHAR(32) | 是 | 权限级别 |
| status | TINYINT | 是 | 处理状态: 0-待处理 1-已处理 2-已忽略 |
| alert\_time | DATETIME | 是 | 告警时间 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |

**命令类型枚举 (command\_type)：**

| 值 | 说明 |
| ----- | ----- |
| file\_delete | 文件删除 |
| privilege\_escalation | 权限提升 |
| permission\_modify | 权限修改 |
| filesystem\_operation | 文件系统操作 |
| network\_scan | 网络扫描 |
| data\_exfiltration | 数据外传 |
| service\_stop | 服务停止 |
| log\_tamper | 日志篡改 |

#### **2、反弹Shell告警表 (alert\_reverse\_shell)**

记录检测到的反向Shell连接事件。

| 字段 | 类型 | 必填 | 说明 |
| ----- | ----- | ----- | ----- |
| id | BIGINT UNSIGNED | 是 | 主键ID |
| agent\_id | VARCHAR(64) | 是 | Agent唯一标识 |
| host\_id | BIGINT UNSIGNED | 否 | 关联主机ID |
| host\_name | VARCHAR(128) | 是 | 主机名称 |
| victim\_ip | VARCHAR(45) | 是 | 受害主机IP |
| command\_line | TEXT | 是 | 反弹Shell命令行 |
| shell\_type | VARCHAR(32) | 否 | Shell类型 |
| target\_host | VARCHAR(45) | 是 | 目标主机(攻击者IP) |
| target\_port | INT | 是 | 目标端口 |
| status | TINYINT | 是 | 处理状态: 0-待处理 1-已处理 2-已忽略 |
| event\_time | DATETIME | 是 | 事件时间 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |

**Shell类型枚举 (shell\_type)：**

| 值 | 说明 |
| ----- | ----- |
| bash | Bash Shell |
| python | Python |
| nc | Netcat |
| perl | Perl |
| php | PHP |
| ruby | Ruby |
| powershell | PowerShell |

#### **3、本地提权告警表 (alert\_privilege\_escalation)**

记录检测到的本地权限提升事件。

| 字段 | 类型 | 必填 | 说明 |
| ----- | ----- | ----- | ----- |
| id | BIGINT UNSIGNED | 是 | 主键ID |
| agent\_id | VARCHAR(64) | 是 | Agent唯一标识 |
| host\_id | BIGINT UNSIGNED | 否 | 关联主机ID |
| host\_name | VARCHAR(128) | 是 | 主机名称 |
| host\_ip | VARCHAR(45) | 是 | 主机IP |
| escalated\_user | VARCHAR(64) | 是 | 提权后用户 |
| parent\_process | VARCHAR(256) | 是 | 父进程名称 |
| parent\_process\_user | VARCHAR(64) | 是 | 父进程所属用户 |
| process\_id | INT | 否 | 进程ID |
| process\_path | VARCHAR(512) | 否 | 进程路径 |
| status | TINYINT | 是 | 处理状态: 0-待处理 1-已处理 2-已忽略 |
| discover\_time | DATETIME | 是 | 发现时间 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |

#### **4、异常登录告警表 (alert\_abnormal\_login)**

记录检测到的异常登录事件。

| 字段 | 类型 | 必填 | 说明 |
| ----- | ----- | ----- | ----- |
| id | BIGINT UNSIGNED | 是 | 主键ID |
| agent\_id | VARCHAR(64) | 是 | Agent唯一标识 |
| host\_id | BIGINT UNSIGNED | 否 | 关联主机ID |
| host\_name | VARCHAR(128) | 是 | 主机名称 |
| host\_ip | VARCHAR(45) | 是 | 主机IP |
| source\_ip | VARCHAR(45) | 是 | 来源IP |
| source\_location | VARCHAR(128) | 否 | 来源地理位置 |
| source\_country | VARCHAR(64) | 否 | 来源国家 |
| source\_city | VARCHAR(64) | 否 | 来源城市 |
| login\_user | VARCHAR(64) | 是 | 登录用户名 |
| login\_time | DATETIME | 是 | 登录时间 |
| risk\_level | VARCHAR(16) | 是 | 危险等级 |
| abnormal\_type | VARCHAR(32) | 否 | 异常类型 |
| status | TINYINT | 是 | 处理状态: 0-待处理 1-已处理 2-已忽略 |
| is\_whitelist | TINYINT | 否 | 是否白名单: 0-否 1-是 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |

**异常类型枚举 (abnormal\_type)：**

| 值 | 说明 |
| ----- | ----- |
| abnormal\_location | 异常地域 |
| abnormal\_time | 异常时间 |
| abnormal\_user | 异常用户 |

#### **5、密码破解告警表 (alert\_brute\_force)**

记录检测到的暴力破解攻击。

| 字段 | 类型 | 必填 | 说明 |
| :---- | :---- | :---- | :---- |
| id | BIGINT UNSIGNED | 是 | 主键ID |
| agent\_id | VARCHAR(64) | 是 | Agent唯一标识 |
| host\_id | BIGINT UNSIGNED | 否 | 关联主机ID |
| host\_name | VARCHAR(128) | 是 | 主机名称 |
| host\_ip | VARCHAR(45) | 是 | 主机IP |
| source\_ip | VARCHAR(45) | 是 | 攻击来源IP |
| source\_location | VARCHAR(128) | 否 | 来源地理位置 |
| attack\_type | VARCHAR(32) | 是 | 攻击类型 |
| target\_ip | VARCHAR(45) | 是 | 目标IP |
| target\_port | INT | 否 | 目标端口 |
| username | VARCHAR(64) | 是 | 被尝试的用户名 |
| attempt\_count | INT | 是 | 尝试次数 |
| attack\_time | DATETIME | 是 | 攻击时间(最近一次) |
| first\_attack\_time | DATETIME | 否 | 首次攻击时间 |
| status | TINYINT | 是 | 处理状态: 0-待处理 1-已处理 2-已忽略 |
| is\_blocked | TINYINT | 否 | 是否已封禁: 0-否 1-是 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |

攻击类型枚举 (attack\_type)：

| 值 | 说明 |
| :---- | :---- |
| ssh | SSH密码暴力破解 |
| rdp | RDP远程桌面破解 |
| ftp | FTP暴力破解 |
| mysql | MySQL暴力破解 |
| redis | Redis未授权访问 |
| web\_login | Web登录暴力破解 |

#### **6、恶意请求告警表 (alert\_malicious\_request)**

记录检测到的恶意域名/IP访问请求。

| 字段 | 类型 | 必填 | 说明 |
| :---- | :---- | :---- | :---- |
| id | BIGINT UNSIGNED | 是 | 主键ID |
| agent\_id | VARCHAR(64) | 是 | Agent唯一标识 |
| host\_id | BIGINT UNSIGNED | 否 | 关联主机ID |
| host\_name | VARCHAR(128) | 是 | 主机名称 |
| host\_ip | VARCHAR(45) | 是 | 主机IP |
| policy\_type | VARCHAR(32) | 是 | 命中策略类型 |
| policy\_name | VARCHAR(128) | 是 | 命中策略名称 |
| malicious\_domain | VARCHAR(256) | 是 | 恶意请求域名 |
| malicious\_ip | VARCHAR(45) | 否 | 恶意请求IP |
| request\_count | INT | 是 | 请求次数 |
| first\_request\_time | DATETIME | 否 | 首次请求时间 |
| last\_request\_time | DATETIME | 否 | 最近请求时间 |
| risk\_description | TEXT | 否 | 危害描述 |
| status | TINYINT | 是 | 处理状态: 0-待处理 1-已处理 2-已忽略 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |

策略类型枚举 (policy\_type)：

| 值 | 说明 |
| :---- | :---- |
| mining | 挖矿 |
| c2 | C2通信 |
| phishing | 钓鱼网站 |
| botnet | 僵尸网络 |
| ransomware | 勒索软件 |

#### **7、网络攻击告警表 (alert\_network\_attack)**

记录检测到的漏洞利用攻击。

| 字段 | 类型 | 必填 | 说明 |
| :---- | :---- | :---- | :---- |
| id | BIGINT UNSIGNED | 是 | 主键ID |
| agent\_id | VARCHAR(64) | 是 | Agent唯一标识 |
| host\_id | BIGINT UNSIGNED | 否 | 关联主机ID |
| host\_name | VARCHAR(128) | 是 | 主机名称 |
| host\_ip | VARCHAR(45) | 是 | 被攻击主机IP |
| target\_port | INT | 是 | 目标端口 |
| attacker\_ip | VARCHAR(45) | 是 | 攻击来源IP |
| attacker\_location | VARCHAR(128) | 否 | 攻击来源地理位置 |
| attacker\_country | VARCHAR(64) | 否 | 攻击来源国家 |
| vulnerability\_name | VARCHAR(256) | 是 | 漏洞名称 |
| vulnerability\_id | VARCHAR(64) | 否 | 漏洞编号(CVE等) |
| attack\_status | VARCHAR(32) | 是 | 攻击状态 |
| attack\_count | INT | 是 | 攻击次数 |
| first\_attack\_time | DATETIME | 否 | 首次攻击时间 |
| last\_attack\_time | DATETIME | 是 | 最近攻击时间 |
| attack\_payload | TEXT | 否 | 攻击载荷 |
| status | TINYINT | 是 | 处理状态: 0-待处理 1-已处理 2-已忽略 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |

#### **8、文件查杀告警表 (alert\_malware\_scan)**

**记录检测到的恶意文件。**

| 字段 | 类型 | 必填 | 说明 |
| :---- | :---- | :---- | :---- |
| id | BIGINT UNSIGNED | 是 | 主键ID |
| agent\_id | VARCHAR(64) | 是 | Agent唯一标识 |
| host\_id | BIGINT UNSIGNED | 否 | 关联主机ID |
| host\_ip | VARCHAR(45) | 是 | 主机IP |
| host\_name | VARCHAR(128) | 是 | 主机名称 |
| threat\_type | VARCHAR(64) | 是 | 威胁类型 |
| file\_name | VARCHAR(256) | 是 | 文件名 |
| file\_path | VARCHAR(512) | 是 | 文件路径 |
| file\_size | BIGINT | 否 | 文件大小(字节) |
| file\_md5 | VARCHAR(32) | 否 | 文件MD5哈希 |
| file\_sha256 | VARCHAR(64) | 否 | 文件SHA256哈希 |
| detection\_engine | VARCHAR(64) | 否 | 检测引擎 |
| malware\_family | VARCHAR(64) | 否 | 恶意软件家族 |
| is\_quarantined | TINYINT | 否 | 是否已隔离: 0-否 1-是 |
| is\_deleted | TINYINT | 否 | 是否已删除: 0-否 1-是 |
| status | TINYINT | 是 | 处理状态: 0-待处理 1-已处理 2-已忽略 |
| scan\_time | DATETIME | 是 | 扫描时间 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |

**威胁类型枚举 (threat\_type)：**

| 值 | 说明 |
| ----- | ----- |
| virus | 病毒程序 |
| trojan | 木马程序 |
| webshell | Webshell |
| backdoor | 后门程序 |
| ransomware | 勒索软件 |
| miner | 挖矿程序 |
| rootkit | Rootkit |

#### **9、核心文件监控告警表 (alert\_file\_integrity)**

记录检测到的关键文件变更事件。

| 字段 | 类型 | 必填 | 说明 |
| :---- | :---- | :---- | :---- |
| id | BIGINT UNSIGNED | 是 | 主键ID |
| agent\_id | VARCHAR(64) | 是 | Agent唯一标识 |
| host\_id | BIGINT UNSIGNED | 否 | 关联主机ID |
| host\_name | VARCHAR(128) | 是 | 主机名称 |
| host\_ip | VARCHAR(45) | 是 | 主机IP |
| rule\_type | VARCHAR(32) | 是 | 规则类型 |
| rule\_name | VARCHAR(128) | 是 | 命中规则名称 |
| rule\_id | BIGINT UNSIGNED | 否 | 关联规则ID |
| threat\_level | VARCHAR(16) | 是 | 威胁等级 |
| threat\_action | VARCHAR(32) | 是 | 威胁行为 |
| file\_path | VARCHAR(512) | 是 | 文件路径 |
| file\_name | VARCHAR(256) | 否 | 文件名 |
| old\_content\_hash | VARCHAR(64) | 否 | 原内容哈希 |
| new\_content\_hash | VARCHAR(64) | 否 | 新内容哈希 |
| change\_detail | TEXT | 否 | 变更详情 |
| operator\_user | VARCHAR(64) | 否 | 操作用户 |
| operator\_process | VARCHAR(256) | 否 | 操作进程 |
| alert\_description | TEXT | 否 | 告警描述 |
| status | TINYINT | 是 | 处理状态: 0-待处理 1-已处理 2-已忽略 |
| alert\_time | DATETIME | 是 | 告警时间 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |

#### **10、告警处理记录表 (alert\_process\_log)**

记录告警的处理历史。

| 字段 | 类型 | 必填 | 说明 |
| :---- | :---- | :---- | :---- |
| id | BIGINT UNSIGNED | 是 | 主键ID |
| alert\_type | VARCHAR(32) | 是 | 告警类型 |
| alert\_id | BIGINT UNSIGNED | 是 | 关联的告警ID |
| old\_status | TINYINT | 否 | 变更前状态 |
| new\_status | TINYINT | 是 | 变更后状态 |
| processor | VARCHAR(64) | 是 | 处理人 |
| remark | VARCHAR(512) | 否 | 处理备注 |
| created\_at | DATETIME | 是 | 创建时间 |

**告警类型枚举 (alert\_type)：**

| 值 | 说明 |
| :---- | :---- |
| dangerous\_command | 高危命令 |
| reverse\_shell | 反弹Shell |
| privilege\_escalation | 本地提权 |
| abnormal\_login | 异常登录 |
| brute\_force | 密码破解 |
| malicious\_request | 恶意请求 |
| network\_attack | 网络攻击 |
| malware\_scan | 文件查杀 |
| file\_integrity | 核心文件监控 |

**处理状态枚举 (status)：**

| 值 | 说明 |
| :---- | :---- |
| 0 | 待处理 |
| 1 | 已处理 |
| 2 | 已忽略 |

### **2、4、4 合规基线数据库设计**

#### **1、 基线模板表 (baseline\_template)**

定义安全基线标准。

| 字段 | 类型 | 必填 | 说明 |
| ----- | ----- | ----- | ----- |
| id | BIGINT UNSIGNED | 是 | 主键ID |
| baseline\_name | VARCHAR(128) | 是 | 基线名称，如"CIS CentOS Linux 7 Benchmark" |
| baseline\_type | VARCHAR(32) | 是 | 基线类型 |
| os\_type | VARCHAR(32) | 否 | 适用操作系统类型: linux/windows |
| version | VARCHAR(32) | 否 | 基线版本号 |
| item\_count | INT | 否 | 检查项总数 |
| description | VARCHAR(512) | 否 | 基线描述 |
| is\_enabled | TINYINT | 否 | 是否启用：0-否 1-是 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |

**基线类型枚举 (baseline\_type)：**

| 值 | 说明 |
| ----- | ----- |
| cis | 标准基线 |
| custom | 自定义基线 |

---

#### **2、基线检查项表 (baseline\_check\_item)**

存储每个基线包含的具体检查项。

| 字段 | 类型 | 必填 | 说明 |
| ----- | ----- | ----- | ----- |
| id | BIGINT UNSIGNED | 是 | 主键ID |
| baseline\_id | BIGINT UNSIGNED | 是 | 关联基线ID |
| item\_name | VARCHAR(256) | 是 | 检查项名称 |
| category | VARCHAR(64) | 是 | 检查项分类 |
| risk\_level | VARCHAR(16) | 是 | 风险等级 |
| check\_type | VARCHAR(32) | 否 | 检查类型: command/file/config |
| expected\_value | VARCHAR(256) | 否 | 期望值 |
| fix\_suggestion | TEXT | 否 | 修复建议 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |

**风险等级枚举 (risk\_level)：**

| 值 | 说明 |
| ----- | ----- |
| high | 高危 |
| medium | 中危 |
| low | 低危 |

---

#### **3、检查结果表 (baseline\_check\_result)**

存储主机检查结果汇总。

| 字段 | 类型 | 必填 | 说明 |
| ----- | ----- | ----- | ----- |
| id | BIGINT UNSIGNED | 是 | 主键ID |
| baseline\_id | BIGINT UNSIGNED | 是 | 基线ID |
| agent\_id | VARCHAR(64) | 是 | Agent唯一标识 |
| host\_ip | VARCHAR(45) | 是 | 主机IP |
| host\_name | VARCHAR(128) | 否 | 主机名 |
| total\_items | INT | 是 | 检查项总数 |
| passed\_items | INT | 是 | 通过项数 |
| failed\_items | INT | 是 | 未通过项数 |
| check\_time | DATETIME | 是 | 检查时间 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |

---

#### **4、检查明细表 (baseline\_check\_detail)**

存储每个检查项的具体检查结果。

| 字段 | 类型 | 必填 | 说明 |
| ----- | ----- | ----- | ----- |
| id | BIGINT UNSIGNED | 是 | 主键ID |
| result\_id | BIGINT UNSIGNED | 是 | 关联检查结果ID |
| item\_id | BIGINT UNSIGNED | 是 | 关联检查项ID |
| agent\_id | VARCHAR(64) | 是 | Agent唯一标识 |
| status | TINYINT | 是 | 检查状态: 0-未通过 1-通过 2-跳过 |
| actual\_value | TEXT | 否 | 实际检测值 |
| expected\_value | TEXT | 否 | 期望值 |
| error\_message | VARCHAR(512) | 否 | 错误信息 |
| check\_time | DATETIME | 是 | 检查时间 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |

**检查状态枚举 (status)：**

| 值 | 说明 |
| ----- | ----- |
| 0 | 未通过 |
| 1 | 通过 |
| 2 | 跳过(不适用) |

### **2、4、5 漏洞发现数据库设计**

#### **1）主机漏洞**

**1、主机漏洞扫描结果表 (host\_vuln\_scan)**

存储每台主机的漏洞扫描汇总信息，对应**主机视图**。

| 字段名 | 类型 | 必填 | 说明 |
| ----- | ----- | ----- | ----- |
| id | BIGINT UNSIGNED | 是 | 主键 |
| agent\_id | VARCHAR(64) | 是 | Agent唯一标识 |
| host\_id | BIGINT UNSIGNED | 否 | 关联主机ID |
| host\_name | VARCHAR(128) | 是 | 主机名称 |
| host\_ip | VARCHAR(45) | 是 | 主机IP |
| critical\_count | INT | 否 | 严重漏洞数 |
| high\_count | INT | 否 | 高危漏洞数 |
| medium\_count | INT | 否 | 中危漏洞数 |
| low\_count | INT | 否 | 低危漏洞数 |
| scan\_time | DATETIME | 是 | 最近扫描时间 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |

---

**1.2 漏洞信息表 (vuln\_info)**

存储CVE漏洞基础信息，对应**漏洞视图**。

| 字段名 | 类型 | 必填 | 说明 |
| ----- | ----- | ----- | ----- |
| id | BIGINT UNSIGNED | 是 | 主键 |
| cve\_id | VARCHAR(32) | 是 | CVE编号 |
| vuln\_name | VARCHAR(256) | 是 | 漏洞名称 |
| severity | VARCHAR(16) | 是 | 漏洞等级(critical/high/medium/low) |
| cvss\_score | DECIMAL(3,1) | 否 | CVSS评分 |
| affected\_host\_count | INT | 否 | 影响主机数 |
| affected\_image\_count | INT | 否 | 影响镜像数 |
| description | TEXT | 否 | 漏洞描述 |
| fix\_suggestion | TEXT | 否 | 修复建议 |
| reference\_urls | TEXT | 否 | 参考链接 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |

---

**1.3 主机漏洞关联表 (host\_vuln\_detail)**

存储主机与漏洞的关联关系。

| 字段名 | 类型 | 必填 | 说明 |
| ----- | ----- | ----- | ----- |
| id | BIGINT UNSIGNED | 是 | 主键 |
| agent\_id | VARCHAR(64) | 是 | Agent唯一标识 |
| host\_id | BIGINT UNSIGNED | 否 | 关联主机ID |
| vuln\_id | BIGINT UNSIGNED | 是 | 漏洞ID |
| cve\_id | VARCHAR(32) | 是 | CVE编号(冗余) |
| package\_name | VARCHAR(128) | 是 | 受影响软件包 |
| installed\_version | VARCHAR(64) | 否 | 当前版本 |
| fixed\_version | VARCHAR(64) | 否 | 修复版本 |
| status | TINYINT | 是 | 状态: 0-未修复 1-已修复 2-已忽略 |
| scan\_time | DATETIME | 是 | 扫描时间 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |

#### **2）容器漏洞**

**2.1 镜像漏洞扫描结果表 (image\_vuln\_scan)**

存储每个镜像的漏洞扫描汇总信息，对应**镜像视图**。

| 字段名 | 类型 | 必填 | 说明 |
| ----- | ----- | ----- | ----- |
| id | BIGINT UNSIGNED | 是 | 主键 |
| agent\_id | VARCHAR(64) | 是 | Agent唯一标识(镜像所在主机) |
| image\_id | VARCHAR(128) | 是 | 镜像ID |
| image\_name | VARCHAR(256) | 是 | 镜像名称(包含tag标签) |
| critical\_count | INT | 否 | 严重漏洞数 |
| high\_count | INT | 否 | 高危漏洞数 |
| medium\_count | INT | 否 | 中危漏洞数 |
| low\_count | INT | 否 | 低危漏洞数 |
| scan\_time | DATETIME | 是 | 最近扫描时间 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |

---

**2.2 漏洞信息表**

> **说明**：容器漏洞与主机漏洞共用同一张漏洞信息表 `vuln_info`（见1.2节），通过 `affected_host_count` 和 `affected_image_count` 字段分别记录影响的主机数和镜像数。

---

**2.3 镜像漏洞关联表 (image\_vuln\_detail)**

存储镜像与漏洞的关联关系。

| 字段名 | 类型 | 必填 | 说明 |
| ----- | ----- | ----- | ----- |
| id | BIGINT UNSIGNED | 是 | 主键 |
| agent\_id | VARCHAR(64) | 是 | Agent唯一标识(镜像所在主机) |
| image\_id | VARCHAR(128) | 是 | 镜像ID |
| vuln\_id | BIGINT UNSIGNED | 是 | 漏洞ID |
| cve\_id | VARCHAR(32) | 是 | CVE编号(冗余) |
| package\_name | VARCHAR(128) | 是 | 受影响软件包 |
| installed\_version | VARCHAR(64) | 否 | 当前版本 |
| fixed\_version | VARCHAR(64) | 否 | 修复版本 |
| status | TINYINT | 是 | 状态: 0-未修复 1-已修复 2-已忽略 |
| scan\_time | DATETIME | 是 | 扫描时间 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |

### **2、4、6 代码安全数据库设计**

#### **1、代码扫描结果表 (code\_scan\_result)**

存储代码仓库扫描汇总结果。

| 字段名 | 类型 | 必填 | 说明 |
| ----- | ----- | ----- | ----- |
| id | BIGINT UNSIGNED | 是 | 主键 |
| repo\_id | BIGINT UNSIGNED | 是 | 关联仓库ID |
| repo\_name | VARCHAR(100) | 是 | 仓库名称 |
| repo\_type | VARCHAR(50) | 否 | 仓库类型（如 Git、GitHub 等） |
| rule\_set\_id | BIGINT UNSIGNED | 否 | 规则集 ID |
| total\_vulnerabilities | INT | 否 | 漏洞总数 |
| critical\_count | INT | 否 | 严重漏洞数量 |
| high\_count | INT | 否 | 高危漏洞数量 |
| medium\_count | INT | 否 | 中危漏洞数量 |
| low\_count | INT | 否 | 低危漏洞数量 |
| scan\_start\_time | DATETIME | 是 | 扫描开始时间 |
| scan\_end\_time | DATETIME | 否 | 扫描结束时间 |
| highest\_risk\_level | VARCHAR(16) | 否 | 最高风险等级: LOW/MEDIUM/HIGH/CRITICAL |
| repo\_status | VARCHAR(16) | 否 | 仓库安全状态: SAFE/WARNING/DANGEROUS |
| scan\_report\_url | VARCHAR(500) | 否 | 扫描报告下载链接 |
| deleted | TINYINT | 否 | 逻辑删除标志: 0-未删除 1-已删除 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |



#### **2、代码漏洞明细表 (code\_scan\_vuln)**

存储扫描发现的具体漏洞明细。

| 字段名 | 类型 | 必填 | 说明 |
| ----- | ----- | ----- | ----- |
| id | BIGINT UNSIGNED | 是 | 主键 |
| result\_id | BIGINT UNSIGNED | 是 | 关联扫描结果ID |
| rule\_id | BIGINT UNSIGNED | 否 | 关联规则ID |
| rule\_key | VARCHAR(100) | 是 | 规则唯一标识 |
| severity | VARCHAR(16) | 是 | 严重等级: LOW/MEDIUM/HIGH/CRITICAL |
| file\_path | VARCHAR(500) | 是 | 漏洞所在文件路径 |
| line\_number | INT | 否 | 漏洞所在行号 |
| code\_snippet | TEXT | 否 | 问题代码片段 |
| message | VARCHAR(500) | 是 | 漏洞简要描述 |
| remediation | TEXT | 否 | 修复建议 |
| status | VARCHAR(16) | 否 | 处理状态: NEW/IGNORED/FIXED/WONT\_FIX |
| ignore\_reason | VARCHAR(255) | 否 | 忽略原因 |
| hash | CHAR(64) | 否 | 漏洞指纹，用于去重 |
| deleted | TINYINT | 否 | 逻辑删除标志: 0-未删除 1-已删除 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |

#### **3、代码仓库表 (code\_repository)**

存储代码仓库基本信息。

| 字段名 | 类型 | 必填 | 说明 |
| ----- | ----- | ----- | ----- |
| id | BIGINT UNSIGNED | 是 | 主键 |
| repo\_name | VARCHAR(100) | 是 | 仓库名称 |
| repo\_url | VARCHAR(500) | 是 | 仓库地址 |
| pull\_method | VARCHAR(16) | 否 | 拉取方式: SSH/HTTPS |
| is\_private | TINYINT | 否 | 是否私有仓库: 0-公开 1-私有 |
| description | VARCHAR(500) | 否 | 描述信息 |
| code\_hash | CHAR(64) | 否 | 代码哈希值（用于判断是否变更） |
| owner | VARCHAR(50) | 是 | 负责人 |
| branch | VARCHAR(100) | 否 | 代码分支（如 main、develop） |
| local\_path | VARCHAR(500) | 否 | 本地代码路径 |
| scan\_frequency | VARCHAR(16) | 否 | 扫描频率: DAILY/WEEKLY/MONTHLY/MANUAL |
| last\_scan\_time | DATETIME | 否 | 最后扫描时间 |
| status | VARCHAR(16) | 否 | 状态: PENDING/SCANNING/SUCCESS/FAILED |
| deleted | TINYINT | 否 | 逻辑删除标志: 0-未删除 1-已删除 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |

#### **4、规则集表 (code\_rule\_set)**

存储代码扫描规则集。

| 字段名 | 类型 | 必填 | 说明 |
| ----- | ----- | ----- | ----- |
| id | BIGINT UNSIGNED | 是 | 主键 |
| rule\_set\_name | VARCHAR(100) | 是 | 规则集名称 |
| applicable\_scene | VARCHAR(100) | 否 | 适用场景 |
| risk\_coverage | VARCHAR(100) | 否 | 风险覆盖范围 |
| total\_rules | INT | 否 | 关联规则数 |
| description | TEXT | 否 | 规则集描述 |
| status | VARCHAR(16) | 否 | 启用状态: ENABLED/DISABLED |
| deleted | TINYINT | 否 | 逻辑删除标志: 0-未删除 1-已删除 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |

#### **5、规则表 (code\_rule)**

存储具体的扫描规则。

| 字段名 | 类型 | 必填 | 说明 |
| ----- | ----- | ----- | ----- |
| id | BIGINT UNSIGNED | 是 | 主键 |
| rule\_set\_id | BIGINT UNSIGNED | 是 | 关联规则集ID |
| check\_item | VARCHAR(100) | 是 | 检查项（如"SQL注入检测"） |
| rule\_key | VARCHAR(100) | 是 | 唯一标识（如 SQL\_INJECTION） |
| severity | VARCHAR(16) | 否 | 严重等级: LOW/MEDIUM/HIGH/CRITICAL |
| category | VARCHAR(50) | 否 | 规则类别（如安全、语法、合规） |
| description | TEXT | 否 | 规则详细说明 |
| status | VARCHAR(16) | 否 | 状态: ACTIVE/DEPRECATED/EXPERIMENTAL |
| deleted | TINYINT | 否 | 逻辑删除标志: 0-未删除 1-已删除 |
| created\_at | DATETIME | 是 | 创建时间 |
| updated\_at | DATETIME | 是 | 更新时间 |

**规则状态枚举 (status)：**

| 值 | 说明 |
| ----- | ----- |
| ACTIVE | 活跃（启用中） |
| DEPRECATED | 已弃用 |
| EXPERIMENTAL | 实验性 |

