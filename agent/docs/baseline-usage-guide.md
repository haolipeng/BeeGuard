# 基线检测插件使用指南

## 目录

- [配置文件结构](#配置文件结构)
- [检查项字段说明](#检查项字段说明)
- [规则配置详解](#规则配置详解)
  - [规则类型 (rules.type)](#规则类型-rulestype)
  - [规则参数 (rules.param)](#规则参数-rulesparam)
  - [过滤器 (rules.filter)](#过滤器-rulesfilter)
  - [前提条件 (rules.require)](#前提条件-rulesrequire)
  - [结果匹配 (rules.result)](#结果匹配-rulesresult)
  - [条件逻辑 (check.condition)](#条件逻辑-checkcondition)
- [各规则类型配置示例](#各规则类型配置示例)

---

## 配置文件结构

基线规则通过 YAML 文件配置，以 `baseline_id` 命名。

```yaml
baseline_id: 1200                                          # 基线 ID（int）
baseline_version: "1.0"                                    # 基线版本
baseline_name: "centos基线检查"                              # 基线名称
baseline_name_en: "centos Security Baseline Check"          # 基线名称（英文）
system:                                                     # 适用系统列表
  - "centos"
check_list:                                                 # 检查项列表
  - check_id: 1
    type: "Identification"
    # ... 检查项字段（见下文）
    check:
      condition: "all"
      rules:
        # ... 规则列表（见下文）
```

**已有基线**：

| baseline_id | 适用系统 | 说明 |
|-------------|----------|------|
| 1200 | centos | CentOS 安全基线 |
| 1300 | debian | Debian 安全基线 |
| 1400 | ubuntu | Ubuntu 安全基线 |

---

## 检查项字段说明

每个检查项（`check_list` 中的元素）包含以下字段：

```yaml
- check_id: 1                      # 检查项 ID（int，基线内唯一）
  type: "Identification"            # 检查类别（英文）
  title: "Ensure password ..."      # 标题（英文）
  description: "The PASS_MAX..."    # 描述（英文）
  solution: "Set the PASS_MAX..."   # 解决方案（英文）
  security: "high"                  # 安全等级：high / mid / low
  type_cn: "身份鉴别"               # 检查类别
  title_cn: "设置密码失效时间"        # 标题
  description_cn: "请设置密码..."    # 描述
  solution_cn: "在 /etc/login..."   # 解决方案
  check:                            # 检查规则（核心部分）
    condition: "all"
    rules:
      - type: "file_line_check"
        param:
          - "/etc/login.defs"
        filter: '\s*\t*PASS_MAX_DAYS\s*\t*(\d+)'
        result: '$(<=)90'
```

**检查类别参考**：

| type | type_cn | 说明 |
|------|---------|------|
| Identification | 身份鉴别 | 用户密码策略相关检查 |
| SSH Configure | SSH检测 | SSH 服务配置检查 |
| security audit | 安全审计 | 日志服务检查 |
| Intrusion prevention | 入侵防范 | 系统安全措施检查 |
| File Permissions | 文件权限 | 文件权限与归属检查 |

---

## 规则配置详解

每个检查项的核心是 `check` 字段，包含 `condition`（条件逻辑）和 `rules`（规则列表）。

### 规则类型 (rules.type)

`rules.type` 指定检查方式，目前支持 7 种内置规则（实现见 `check/rules.go`）：

| 规则类型 | 含义 | 参数 | 返回值 |
|----------|------|------|--------|
| `command_check` | 执行命令行语句 | 1: 命令语句<br>2: 特殊参数（可选） | 命令输出（string）或 bool |
| `file_line_check` | 逐行遍历文件匹配 | 1: 文件路径<br>2: 行标记（可选）<br>3: 注释符（可选，默认 `#`） | true / false |
| `file_permission` | 检测文件权限 | 1: 文件路径<br>2: 最低权限（8进制，如 `644`） | true / false |
| `if_file_exist` | 判断文件是否存在 | 1: 文件路径 | true / false |
| `file_user_group` | 判断文件所属用户组 | 1: 文件路径<br>2: `用户ID:组ID` | true / false |
| `file_md5_check` | 判断文件 MD5 是否一致 | 1: 文件路径<br>2: 期望的 MD5 值 | true / false |
| `func_check` | 内置特殊检查函数 | 1: 函数标识 | true / false |

---

### 规则参数 (rules.param)

`param` 是字符串数组，不同规则类型的参数含义不同，详见下方各规则示例。

---

### 过滤器 (rules.filter)

`filter` 仅在 `file_line_check` 类型中生效，用于从匹配行中提取子串，供 `result` 做进一步判断。

**工作流程**：
1. 逐行读取文件，跳过注释行（以注释符开头的行）
2. 如果设置了 `param[1]`（行标记），只处理包含该标记的行
3. 使用 `filter` 正则表达式匹配当前行
4. 如果正则包含捕获组 `()`，提取第一个捕获组的值
5. 将提取的值传给 `result` 做比较判断

**示例**：从 `/etc/login.defs` 中提取 `PASS_MAX_DAYS` 的数值

```yaml
rules:
  - type: "file_line_check"
    param:
      - "/etc/login.defs"
    filter: '\s*\t*PASS_MAX_DAYS\s*\t*(\d+)'   # 捕获组提取数值部分
    result: '$(<=)90'                            # 对提取的数值做 <=90 判断
```

文件中如果有行 `PASS_MAX_DAYS  90`，filter 会捕获 `90`，然后与 `result` 中 `$(<=)90` 比较，90 <= 90 成立，检查通过。

---

### 前提条件 (rules.require)

`require` 用于设置规则的前提条件。某些安全配置只有在特定场景下才需要检查，如果前提条件不满足，则直接视为通过。

**目前支持的前提条件**：

| require 值 | 含义 | 说明 |
|-------------|------|------|
| `allow_ssh_passwd` | SSH 允许密码登录 | 检查 `/etc/ssh/sshd_config` 中 `PasswordAuthentication` 是否为 `yes`。如果 SSH 未开启密码登录，则跳过此规则（直接通过） |

**示例**：仅在 SSH 开启密码登录时才检查最大尝试次数

```yaml
rules:
  - type: "file_line_check"
    require: "allow_ssh_passwd"
    param:
      - "/etc/ssh/sshd_config"
    filter: '^\s*MaxAuthTries\s*\t*(\d+)'
    result: '$(<)5'
```

> 如果系统禁用了 SSH 密码登录（使用密钥认证），则该检查自动通过，不会产生误报。

---

### 结果匹配 (rules.result)

`result` 定义期望的检查结果，支持 bool、int、string 三种类型，其中 string 类型支持丰富的特殊语法。

#### 基础类型

```yaml
# bool 类型 —— 直接匹配返回值
result: true
result: false

# int 类型 —— 精确匹配数值
result: 0
result: 2

# string 类型 —— 正则匹配
result: 'enabled'                          # 匹配包含 enabled 的结果
result: 'active \(running\)'               # 正则匹配（注意转义括号）
result: 'PermitEmptyPasswords\s*\t*no'     # 正则匹配配置行
```

> **注意**：如果 `result` 未设置（为 nil），默认视为 `true`。

#### 特殊语法

string 类型的 `result` 支持以 `$()` 包裹的特殊运算符：

| 语法 | 说明 | 示例 | 含义 |
|------|------|------|------|
| `$(<=)` | 小于等于 | `$(<=)90` | 结果 <= 90 |
| `$(>=)` | 大于等于 | `$(>=)2` | 结果 >= 2 |
| `$(<)` | 小于 | `$(<)5` | 结果 < 5 |
| `$(>)` | 大于 | `$(>)0` | 结果 > 0 |
| `$(&&)` | 逻辑与（多条件分隔符） | `ok$(&&)success` | 结果匹配 ok **且** 匹配 success |
| `$(not)` | 逻辑取反 | `$(not)error` | 结果**不**匹配 error |

#### 组合使用

多个条件通过 `$(&&)` 连接，每个子条件按顺序依次判断，任一子条件不通过则整体不通过（AND 语义）：

```yaml
# 数值小于 8 且不等于 2
result: '$(<)8$(&&)$(not)2'

# 该行不以 root: 开头 且 匹配 UID 为 0 的格式
result: '$(not)^root:$(&&)^\w+:\w+:0:'
```

#### filter + result 协同工作

当同时设置 `filter` 和 `result` 时：

1. `filter` 的正则捕获组提取出子串
2. `result` 中的运算符对提取的子串做判断

```yaml
# 从文件中提取 minclass 的值，判断是否 >= 3
filter: '^\s*minclass\s+\t*=\s+\t*(\d+)'
result: '$(>=)3'

# 从文件中提取 retry 的值，判断是否 <= 3
filter: 'try_first_pass.*retry=(\d+)'
result: '$(<=)3'
```

如果没有 `filter`，`result` 直接对规则函数的返回值做匹配。

---

### 条件逻辑 (check.condition)

一个检查项可能包含多条规则，`condition` 定义这些规则之间的逻辑关系：

| condition | 含义 | 说明 |
|-----------|------|------|
| `all`（默认） | 全部通过 | 所有规则都通过，检查项才通过（AND） |
| `any` | 任一通过 | 任一规则通过，检查项即通过（OR） |
| `none` | 全部不通过 | 所有规则都不通过，检查项才通过（NOR） |

> 如果未设置 `condition`，默认为 `all`。

**示例**：

```yaml
# all —— 同时检查多个文件权限，全部合规才通过
check:
  condition: "all"
  rules:
    - type: "file_permission"
      param: ["/etc/passwd", "644"]
    - type: "file_permission"
      param: ["/etc/shadow", "400"]

# any —— 只要有一种方式检测到服务在运行
check:
  condition: "any"
  rules:
    - type: "command_check"
      param: ["systemctl is-enabled rsyslog"]
      result: 'enabled'
    - type: "command_check"
      param: ["service rsyslog status"]
      result: 'running'

# none —— 确保不存在 UID 为 0 的非 root 用户
check:
  condition: "none"
  rules:
    - type: "file_line_check"
      param: ["/etc/passwd"]
      result: '$(not)^root:$(&&)^\w+:\w+:0:'
```

---



## 各规则类型配置示例

### 1. command_check — 命令行检查

执行 shell 命令并对输出结果做匹配。

**参数**：
- `param[0]`：命令行语句（必填）
- `param[1]`：特殊参数（可选），目前支持 `ignore_exit`（命令执行报错时视为通过）

```yaml
# 示例 1：检查 auditd 服务是否启用
rules:
  - type: "command_check"
    param:
      - "systemctl is-enabled auditd"
    result: 'enabled'

# 示例 2：检查服务运行状态（正则需转义括号）
rules:
  - type: "command_check"
    param:
      - "systemctl status auditd"
    result: 'active \(running\)'

# 示例 3：检查内核参数 ASLR
rules:
  - type: "command_check"
    param:
      - "sysctl kernel.randomize_va_space"
    result: '^\s*kernel.randomize_va_space\s*=\s*2'

# 示例 4：使用 ignore_exit，命令不存在或报错时视为通过
rules:
  - type: "command_check"
    param:
      - "systemctl is-enabled some_optional_service"
      - "ignore_exit"
    result: 'enabled'
```

> **注意**：命令通过空格分割为参数数组传给 `exec.Command`，不经过 shell 解释。如需管道、重定向等 shell 特性，可使用 `grep -Rh` 等单命令替代。

---

### 2. file_line_check — 文件逐行匹配

逐行读取文件，对每一行做正则匹配。这是最常用的规则类型。

**参数**：
- `param[0]`：文件绝对路径（必填）
- `param[1]`：行标记 flag（可选），用于快速筛选行，只有包含该字符串的行才会进入正则匹配
- `param[2]`：注释符（可选，默认 `#`），以注释符开头的行会被跳过

**匹配逻辑**：
1. 跳过以注释符开头的行
2. 如果设定了行标记，跳过不包含标记的行
3. 如果设定了 `filter`，用 filter 正则提取子串后交给 `result` 判断
4. 如果未设 `filter`，用 `result` 直接正则匹配整行
5. 只要有一行匹配成功，即返回 `true`

```yaml
# 示例 1：简单正则匹配 — 检查 SSH 是否禁止空密码
rules:
  - type: "file_line_check"
    param:
      - "/etc/ssh/sshd_config"
    result: 'PermitEmptyPasswords\s*\t*no'

# 示例 2：filter 提取 + 数值比较 — 密码最长有效期
rules:
  - type: "file_line_check"
    param:
      - "/etc/login.defs"
    filter: '\s*\t*PASS_MAX_DAYS\s*\t*(\d+)'
    result: '$(<=)90'

# 示例 3：filter 提取 + 数值比较 — 密码复杂度
rules:
  - type: "file_line_check"
    param:
      - "/etc/security/pwquality.conf"
    filter: '^\s*minlen\s+\t*=\s+\t*(\d+)'
    result: '$(>=)8'

# 示例 4：复杂正则 — 检测 /etc/shadow 中空密码账户
check:
  condition: "none"
  rules:
    - type: "file_line_check"
      param:
        - "/etc/shadow"
      result: '^\w+::'

# 示例 5：带前提条件 — SSH 最大尝试次数
rules:
  - type: "file_line_check"
    require: "allow_ssh_passwd"
    param:
      - "/etc/ssh/sshd_config"
    filter: '^\s*MaxAuthTries\s*\t*(\d+)'
    result: '$(<)5'

# 示例 6：使用行标记 flag 加速匹配
rules:
  - type: "file_line_check"
    param:
      - "/etc/ssh/sshd_config"
      - "LogLevel"                    # 只匹配包含 LogLevel 的行
    result: 'LogLevel\s*\t*INFO'

# 示例 7：指定自定义注释符
rules:
  - type: "file_line_check"
    param:
      - "/etc/some/config.ini"
      - ""                           # flag 留空
      - ";"                          # 注释符为分号
    result: 'some_key\s*=\s*expected_value'
```

> **文件不存在时的行为**：如果文件不存在，`file_line_check` 不会报错，而是跳过该规则（返回 false）。

---

### 3. file_permission — 文件权限检查

检测文件实际权限是否比要求的权限更严格（数值更小）。

**参数**：
- `param[0]`：文件绝对路径
- `param[1]`：最低权限要求（8进制格式，如 `644`）

**判断逻辑**：`实际权限 < 要求权限` 时返回 true（即实际权限更严格则通过）。

```yaml
# 示例：检查关键系统文件权限
check:
  condition: "all"
  rules:
    - type: "file_permission"
      param:
        - "/etc/passwd"
        - "644"
    - type: "file_permission"
      param:
        - "/etc/shadow"
        - "400"
    - type: "file_permission"
      param:
        - "/etc/group"
        - "644"
    - type: "file_permission"
      param:
        - "/etc/gshadow"
        - "400"
```

> **文件不存在时的行为**：如果文件不存在，返回 true（通过）。

---

### 4. if_file_exist — 文件存在性检查

判断指定路径的文件是否存在。

**参数**：
- `param[0]`：文件绝对路径

```yaml
# 示例 1：确保某安全配置文件存在
rules:
  - type: "if_file_exist"
    param:
      - "/etc/security/pwquality.conf"
    result: true

# 示例 2：确保某危险文件不存在
rules:
  - type: "if_file_exist"
    param:
      - "/etc/hosts.equiv"
    result: false
```

---

### 5. file_user_group — 文件归属检查

检查文件的所属用户 ID 和组 ID 是否符合要求。

**参数**：
- `param[0]`：文件绝对路径
- `param[1]`：`用户ID:组ID`（如 `0:0` 表示 root:root）

```yaml
# 示例：确保关键文件属于 root 用户
rules:
  - type: "file_user_group"
    param:
      - "/etc/passwd"
      - "0:0"
    result: true

# 示例：检查多个文件归属
check:
  condition: "all"
  rules:
    - type: "file_user_group"
      param:
        - "/etc/passwd"
        - "0:0"
    - type: "file_user_group"
      param:
        - "/etc/shadow"
        - "0:0"
    - type: "file_user_group"
      param:
        - "/etc/group"
        - "0:0"
```

---

### 6. file_md5_check — 文件 MD5 校验

通过 MD5 哈希值验证文件内容完整性。

**参数**：
- `param[0]`：文件绝对路径
- `param[1]`：期望的 MD5 哈希值

```yaml
# 示例：校验关键二进制文件未被篡改
rules:
  - type: "file_md5_check"
    param:
      - "/usr/sbin/sshd"
      - "a1b2c3d4e5f6..."
    result: true
```

> 可通过 `md5sum /path/to/file` 获取文件的 MD5 值。
