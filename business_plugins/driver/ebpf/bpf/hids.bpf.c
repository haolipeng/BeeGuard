// SPDX-License-Identifier: GPL-2.0
#include <vmlinux.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_tracing.h>
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

// Per-CPU Array Map for commit_creds event - 用于提权检测
struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __type(key, u32);
    __type(value, struct commit_creds_event);
    __uint(max_entries, 1);
} percpu_creds_buf SEC(".maps");

// 完整路径采集在eBPF中实现复杂，受限于验证器
// 当前使用简化实现（读取文件名），后续可在Go层通过/proc补全
static __always_inline int read_full_exe_path(struct task_struct *task, char *buf, int buf_size)
{
    struct file *exe_file;
    struct dentry *dentry;
    struct qstr d_name;
    int len;

    // 初始化缓冲区
    __builtin_memset(buf, 0, buf_size);

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
    struct task_struct *parent;

    // 从Per-CPU Map获取缓冲区（避免栈溢出）
    struct execve_event *evt = bpf_map_lookup_elem(&percpu_buf, &key);
    if (!evt)
        return 0;

    __builtin_memset(evt, 0, sizeof(*evt));
    evt->event_type = EVENT_TYPE_EXECVE;

    // 获取当前task_struct
    task = (struct task_struct *)bpf_get_current_task();

    // 获取当前进程的PID和TGID
    u64 id = bpf_get_current_pid_tgid();
    evt->pid = id;           // 低32位：线程ID
    evt->tgid = id >> 32;    // 高32位：进程ID（线程组ID）

    // 获取父进程信息
    parent = BPF_CORE_READ(task, real_parent);
    if (parent) {
        evt->ppid = BPF_CORE_READ(parent, tgid);
    }

    // 获取进程组ID（简化版）
    // 完整的PGID需要复杂的namespace处理，这里使用简化实现
    evt->pgid = BPF_CORE_READ(task, tgid);

    // 获取当前进程的UID
    u64 uid_gid = bpf_get_current_uid_gid();
    evt->uid = uid_gid;      // 低32位：UID

    // 获取进程名（comm字段，最多16字节）
    bpf_get_current_comm(&evt->comm, sizeof(evt->comm));

    // 读取完整可执行文件路径
    read_full_exe_path(task, evt->exe_path, sizeof(evt->exe_path));

    // 读取命令行参数
    read_args(task, evt->args, sizeof(evt->args));

    // 通过Perf Event Array输出事件到用户态
    bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU,
                          evt, sizeof(*evt));

    return 0;
}

// 监听commit_creds调用，检测本地提权行为
// Hook点: kprobe/commit_creds - 在凭证提交前触发
// 检测条件: 原uid和euid都非0，新uid或euid为0（非root → root）
SEC("kprobe/commit_creds")
int kp_commit_creds(struct pt_regs *ctx)
{
    u32 key = 0;
    struct task_struct *task;
    struct task_struct *parent;

    // 从Per-CPU Map获取缓冲区
    struct commit_creds_event *evt = bpf_map_lookup_elem(&percpu_creds_buf, &key);
    if (!evt)
        return 0;

    __builtin_memset(evt, 0, sizeof(*evt));
    evt->event_type = EVENT_TYPE_COMMIT_CREDS;

    // 获取当前task_struct
    task = (struct task_struct *)bpf_get_current_task();

    // 获取新凭证（commit_creds的第一个参数）
    struct cred *new_cred = (struct cred *)PT_REGS_PARM1_CORE(ctx);
    if (!new_cred)
        return 0;

    // 读取旧凭证（当前task的real_cred）
    int old_uid = BPF_CORE_READ(task, real_cred, uid.val);
    int old_euid = BPF_CORE_READ(task, real_cred, euid.val);

    // 读取新凭证
    int new_uid = BPF_CORE_READ(new_cred, uid.val);
    int new_euid = BPF_CORE_READ(new_cred, euid.val);

    // 检测条件: 原uid和euid都非0，新uid或euid为0（提权到root）
    if (old_uid != 0 && old_euid != 0 && (new_uid == 0 || new_euid == 0)) {
        // 填充事件数据
        u64 id = bpf_get_current_pid_tgid();
        evt->pid = id;           // 低32位：线程ID
        evt->tgid = id >> 32;    // 高32位：进程ID

        // 获取父进程信息
        parent = BPF_CORE_READ(task, real_parent);
        if (parent) {
            evt->ppid = BPF_CORE_READ(parent, tgid);
        }

        // 获取当前UID
        evt->uid = bpf_get_current_uid_gid();

        // 记录uid变化
        evt->old_uid = old_uid;
        evt->old_euid = old_euid;
        evt->new_uid = new_uid;
        evt->new_euid = new_euid;

        // 获取进程名
        bpf_get_current_comm(&evt->comm, sizeof(evt->comm));

        // 读取可执行文件路径
        read_full_exe_path(task, evt->exe_path, sizeof(evt->exe_path));

        // 通过Perf Event Array输出事件到用户态
        bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU,
                              evt, sizeof(*evt));
    }

    return 0;
}

// eBPF程序许可证声明（必需）
char LICENSE[] SEC("license") = "GPL";
