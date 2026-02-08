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

// 可信任可执行文件白名单 Map - 用于过滤提权事件
struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __uint(max_entries, 2048);
    __type(key, __u64);      // Murmur 哈希值
    __type(value, struct exe_item);
} trusted_exes SEC(".maps");

// Per-CPU buffer for path construction - 避免栈溢出
struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __type(key, u32);
    __type(value, struct path_buf);
    __uint(max_entries, 1);
} percpu_path_buf SEC(".maps");

// 从缓冲区末尾反向写入路径分量（'/' + name）
static __always_inline int prepend_path(char *buf, int *pos, const char *name, int name_len)
{
    // 边界检查
    if (name_len <= 0 || name_len >= PATH_NAME_LEN)
        return -1;
    if (*pos < name_len + 1)  // '/' + name
        return -1;

    // 先写名称
    *pos -= name_len;
    bpf_probe_read_kernel(&buf[*pos & (PATH_BUF_SIZE - 1)],
                          name_len & (PATH_NAME_LEN - 1), name);
    // 再写 '/'
    (*pos)--;
    buf[*pos & (PATH_BUF_SIZE - 1)] = '/';

    return 0;
}

// 读取 dentry 名称并反向写入缓冲区
static __always_inline int prepend_entry(char *buf, int *pos, char *swap, struct dentry *de)
{
    struct qstr d_name = BPF_CORE_READ(de, d_name);
    int name_len = d_name.len;
    if (name_len <= 0 || name_len >= PATH_NAME_LEN)
        return -1;

    // 读取 dentry 名称到 swap 缓冲区
    int rc = bpf_probe_read_kernel_str(swap, PATH_NAME_LEN, d_name.name);
    if (rc <= 1)
        return -1;

    // rc 包含 \0，实际名称长度 = rc - 1
    return prepend_path(buf, pos, swap, rc - 1);
}

