// SPDX-License-Identifier: GPL-2.0
#ifndef __TYPES_H
#define __TYPES_H

// 事件类型标识
#define EVENT_TYPE_EXECVE       1
#define EVENT_TYPE_COMMIT_CREDS 2

// 路径相关常量
#define PATH_MAX_ENTS   16    // dentry 链最大遍历深度
#define PATH_BUF_SIZE   512   // 路径重建工作缓冲区大小
#define PATH_NAME_LEN   256   // 单个 dentry 名称最大长度

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

// 可信任可执行文件条目 (必须与 Go ExeItem 结构体匹配)
#define CMDLINE_LEN 2048

struct exe_item {
	int   len;      // 字符串长度 (不含 \0)
	__u32 sid;      // 预留字段
	__u64 hash;     // Murmur OAAT64 哈希
	char  name[CMDLINE_LEN];  // 可执行文件路径
} __attribute__((packed));

#endif // __TYPES_H
