// SPDX-License-Identifier: GPL-2.0
#include <vmlinux.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_tracing.h>
#include "types.h"

// container_of 宏 (BPF 程序中需自行定义)
#ifndef container_of
#define container_of(ptr, type, member) \
    ((type *)((void *)(ptr) - __builtin_offsetof(type, member)))
#endif

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

// Per-CPU Array Map for reverse_shell_event - 用于反弹Shell检测
struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __type(key, u32);
    __type(value, struct reverse_shell_event);
    __uint(max_entries, 1);
} percpu_rs_buf SEC(".maps");

// ========== 路径构建函数 ==========
// 向路径缓冲区追加路径分量 (缓冲区末尾反向写入，正向计数长度)
// data: 路径缓冲区, len: 已累积长度(含末尾\0), entry: 路径分量, num: 分量字节数
static __noinline int prepend_path(char *data, __u32 *len, char *entry, int num)
{
    if (!num)
        return 0;
    if (*len + num > PATH_BUF_SIZE)
        return -1;

    bpf_probe_read_kernel(&data[(PATH_BUF_SIZE - *len - num) & PATH_BUF_MASK],
                          num & PATH_NAME_MASK, entry);
    *len += num;
    return 0;
}

// 读取 dentry 名称并追加到路径缓冲区
// swap 布局: swap[3]='/' 前缀, swap[4..] 存 dentry 名称
static __noinline int prepend_entry(char *data, __u32 *len, char *swap, struct dentry *de)
{
    char *name;
    int rc;

    if (!de)
        return -1;
    name = (char *)BPF_CORE_READ(de, d_name.name);
    if (!name)
        return -1;
    rc = bpf_probe_read_kernel_str(&swap[4], PATH_NAME_LEN, name);
    if (rc <= 0)
        return -1;
    // 非 '/' 开头: 从 swap[3] 写入 "/name" (rc 字节)
    if (swap[4] != '/')
        rc = prepend_path(data, len, &swap[3], rc);
    else if (rc > 2)  // '/' 开头且长度>2: 跳过顶层 '/'
        rc = prepend_path(data, len, &swap[4], rc - 1);
    return rc;
}

// 从 vfsmount 反推 mount 结构体
static __always_inline struct mount *real_mount(struct vfsmount *mnt)
{
    return container_of(mnt, struct mount, mnt);
}

// 构建完整路径 (支持跨挂载点遍历)
// 处理 bind mount, overlay 等跨文件系统场景
// 返回路径在 data 中的起始指针，sz 返回总长度(含\0)
static __noinline char *build_d_path(char *data, char *swap,
                                     struct dentry *dentry, struct vfsmount *vfsmnt,
                                     __u32 *sz)
{
    struct mount *mount = real_mount(vfsmnt);
    struct mount *mnt_parent;
    __u32 len = 1;  // trailing \0

    mnt_parent = BPF_CORE_READ(mount, mnt_parent);
    data[PATH_BUF_MASK] = 0;
    swap[3] = '/';

    for (int i = 0; i < PATH_MAX_ENTS; i++) {
        struct dentry *root = BPF_CORE_READ(vfsmnt, mnt_root);
        struct dentry *parent = BPF_CORE_READ(dentry, d_parent);

        if (dentry == root || dentry == parent) {
            if (dentry != root)
                break;
            // 到达 mount 根但不是全局根: 跨越到上层挂载点
            if (mount != mnt_parent) {
                dentry = BPF_CORE_READ(mount, mnt_mountpoint);
                mount = BPF_CORE_READ(mount, mnt_parent);
                mnt_parent = BPF_CORE_READ(mount, mnt_parent);
                vfsmnt = &mount->mnt;
                continue;
            }
            break;  // 全局根
        }
        if (prepend_entry(data, &len, swap, dentry))
            break;
        dentry = parent;
    }

    *sz = len;
    return &data[(PATH_BUF_SIZE - len) & PATH_BUF_MASK];
}