// 通过遍历 dentry 链读取可执行文件完整路径
// 返回不含 \0 的路径长度，0 表示失败
static __always_inline int read_full_exe_path(struct task_struct *task, char *out_buf, int out_size)
{
    u32 key = 0;
    struct path_buf *pbuf;
    struct file *exe_file;
    struct dentry *dentry, *parent;
    int pos, path_len, i;

    __builtin_memset(out_buf, 0, out_size);

    // 获取 per-CPU 工作缓冲区
    pbuf = bpf_map_lookup_elem(&percpu_path_buf, &key);
    if (!pbuf)
        return 0;

    // 获取 exe_file->f_path.dentry
    exe_file = BPF_CORE_READ(task, mm, exe_file);
    if (!exe_file)
        return 0;
    dentry = BPF_CORE_READ(exe_file, f_path.dentry);
    if (!dentry)
        return 0;

    // 从缓冲区末尾开始反向构建��径
    pos = PATH_BUF_SIZE - 1;
    pbuf->data[pos] = '\0';

    // 遍历 dentry 链（最多 PATH_MAX_ENTS=16 层）
    for (i = 0; i < PATH_MAX_ENTS; i++) {
        parent = BPF_CORE_READ(dentry, d_parent);
        if (!parent || parent == dentry)
            break;  // 到达根节点

        if (prepend_entry(pbuf->data, &pos, pbuf->swap, dentry))
            break;  // 出错

        dentry = parent;
    }

    // 检查是否成功写入了路径内容
    if (pos >= PATH_BUF_SIZE - 1) {
        // 没有写入任何内容，回退到读取文件名
        struct dentry *orig_dentry;
        orig_dentry = BPF_CORE_READ(exe_file, f_path.dentry);
        if (!orig_dentry)
            return 0;
        struct qstr d_name = BPF_CORE_READ(orig_dentry, d_name);
        int len = d_name.len;
        if (len <= 0 || len >= out_size)
            return 0;
        len = bpf_probe_read_kernel_str(out_buf, out_size, d_name.name);
        if (len <= 1)
            return 0;
        return len - 1;
    }

    // 计算路径长度并复制到输出缓冲区
    path_len = PATH_BUF_SIZE - 1 - pos;
    if (path_len <= 0 || path_len >= out_size)
        return 0;

    bpf_probe_read_kernel(out_buf, path_len & (out_size - 1),
                          &pbuf->data[pos & (PATH_BUF_SIZE - 1)]);

    return path_len;
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

// Murmur OAAT64 哈希函数 (必须与 Go 版本字节级兼容)
// BPF 验证器限制: 使用有界循环,最多处理 256 字节
static __always_inline __u64 hash_murmur_OAAT64(const char *s, int len)
{
    __u64 h = 525201411107845655ull;
    int i;

    // 限制长度以满足 BPF 验证器要求
    if (len > 256)
        len = 256;

    // 使用有界循环 (不使用 #pragma unroll)
    for (i = 0; i < 256 && i < len; i++) {
        h ^= (__u64)(s[i]);
        h *= 0x5bd1e9955bd1e995;
        h ^= h >> 47;
    }

    return h;
}

// 检查可执行文件是否可信任
static __always_inline int exe_is_trusted(const char *exe_path, int path_len)
{
    __u64 hash = hash_murmur_OAAT64(exe_path, path_len);
    struct exe_item *ei;

    ei = bpf_map_lookup_elem(&trusted_exes, &hash);

    // 同时检查哈希和长度匹配 (防止哈希碰撞)
    return (ei && ei->len == path_len);
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
    struct cred *new_cred = (void *)PT_REGS_PARM1_CORE(ctx);
    if (!new_cred)
        return 0;

    // 读取旧凭证（当前task的real_cred）
    int old_uid = BPF_CORE_READ(task, real_cred, uid.val);
    int old_euid = BPF_CORE_READ(task, real_cred, euid.val);

    // 读取新凭证
    int new_uid = BPF_CORE_READ(new_cred, uid.val);
    int new_euid = BPF_CORE_READ(new_cred, euid.val);

    // 调试日志1: hook点被调用，打印所有凭证信息
    bpf_printk("hids: commit_creds CALLED old_uid=%u old_euid=%u new_uid=%u new_euid=%u\n",
               old_uid, old_euid);
    bpf_printk("hids: commit_creds new creds: new_uid=%u new_euid=%u\n",
               new_uid, new_euid);

    // 检测条件: 原uid和euid都非0，新uid或euid为0（提权到root）
    if ((old_uid != 0 || old_euid != 0) && (new_uid == 0 || new_euid == 0)) {
        // 调试日志2: 提权条件满足
        bpf_printk("hids: PRIVILEGE ESCALATION DETECTED! Condition matched\n");
        // **新增**: 读取可执行文件路径并检查是否可信任
        char exe_path_buf[256];
        __builtin_memset(exe_path_buf, 0, sizeof(exe_path_buf));
        int path_len = read_full_exe_path(task, exe_path_buf, sizeof(exe_path_buf));

        // 如果是可信任的可执行文件,跳过事件上报
        if (path_len > 0 && exe_is_trusted(exe_path_buf, path_len)) {
            bpf_printk("hids: exe_path=%s is in whitelist, skipping\n", exe_path_buf);
            return 0;  // 内核层直接过滤
        }

        // 调试日志3: 白名单检查通过，准备上报事件
        bpf_printk("hids: exe_path=%s NOT in whitelist, reporting event\n", exe_path_buf);

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

        // 调试打印：evt 所有字段（bpf_printk 每次最多约3个格式化参数，故分多行）
        bpf_printk("hids: commit_creds pid=%u tgid=%u ppid=%u\n", evt->pid, evt->tgid, evt->ppid);
        bpf_printk("hids: commit_creds uid=%u old_uid=%u old_euid=%u\n", evt->uid, evt->old_uid, evt->old_euid);
        bpf_printk("hids: commit_creds new_uid=%u new_euid=%u\n", evt->new_uid, evt->new_euid);
        bpf_printk("hids: commit_creds comm=%s\n", evt->comm);
        bpf_printk("hids: commit_creds exe_path=%s\n", evt->exe_path);

        // 通过Perf Event Array输出事件到用户态
        bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU,
                              evt, sizeof(*evt));
    }

    return 0;
}

// eBPF程序许可证声明（必需）
char LICENSE[] SEC("license") = "GPL";
