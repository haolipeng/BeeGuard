// SPDX-License-Identifier: GPL-2.0
#ifndef __TYPES_H
#define __TYPES_H

// 事件类型标识
#define EVENT_TYPE_EXECVE       1
#define EVENT_TYPE_COMMIT_CREDS 2

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

#endif // __TYPES_H
