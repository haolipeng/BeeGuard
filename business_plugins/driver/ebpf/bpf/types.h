// SPDX-License-Identifier: GPL-2.0
#ifndef __TYPES_H
#define __TYPES_H

// execve事件结构体
struct execve_event {
    __u32 pid;           // 进程ID
    __u32 tgid;          // 线程组ID
    __u32 uid;           // 用户ID
    char comm[16];       // 进程名（最多16字节）
    char exe_path[256];  // 可执行文件路径（预留，批次2实现）
    char args[512];      // 命令行参数（预留，批次2实现）
} __attribute__((packed));

#endif // __TYPES_H
