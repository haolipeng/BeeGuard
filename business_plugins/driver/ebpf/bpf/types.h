// SPDX-License-Identifier: GPL-2.0
#ifndef __TYPES_H
#define __TYPES_H

// 事件类型标识
#define EVENT_TYPE_EXECVE        1
#define EVENT_TYPE_COMMIT_CREDS  2
#define EVENT_TYPE_REVERSE_SHELL 3
#define EVENT_TYPE_CONNECT       4
#define EVENT_TYPE_BIND          5
#define EVENT_TYPE_ACCEPT        6
#define EVENT_TYPE_DNS           7

// 路径相关常量
#define PATH_MAX_ENTS   16    // dentry 链最大遍历深度
#define PATH_BUF_SIZE   512   // 路径重建工作缓冲区大小（必须为2的幂）
#define PATH_BUF_MASK   (PATH_BUF_SIZE - 1)  // 位掩码，用于安全索引
#define PATH_NAME_LEN   256   // 单个 dentry 名称最大长度（必须为2的幂）
#define PATH_NAME_MASK  (PATH_NAME_LEN - 1)  // 位掩码，满足 BPF 验证器要求

// Per-CPU 路���构建缓冲区（避免栈溢出）
struct path_buf {
    char data[PATH_BUF_SIZE];       // 路径重建缓冲区
    char swap[PATH_NAME_LEN + 4];   // dentry 名称临时缓冲区
};

// execve事件结构体（批次3增强：添加父进程信息）
struct execve_event {
    __u8  event_type;    // 事件类型标识 (EVENT_TYPE_EXECVE = 1)
    __u8  padding1[3];   // 对齐填充
    __u32 pid;           // 进程ID（��程ID）
    __u32 tgid;          // 线程组ID（进程ID）
    __u32 ppid;          // 父进程ID（批次3新增）
    __u32 pgid;          // 进程组ID（批次3新增）
    __u32 uid;           // 用户ID
    __u32 padding;       // 对齐填充
    char comm[16];       // 进程名（最多16字节）
    char exe_path[256];  // 可执行文件完整路径（批次3增强）
    char args[512];      // 命令行参数
} __attribute__((packed));

// commit_creds提权事���结构体
struct commit_creds_event {
    __u8  event_type;    // 事件类型标识 (EVENT_TYPE_COMMIT_CREDS = 2)
    __u8  padding1[3];   // 对齐填充
    __u32 pid;           // 进程ID
    __u32 tgid;          // 线程组ID
    __u32 ppid;          // 父进程ID
    __u32 uid;           // 当前用户ID
    __u32 old_uid;       // 提权前的uid
    __u32 old_euid;      // 提权前的euid
    __u32 new_uid;       // 提权后的uid
    __u32 new_euid;      // 提权后的euid
    char comm[16];       // 进程名
    char exe_path[256];  // 可执行文件路径
} __attribute__((packed));

// 反弹Shell检测事件结构体
struct reverse_shell_event {
    __u8  event_type;     // EVENT_TYPE_REVERSE_SHELL = 3
    __u8  fd_type;        // 触发的FD: 1=stdin, 2=stdout, 3=both
    __u8  padding1[2];    // 对齐填充
    __u32 pid;            // 进程ID（线程ID）
    __u32 tgid;           // 线程组ID（进程ID）
    __u32 ppid;           // 父进程ID
    __u32 pgid;           // 进程组ID
    __u32 uid;            // 用户ID
    __u32 remote_ip;      // 远程IPv4地址（网络字节序）
    __u16 remote_port;    // 远程端口（网络字节序）
    __u16 local_port;     // 本地端口（主机字节序）
    __u32 local_ip;       // 本地IPv4地址（网络字节序）
    char  comm[16];       // 进程名
    char  exe_path[256];  // 可执行文件路径
    char  args[512];      // 命令行参数
} __attribute__((packed));

// connect 出站连接事件结构体
struct connect_event {
    __u8  event_type;     // EVENT_TYPE_CONNECT = 4
    __u8  protocol;       // IPPROTO_TCP=6, IPPROTO_UDP=17
    __u8  padding1[2];
    __u32 pid;
    __u32 tgid;
    __u32 ppid;
    __u32 uid;
    __u32 remote_ip;      // 目�� IP（网络字节序）
    __u16 remote_port;    // 目标端口（网络字节序）
    __u16 local_port;     // 本地端口
    __u32 local_ip;       // 本地 IP（网络字节序）
    __s32 retval;         // 系统调用返回值（0=成功，负数=失败）
    char  comm[16];
    char  exe_path[256];
} __attribute__((packed));

// bind 端口绑定事件结构体
struct bind_event {
    __u8  event_type;     // EVENT_TYPE_BIND = 5
    __u8  protocol;
    __u8  padding1[2];
    __u32 pid;
    __u32 tgid;
    __u32 ppid;
    __u32 uid;
    __u32 bind_ip;        // 绑定 IP
    __u16 bind_port;      // 绑定端口
    __u16 padding2;
    __s32 retval;
    char  comm[16];
    char  exe_path[256];
} __attribute__((packed));

// accept 入站连接事件结构体
struct accept_event {
    __u8  event_type;     // EVENT_TYPE_ACCEPT = 6
    __u8  protocol;
    __u8  padding1[2];
    __u32 pid;
    __u32 tgid;
    __u32 ppid;
    __u32 uid;
    __u32 remote_ip;      // 连接来源 IP
    __u16 remote_port;    // 连接来源端口
    __u16 local_port;     // 本地监听端口
    __u32 local_ip;
    __s32 retval;
    char  comm[16];
    char  exe_path[256];
} __attribute__((packed));

// DNS 相关常量
#define DNS_DOMAIN_MAX  256
#define DNS_RECORD_MAX  512
#define DNS_RECORD_MASK (DNS_RECORD_MAX - 1)

// DNS 原始包临时缓冲区（per-CPU map 使用）
struct dns_data_buf {
    char data[DNS_RECORD_MAX];
};

struct dns_event {
    __u8  event_type;     // EVENT_TYPE_DNS = 7
    __u8  padding1[3];
    __u32 pid;
    __u32 tgid;
    __u32 ppid;
    __u32 uid;
    __u32 dns_server_ip;  // DNS 服务器 IP
    __u16 dns_server_port;// DNS 服务器端口（通常 53）
    __u16 query_type;     // DNS 查询类型 (A=1, AAAA=28, TXT=16, MX=15, CNAME=5)
    __s32 opcode;         // DNS 操作码
    __s32 rcode;          // DNS 响应码
    char  comm[16];
    char  exe_path[256];
    char  domain[DNS_DOMAIN_MAX]; // 查询域名
} __attribute__((packed));

// 可信任可执行文件条目 (必须与 Go ExeItem 结构体匹配)
#define CMDLINE_LEN 2048

struct exe_item {
	int   len;      // 字符串长度 (不含 \0)
	__u32 sid;      // 预留字段
	__u64 hash;     // Murmur OAAT64 哈希
	char  name[CMDLINE_LEN];  // 可执行文件路径
} __attribute__((packed));

#endif // __TYPES_H
