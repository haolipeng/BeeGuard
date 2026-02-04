// SPDX-License-Identifier: GPL-2.0
#ifndef __TYPES_H
#define __TYPES_H

// execve事件结构体（批次3增强：添加父进程信息）
struct execve_event {
    __u32 pid;           // 进程ID（线程ID）
    __u32 tgid;          // 线程组ID（进程ID）
    __u32 ppid;          // 父进程ID（批次3新增）
    __u32 pgid;          // 进程组ID（批次3新增）
    __u32 uid;           // 用户ID
    __u32 padding;       // 对齐填充
    char comm[16];       // 进程名（最多16字节）
    char exe_path[256];  // 可执行文件完整路径（批次3增强）
    char args[512];      // 命令行参数
} __attribute__((packed));

#endif // __TYPES_H
