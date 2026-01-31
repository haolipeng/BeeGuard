// SPDX-License-Identifier: GPL-2.0
#include <vmlinux.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_core_read.h>
#include "types.h"

// Perf Event Array Map - 用于将事件从内核态传递到用户态
struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
    __uint(key_size, sizeof(u32));
    __uint(value_size, sizeof(u32));
} events SEC(".maps");

// Per-CPU Array Map - 用于存储大结构体，避免栈溢出（eBPF栈限制512字节）
struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __type(key, u32);
    __type(value, struct execve_event);
    __uint(max_entries, 1);
} percpu_buf SEC(".maps");

// Helper: 读取可执行文件路径
// 批次2简化版：只读取文件名，完整路径在后续批次实现
static __always_inline int read_exe_path(struct task_struct *task, char *buf, int buf_size)
{
    struct file *exe_file;
    struct dentry *dentry;
    struct qstr d_name;
    int len;

    // 读取 task->mm->exe_file
    exe_file = BPF_CORE_READ(task, mm, exe_file);
    if (!exe_file)
        return 0;

    // 读取 dentry
    dentry = BPF_CORE_READ(exe_file, f_path.dentry);
    if (!dentry)
        return 0;

    // 读取文件名（dentry的名称）
    d_name = BPF_CORE_READ(dentry, d_name);
    len = d_name.len;

    if (len <= 0 || len >= buf_size)
        return 0;

    // 使用 bpf_probe_read_kernel_str 读取文件名
    len = bpf_probe_read_kernel_str(buf, buf_size, d_name.name);
    if (len < 0)
        return 0;

    return len;
}

// Helper: 读取命令行参数
// 从 task->mm->arg_start 到 arg_end 读取用户态内存
static __always_inline int read_args(struct task_struct *task, char *buf, int buf_size)
{
    unsigned long arg_start, arg_end;
    unsigned long args_len;
    int ret;

    // 读取参数的起始和结束地址
    arg_start = BPF_CORE_READ(task, mm, arg_start);
    arg_end = BPF_CORE_READ(task, mm, arg_end);

    if (!arg_start || !arg_end || arg_start >= arg_end)
        return 0;

    // 计算参数长度，并确保为正数（满足eBPF验证器要求）
    args_len = arg_end - arg_start;

    // 限制长度范围，确保验证器知道这是一个合法的正数
    if (args_len <= 0 || args_len > 4096)
        return 0;

    if (args_len > buf_size)
        args_len = buf_size;

    // 再次边界检查，确保args_len在合法范围内
    if (args_len <= 0 || args_len > buf_size)
        return 0;

    // 从用户态内存读取参数（使用按位与确保验证器知道范围）
    // 注意：参数之间使用NULL分隔，在Go层面会进行后处理
    ret = bpf_probe_read_user(buf, args_len & 511, (void *)arg_start);
    if (ret < 0)
        return 0;

    return (int)args_len;
}

// 监听进程执行事件
// Hook点: sched_process_exec - 在execve系统调用成功后触发
SEC("raw_tracepoint/sched_process_exec")
int tp_proc_exec(struct bpf_raw_tracepoint_args *ctx)
{
    u32 key = 0;
    struct task_struct *task;

    // 从Per-CPU Map获取缓冲区（避免栈溢出）
    struct execve_event *evt = bpf_map_lookup_elem(&percpu_buf, &key);
    if (!evt)
        return 0;

    // 清零结构体
    __builtin_memset(evt, 0, sizeof(*evt));

    // 获取当前进程的PID和TGID
    u64 id = bpf_get_current_pid_tgid();
    evt->pid = id;           // 低32位：线程ID
    evt->tgid = id >> 32;    // 高32位：进程ID（线程组ID）

    // 获取当前进程的UID
    u64 uid_gid = bpf_get_current_uid_gid();
    evt->uid = uid_gid;      // 低32位：UID

    // 获取进程名（comm字段，最多16字节）
    bpf_get_current_comm(&evt->comm, sizeof(evt->comm));

    // 获取当前task_struct
    task = (struct task_struct *)bpf_get_current_task();

    // 读取可执行文件路径（批次2新增）
    read_exe_path(task, evt->exe_path, sizeof(evt->exe_path));

    // 读取命令行参数（批次2新增）
    read_args(task, evt->args, sizeof(evt->args));

    // 通过Perf Event Array输出事件到用户态
    // BPF_F_CURRENT_CPU: 使用当前CPU的缓冲区，避免锁竞争
    bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU,
                          evt, sizeof(*evt));

    return 0;
}

// eBPF程序许可证声明（必需）
char LICENSE[] SEC("license") = "GPL";