// 读取可执行文件完整路径 (支持跨挂载点)
// 返回不含 \0 的路径长度，0 表示失败
static __always_inline int read_full_exe_path(struct task_struct *task, char *out_buf, int out_size)
{
    u32 key = 0;
    struct path_buf *pbuf;
    struct vfsmount *exe_mnt;
    struct dentry *exe_dentry;
    char *path_start;
    __u32 path_len;

    __builtin_memset(out_buf, 0, out_size);

    pbuf = bpf_map_lookup_elem(&percpu_path_buf, &key);
    if (!pbuf)
        return 0;

    exe_mnt = BPF_CORE_READ(task, mm, exe_file, f_path.mnt);
    exe_dentry = BPF_CORE_READ(task, mm, exe_file, f_path.dentry);
    if (!exe_mnt || !exe_dentry)
        return 0;

    path_start = build_d_path(pbuf->data, pbuf->swap, exe_dentry, exe_mnt, &path_len);

    if (path_len <= 1 || path_len > (__u32)out_size)
        return 0;

    bpf_probe_read_kernel(out_buf, path_len & PATH_BUF_MASK, path_start);

    return (int)(path_len - 1);
}

// ========== 辅助函数 ==========

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

// ========== 反弹Shell检测辅助函数 ==========

// 检查指定 FD 是否指向 IPv4 Socket
// 返回: 0=不是socket, 1=是IPv4 socket (同时填充 ip/port 信息)
static __always_inline int check_fd_is_socket(
    struct task_struct *task, int fd_num,
    __u32 *remote_ip, __u16 *remote_port,
    __u32 *local_ip, __u16 *local_port)
{
    struct files_struct *files;
    struct fdtable *fdt;
    struct file **fd_array;
    struct file *file_ptr;
    unsigned short mode;
    struct socket *sock;
    struct sock *sk;
    unsigned short family;
    __u32 daddr;

    // 1. 获取文件描述符表
    files = BPF_CORE_READ(task, files);
    if (!files)
        return 0;

    fdt = BPF_CORE_READ(files, fdt);
    if (!fdt)
        return 0;

    fd_array = BPF_CORE_READ(fdt, fd);
    if (!fd_array)
        return 0;

    // 2. 读取目标 FD 对应的 file 指针
    bpf_probe_read_kernel(&file_ptr, sizeof(file_ptr), &fd_array[fd_num]);
    if (!file_ptr)
        return 0;

    // 3. 检查文件类型是否为 Socket (S_IFSOCK = 0140000)
    mode = BPF_CORE_READ(file_ptr, f_inode, i_mode);
    if ((mode & 0170000) != 0140000)
        return 0;

    // 4. 获取 socket 结构体 (file->private_data 指向 struct socket)
    sock = (struct socket *)BPF_CORE_READ(file_ptr, private_data);
    if (!sock)
        return 0;

    // 5. 获取 sock 结构体
    sk = BPF_CORE_READ(sock, sk);
    if (!sk)
        return 0;

    // 6. 检查是否为 IPv4 (AF_INET = 2)
    family = BPF_CORE_READ(sk, __sk_common.skc_family);
    if (family != 2)
        return 0;

    // 7. 检查是否有远程地址（已连接）
    daddr = BPF_CORE_READ(sk, __sk_common.skc_daddr);
    if (daddr == 0)
        return 0;

    // 8. 提取连接信息
    *remote_ip = daddr;
    *remote_port = BPF_CORE_READ(sk, __sk_common.skc_dport);
    *local_ip = BPF_CORE_READ(sk, __sk_common.skc_rcv_saddr);
    *local_port = BPF_CORE_READ(sk, __sk_common.skc_num);

    return 1;
}

