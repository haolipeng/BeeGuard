// SPDX-License-Identifier: GPL-2.0
#ifndef __TYPES_H
#define __TYPES_H

// 事件类型标识
#define EVENT_TYPE_EXECVE        1
#define EVENT_TYPE_COMMIT_CREDS  2
#define EVENT_TYPE_CONNECT       4
#define EVENT_TYPE_BIND          5
#define EVENT_TYPE_ACCEPT        6
#define EVENT_TYPE_DNS           7
#define EVENT_TYPE_FILE          8   // 文件操作事件

// 文件操作 action 常量
#define FILE_ACTION_CREATE  1   // 文件创建
#define FILE_ACTION_RENAME  2   // 文件重命名
#define FILE_ACTION_DELETE  3   // 文件删除

// 文件系统 ID 常量
#define FS_ID_LEN  32   // 文件系统 ID 最大长度（与内核 s_id 一致）

// 路径相关常量
#define PATH_MAX_ENTS   16    // dentry 链最大遍历深度
#define PATH_BUF_SIZE   512   // 路径重建工作缓冲区大小（必须为2的幂）
#define PATH_BUF_MASK   (PATH_BUF_SIZE - 1)  // 位掩码，用于安全索引
#define PATH_NAME_LEN   256   // 单个 dentry 名称最大长度（必须为2的幂）
#define PATH_NAME_MASK  (PATH_NAME_LEN - 1)  // 位掩码，满足 BPF 验证器要求

// 反弹 Shell 增强采集常量
#define STDIO_PATH_LEN   64    // stdin/stdout 路径最大长度
#define TTY_NAME_LEN     64    // tty 名称最大长度
#define SOCK_FD_LIMIT    16    // FD 扫描范围 0-15

// Per-CPU 路径构建缓冲区（避免栈溢出）
struct path_buf {
    char data[PATH_BUF_SIZE];       // 路径重建缓冲区
    char swap[PATH_NAME_LEN + 4];   // dentry 名称临时缓冲区
};

// Per-CPU stdin/stdout 路径构建缓冲区（与 percpu_path_buf 独立，避免冲突）
struct stdio_path_buf {
    char data[PATH_BUF_SIZE];       // 路径重建缓冲区 (512 字节)
    char swap[PATH_NAME_LEN + 4];   // dentry 名称临时缓冲区 (260 字节)
};

// execve事件结构体（增强：反弹 shell 数据采集）
struct execve_event {
    // --- 基础字段 ---
    __u8  event_type;    // 事件类型标识 (EVENT_TYPE_EXECVE = 1)
    __u8  fd_type;       // 内核预过滤标志: 0=无socket, 1=stdin是socket, 2=stdout, 3=both
    __u8  padding1[2];   // 对齐填充
    __u32 pid;           // 进程ID（线程ID）
    __u32 tgid;          // 线程组ID（进程ID）
    __u32 ppid;          // 父进程ID
    __u32 pgid;          // 进程组ID
    __u32 uid;           // 用户ID
    __u32 socket_pid;    // 持有 socket 的进程 PID
    char  comm[16];      // 进程名（最多16字节）
    char  exe_path[256]; // 可执行文件完整路径
    char  args[512];     // 命令行参数
    // --- 反弹 shell 增强字段 ---
    char  stdin_path[STDIO_PATH_LEN];   // FD 0 的文件路径（如 /dev/pts/0 或 socket:[xxx]）
    char  stdout_path[STDIO_PATH_LEN];  // FD 1 的文件路径
    char  tty_name[TTY_NAME_LEN];       // 控制终端名称
    __u32 remote_ip;     // socket 远程 IP（网络字节序）
    __u16 remote_port;   // socket 远程端口（网络字节序）
    __u16 local_port;    // socket 本地端口（主机字节序）
    __u32 local_ip;      // socket 本地 IP（网络字节序）
} __attribute__((packed));

// commit_creds提权事件结构体
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

// connect 出站连接事件结构体
struct connect_event {
    __u8  event_type;     // EVENT_TYPE_CONNECT = 4
    __u8  protocol;       // IPPROTO_TCP=6, IPPROTO_UDP=17
    __u8  padding1[2];
    __u32 pid;
    __u32 tgid;
    __u32 ppid;
    __u32 uid;
    __u32 remote_ip;      // 目标 IP（网络字节序）
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

// 文件操作事件结构体（文件创建/重命名/删除监控）
struct file_event {
    __u8  event_type;       // EVENT_TYPE_FILE = 8
    __u8  action;           // FILE_ACTION_CREATE=1, FILE_ACTION_RENAME=2, FILE_ACTION_DELETE=3
    __u8  padding1[2];
    __u32 pid;
    __u32 tgid;
    __u32 ppid;
    __u32 uid;
    __u32 socket_pid;       // 持有 socket 的进程 PID（仅创建事件）
    __u32 remote_ip;        // socket 远程 IP（网络字节序）
    __u16 remote_port;      // socket 远程端口（网络字节序）
    __u16 local_port;       // socket 本地端口
    __u32 local_ip;         // socket 本地 IP（网络字节序）
    char  comm[16];
    char  exe_path[256];    // 操作进程的可执行文件路径
    char  new_path[PATH_BUF_SIZE]; // 创建：文件路径；重命名：新路径
    char  old_path[PATH_BUF_SIZE]; // 仅重命名：旧路径（创建时全零）
    char  s_id[FS_ID_LEN];  // 文件系统 ID（ext4/xfs/tmpfs 等）
} __attribute__((packed));

#endif // __TYPES_H