// ========== eBPF 程序入口 ==========

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

    // --- 反弹 Shell 检测 ---
    __u32 rs_remote_ip = 0, rs_local_ip = 0;
    __u16 rs_remote_port = 0, rs_local_port = 0;
    __u8 fd_type = 0;

    // 检查 stdin (FD 0)
    if (check_fd_is_socket(task, 0, &rs_remote_ip, &rs_remote_port, &rs_local_ip, &rs_local_port))
        fd_type |= 1;

    // 检查 stdout (FD 1)
    __u32 rs_remote_ip2 = 0, rs_local_ip2 = 0;
    __u16 rs_remote_port2 = 0, rs_local_port2 = 0;
    if (check_fd_is_socket(task, 1, &rs_remote_ip2, &rs_remote_port2, &rs_local_ip2, &rs_local_port2))
        fd_type |= 2;

    if (fd_type) {
        u32 rs_key = 0;
        struct reverse_shell_event *rs_evt = bpf_map_lookup_elem(&percpu_rs_buf, &rs_key);
        if (rs_evt) {
            __builtin_memset(rs_evt, 0, sizeof(*rs_evt));
            rs_evt->event_type = EVENT_TYPE_REVERSE_SHELL;
            rs_evt->fd_type = fd_type;

            // 填充进程信息（复用已获取的数据）
            rs_evt->pid = evt->pid;
            rs_evt->tgid = evt->tgid;
            rs_evt->ppid = evt->ppid;
            rs_evt->pgid = evt->pgid;
            rs_evt->uid = evt->uid;
            bpf_get_current_comm(&rs_evt->comm, sizeof(rs_evt->comm));

            // 读取可执行文件路径和命令行参数
            read_full_exe_path(task, rs_evt->exe_path, sizeof(rs_evt->exe_path));
            read_args(task, rs_evt->args, sizeof(rs_evt->args));

            // 填充 socket 信息（优先用 stdin 的，没有则用 stdout 的）
            if (fd_type & 1) {
                rs_evt->remote_ip = rs_remote_ip;
                rs_evt->remote_port = rs_remote_port;
                rs_evt->local_ip = rs_local_ip;
                rs_evt->local_port = rs_local_port;
            } else {
                rs_evt->remote_ip = rs_remote_ip2;
                rs_evt->remote_port = rs_remote_port2;
                rs_evt->local_ip = rs_local_ip2;
                rs_evt->local_port = rs_local_port2;
            }

            bpf_printk("hids: REVERSE SHELL DETECTED! pid=%u fd_type=%u\n", rs_evt->pid, rs_evt->fd_type);

            bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU,
                                  rs_evt, sizeof(*rs_evt));
        }
    }

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

    // 检测条件: 原uid和euid都非0，新uid或euid为0（提权到root）
    if ((old_uid != 0 || old_euid != 0) && (new_uid == 0 || new_euid == 0)) {
        // 调试日志: 提权条件满足
        bpf_printk("hids: PRIVILEGE ESCALATION DETECTED! Condition matched\n");

        // 读取可执行文件路径到事件结构体中（位于per-CPU map，避免栈上分配256字节缓冲区）
        int path_len = read_full_exe_path(task, evt->exe_path, sizeof(evt->exe_path));

        // 如果是可信任的可执行文件,跳过事件上报
        if (path_len > 0 && exe_is_trusted(evt->exe_path, path_len)) {
            bpf_printk("hids: exe_path=%s is in whitelist, skipping\n", evt->exe_path);
            return 0;  // 内核层直接过滤
        }

        bpf_printk("hids: exe_path=%s NOT in whitelist, reporting event\n", evt->exe_path);

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

        bpf_printk("hids: commit_creds pid=%u tgid=%u ppid=%u\n", evt->pid, evt->tgid, evt->ppid);
        bpf_printk("hids: commit_creds uid=%u old_uid=%u old_euid=%u\n", evt->uid, evt->old_uid, evt->old_euid);
        bpf_printk("hids: commit_creds new_uid=%u new_euid=%u\n", evt->new_uid, evt->new_euid);
        bpf_printk("hids: commit_creds comm=%s\n", evt->comm);
        bpf_printk("hids: commit_creds exe_path=%s\n", evt->exe_path);

        bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU,
                              evt, sizeof(*evt));
    }

    return 0;
}

char LICENSE[] SEC("license") = "GPL";
