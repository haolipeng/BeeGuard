// SPDX-License-Identifier: GPL-2.0
#include <vmlinux.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_endian.h>
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

// Per-CPU Array Map for connect_event - 用于出站连接监控
struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __type(key, u32);
    __type(value, struct connect_event);
    __uint(max_entries, 1);
} percpu_connect_buf SEC(".maps");

// Per-CPU Array Map for bind_event - 用于端口绑定监控
struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __type(key, u32);
    __type(value, struct bind_event);
    __uint(max_entries, 1);
} percpu_bind_buf SEC(".maps");

// Per-CPU Array Map for accept_event - 用于入站连接监控
struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __type(key, u32);
    __type(value, struct accept_event);
    __uint(max_entries, 1);
} percpu_accept_buf SEC(".maps");

// Per-CPU Array Map for dns_event - 用于DNS查询监控
struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __type(key, u32);
    __type(value, struct dns_event);
    __uint(max_entries, 1);
} percpu_dns_buf SEC(".maps");

// 系统调用参数保存 Map (sys_enter → sys_exit 传递)
// key: pid_tgid, value: syscall fd 参数
struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __uint(max_entries, 4096);
    __type(key, __u64);     // pid_tgid
    __type(value, __u64);   // fd (syscall 第一个参数)
} syscall_fd_map SEC(".maps");

// 系统调用 sockaddr 参数保存 Map
// key: pid_tgid, value: sockaddr 指针 (用户态地址)
struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __uint(max_entries, 4096);
    __type(key, __u64);     // pid_tgid
    __type(value, __u64);   // sockaddr 用户态指针
} syscall_sockaddr_map SEC(".maps");

// recvfrom/recvmsg 用户态 buffer 参数保存 Map
struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __uint(max_entries, 4096);
    __type(key, __u64);     // pid_tgid
    __type(value, __u64);   // 用户态 buffer 指针
} syscall_buf_map SEC(".maps");

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

// ========== 网络监控辅助函数 ==========

// 通过 fd 获取 socket 的 sock 结构体
// 返回 NULL 表示 fd 不是 AF_INET socket
static __always_inline struct sock *get_sock_from_fd(struct task_struct *task, int fd_num)
{
    struct files_struct *files;
    struct fdtable *fdt;
    struct file **fd_array;
    struct file *file_ptr;
    unsigned short mode;
    struct socket *sock;
    struct sock *sk;
    unsigned short family;

    files = BPF_CORE_READ(task, files);
    if (!files)
        return NULL;

    fdt = BPF_CORE_READ(files, fdt);
    if (!fdt)
        return NULL;

    fd_array = BPF_CORE_READ(fdt, fd);
    if (!fd_array)
        return NULL;

    bpf_probe_read_kernel(&file_ptr, sizeof(file_ptr), &fd_array[fd_num]);
    if (!file_ptr)
        return NULL;

    // 检查是否为 Socket 文件
    mode = BPF_CORE_READ(file_ptr, f_inode, i_mode);
    if ((mode & 0170000) != 0140000)
        return NULL;

    sock = (struct socket *)BPF_CORE_READ(file_ptr, private_data);
    if (!sock)
        return NULL;

    sk = BPF_CORE_READ(sock, sk);
    if (!sk)
        return NULL;

    // 仅处理 IPv4
    family = BPF_CORE_READ(sk, __sk_common.skc_family);
    if (family != 2)  // AF_INET = 2
        return NULL;

    return sk;
}

// 获取 socket 协议类型 (TCP=6, UDP=17)
static __always_inline __u8 get_sock_protocol(struct sock *sk)
{
    // sk_protocol 在 sock_common 中
    __u8 proto = BPF_CORE_READ(sk, sk_protocol);
    return proto;
}

// 填充通用进程信息到事件中（使用局部变量避免 packed struct 地址警告）
#define FILL_PROCESS_INFO(task, evt) do { \
    u64 _id = bpf_get_current_pid_tgid(); \
    (evt)->pid = _id; \
    (evt)->tgid = _id >> 32; \
    (evt)->uid = bpf_get_current_uid_gid(); \
    struct task_struct *_parent = BPF_CORE_READ(task, real_parent); \
    if (_parent) \
        (evt)->ppid = BPF_CORE_READ(_parent, tgid); \
    bpf_get_current_comm(&(evt)->comm, 16); \
} while (0)

// ========== sys_enter: 保存系统调用参数 ==========

// 拦截 connect/bind/accept4 的 sys_enter，保存 fd 参数
SEC("raw_tracepoint/sys_enter")
int tp_sys_enter(struct bpf_raw_tracepoint_args *ctx)
{
    // raw_tracepoint/sys_enter 参数:
    // ctx->args[0] = struct pt_regs *regs
    // ctx->args[1] = long syscall_nr
    unsigned long syscall_nr = ctx->args[1];
    struct pt_regs *regs = (struct pt_regs *)ctx->args[0];

    // 仅处理网络相关系统调用
    // x86_64: connect=42, bind=49, accept=43, accept4=288, recvfrom=45, recvmsg=47
    if (syscall_nr != 42 && syscall_nr != 49 && syscall_nr != 43 &&
        syscall_nr != 288 && syscall_nr != 45 && syscall_nr != 47)
        return 0;

    __u64 pid_tgid = bpf_get_current_pid_tgid();

    // 获取第一个参数 (fd) - x86_64: rdi
    __u64 fd = 0;
    bpf_probe_read_kernel(&fd, sizeof(fd), &regs->di);

    // 保存 fd 参数
    bpf_map_update_elem(&syscall_fd_map, &pid_tgid, &fd, BPF_ANY);

    // 对于 connect 和 bind，还需保存 sockaddr 指针 (第二个参数 rsi)
    if (syscall_nr == 42 || syscall_nr == 49) {
        __u64 sockaddr_ptr = 0;
        bpf_probe_read_kernel(&sockaddr_ptr, sizeof(sockaddr_ptr), &regs->si);
        bpf_map_update_elem(&syscall_sockaddr_map, &pid_tgid, &sockaddr_ptr, BPF_ANY);
    }

    // 对于 recvfrom，保存 buffer 指针 (第二个参数 rsi)
    if (syscall_nr == 45) {
        __u64 buf_ptr = 0;
        bpf_probe_read_kernel(&buf_ptr, sizeof(buf_ptr), &regs->si);
        bpf_map_update_elem(&syscall_buf_map, &pid_tgid, &buf_ptr, BPF_ANY);
    }

    // 对于 recvmsg，保存 msghdr 指针 (第二个参数 rsi)
    if (syscall_nr == 47) {
        __u64 msghdr_ptr = 0;
        bpf_probe_read_kernel(&msghdr_ptr, sizeof(msghdr_ptr), &regs->si);
        bpf_map_update_elem(&syscall_buf_map, &pid_tgid, &msghdr_ptr, BPF_ANY);
    }

    return 0;
}

// ========== sys_exit: 处理系统调用返回 ==========

// DNS 域名解析辅助函数
// 从已读入内核栈的 DNS 包缓冲区中提取查询域名和查询类型
// DNS 报文格式: Header(12字节) + Query(变长)
// 域名格式: 3www6google3com0 → www.google.com
// 注意: buf 是内核栈空间地址（已通过 bpf_probe_read_user 读入）
static __noinline int parse_dns_query(char *buf, int buf_len, char *domain, int domain_size, __u16 *query_type)
{
    // DNS header 至少 12 字节
    if (buf_len < 12)
        return -1;

    // 跳过 DNS header (12 字节)
    int offset = 12;
    int domain_offset = 0;

    // 解析 Query 部分的域名 (最多 32 个 label)
    for (int i = 0; i < 32; i++) {
        if (offset >= buf_len || offset >= 256)
            return -1;

        __u8 label_len = (__u8)buf[offset];

        // label_len == 0 表示域名结束
        if (label_len == 0) {
            offset += 1;
            break;
        }

        // 检查指针压缩 (高2位为11)
        if ((label_len & 0xC0) == 0xC0)
            return -1;  // 简化版不处理压缩指针

        // 安全检查
        if (label_len > 63)
            return -1;
        if (offset + 1 + label_len > buf_len || offset + 1 + label_len > 256)
            return -1;

        // 添加 '.' 分隔符（非第一个 label）
        if (domain_offset > 0 && domain_offset < domain_size - 1) {
            domain[domain_offset] = '.';
            domain_offset++;
        }

        // 逐字节复制 label（满足 BPF 验证器的有界访问要求）
        int copy_len = label_len;
        if (domain_offset + copy_len >= domain_size - 1)
            copy_len = domain_size - 1 - domain_offset;
        if (copy_len <= 0)
            break;
        if (copy_len > 63)
            copy_len = 63;

        for (int j = 0; j < 63 && j < copy_len; j++) {
            int src_idx = offset + 1 + j;
            int dst_idx = domain_offset + j;
            if (src_idx >= 256 || dst_idx >= domain_size - 1)
                break;
            domain[dst_idx] = buf[src_idx];
        }

        domain_offset += copy_len;
        offset += 1 + label_len;
    }

    // 终止域名字符串
    if (domain_offset >= 0 && domain_offset < domain_size)
        domain[domain_offset] = 0;

    // 读取 query type (域名后的 2 字节, 网络字节序)
    if (offset + 2 <= buf_len && offset + 2 <= 256) {
        *query_type = ((__u16)(__u8)buf[offset] << 8) | (__u8)buf[offset + 1];
    }

    return 0;
}

SEC("raw_tracepoint/sys_exit")
int tp_sys_exit(struct bpf_raw_tracepoint_args *ctx)
{
    // raw_tracepoint/sys_exit 参数:
    // ctx->args[0] = struct pt_regs *regs
    // ctx->args[1] = long ret
    struct pt_regs *regs = (struct pt_regs *)ctx->args[0];
    long retval = ctx->args[1];

    // 获取 syscall nr (从 regs->orig_ax)
    unsigned long syscall_nr = 0;
    bpf_probe_read_kernel(&syscall_nr, sizeof(syscall_nr), &regs->orig_ax);

    // 仅处理网络相关系统调用
    if (syscall_nr != 42 && syscall_nr != 49 && syscall_nr != 43 &&
        syscall_nr != 288 && syscall_nr != 45 && syscall_nr != 47)
        return 0;

    __u64 pid_tgid = bpf_get_current_pid_tgid();
    struct task_struct *task = (struct task_struct *)bpf_get_current_task();

    // 查找保存的 fd 参数
    __u64 *fd_ptr = bpf_map_lookup_elem(&syscall_fd_map, &pid_tgid);
    if (!fd_ptr)
        return 0;
    int fd = (int)*fd_ptr;

    // ========== connect (syscall 42) ==========
    if (syscall_nr == 42) {
        // 清理保存的参数
        bpf_map_delete_elem(&syscall_fd_map, &pid_tgid);

        // 获取 sockaddr 指针
        __u64 *sa_ptr = bpf_map_lookup_elem(&syscall_sockaddr_map, &pid_tgid);
        if (!sa_ptr) goto cleanup_connect;
        __u64 sockaddr_uptr = *sa_ptr;
        bpf_map_delete_elem(&syscall_sockaddr_map, &pid_tgid);

        // 读取 sockaddr_in (sa_family + sin_port + sin_addr)
        // struct sockaddr_in { sa_family_t sin_family; __be16 sin_port; struct in_addr sin_addr; }
        __u16 sa_family = 0;
        bpf_probe_read_user(&sa_family, sizeof(sa_family), (void *)sockaddr_uptr);
        if (sa_family != 2)  // AF_INET only
            return 0;

        __u16 sin_port = 0;
        bpf_probe_read_user(&sin_port, sizeof(sin_port), (void *)(sockaddr_uptr + 2));

        __u32 sin_addr = 0;
        bpf_probe_read_user(&sin_addr, sizeof(sin_addr), (void *)(sockaddr_uptr + 4));

        // 过滤 loopback 和全零地址
        // sin_addr 是网络字节序，127.0.0.1 = 0x0100007f
        if (sin_addr == 0)
            return 0;

        u32 key = 0;
        struct connect_event *evt = bpf_map_lookup_elem(&percpu_connect_buf, &key);
        if (!evt)
            return 0;

        __builtin_memset(evt, 0, sizeof(*evt));
        evt->event_type = EVENT_TYPE_CONNECT;

        FILL_PROCESS_INFO(task, evt);

        evt->remote_ip = sin_addr;
        evt->remote_port = sin_port;
        evt->retval = (__s32)retval;

        // 通过 socket 获取本地地址和协议
        struct sock *sk = get_sock_from_fd(task, fd);
        if (sk) {
            evt->local_ip = BPF_CORE_READ(sk, __sk_common.skc_rcv_saddr);
            evt->local_port = BPF_CORE_READ(sk, __sk_common.skc_num);
            evt->protocol = get_sock_protocol(sk);
        }

        read_full_exe_path(task, evt->exe_path, sizeof(evt->exe_path));

        bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, evt, sizeof(*evt));
        return 0;

    cleanup_connect:
        bpf_map_delete_elem(&syscall_sockaddr_map, &pid_tgid);
        return 0;
    }

    // ========== bind (syscall 49) ==========
    if (syscall_nr == 49) {
        bpf_map_delete_elem(&syscall_fd_map, &pid_tgid);

        __u64 *sa_ptr = bpf_map_lookup_elem(&syscall_sockaddr_map, &pid_tgid);
        if (!sa_ptr) goto cleanup_bind;
        __u64 sockaddr_uptr = *sa_ptr;
        bpf_map_delete_elem(&syscall_sockaddr_map, &pid_tgid);

        // 仅处理成功的 bind (retval == 0)
        if (retval != 0)
            return 0;

        __u16 sa_family = 0;
        bpf_probe_read_user(&sa_family, sizeof(sa_family), (void *)sockaddr_uptr);
        if (sa_family != 2)
            return 0;

        __u16 sin_port = 0;
        bpf_probe_read_user(&sin_port, sizeof(sin_port), (void *)(sockaddr_uptr + 2));

        __u32 sin_addr = 0;
        bpf_probe_read_user(&sin_addr, sizeof(sin_addr), (void *)(sockaddr_uptr + 4));

        u32 key = 0;
        struct bind_event *evt = bpf_map_lookup_elem(&percpu_bind_buf, &key);
        if (!evt)
            return 0;

        __builtin_memset(evt, 0, sizeof(*evt));
        evt->event_type = EVENT_TYPE_BIND;

        FILL_PROCESS_INFO(task, evt);

        evt->bind_ip = sin_addr;
        evt->bind_port = sin_port;
        evt->retval = (__s32)retval;

        struct sock *sk = get_sock_from_fd(task, fd);
        if (sk)
            evt->protocol = get_sock_protocol(sk);

        read_full_exe_path(task, evt->exe_path, sizeof(evt->exe_path));

        bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, evt, sizeof(*evt));
        return 0;

    cleanup_bind:
        bpf_map_delete_elem(&syscall_sockaddr_map, &pid_tgid);
        return 0;
    }

    // ========== accept/accept4 (syscall 43/288) ==========
    if (syscall_nr == 43 || syscall_nr == 288) {
        bpf_map_delete_elem(&syscall_fd_map, &pid_tgid);

        // retval 是新的 fd（accept 成功返回新 fd，失败返回负数）
        if (retval < 0)
            return 0;

        int new_fd = (int)retval;

        // 从新 fd 获取 socket 信息
        struct sock *sk = get_sock_from_fd(task, new_fd);
        if (!sk)
            return 0;

        u32 key = 0;
        struct accept_event *evt = bpf_map_lookup_elem(&percpu_accept_buf, &key);
        if (!evt)
            return 0;

        __builtin_memset(evt, 0, sizeof(*evt));
        evt->event_type = EVENT_TYPE_ACCEPT;

        FILL_PROCESS_INFO(task, evt);

        evt->remote_ip = BPF_CORE_READ(sk, __sk_common.skc_daddr);
        evt->remote_port = BPF_CORE_READ(sk, __sk_common.skc_dport);
        evt->local_ip = BPF_CORE_READ(sk, __sk_common.skc_rcv_saddr);
        evt->local_port = BPF_CORE_READ(sk, __sk_common.skc_num);
        evt->protocol = get_sock_protocol(sk);
        evt->retval = (__s32)retval;

        // 过滤非 IPv4 连接（remote_ip 为 0 表示没有对端）
        if (evt->remote_ip == 0)
            return 0;

        read_full_exe_path(task, evt->exe_path, sizeof(evt->exe_path));

        bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, evt, sizeof(*evt));
        return 0;
    }

    // ========== recvfrom (syscall 45) / recvmsg (syscall 47) — DNS 监控 ==========
    if (syscall_nr == 45 || syscall_nr == 47) {
        bpf_map_delete_elem(&syscall_fd_map, &pid_tgid);

        // 失败则跳过
        if (retval <= 0)
            goto cleanup_dns;

        // 检查 socket 源端口是否为 53 (DNS) 或 5353 (mDNS)
        struct sock *sk = get_sock_from_fd(task, fd);
        if (!sk)
            goto cleanup_dns;

        __u16 src_port = BPF_CORE_READ(sk, __sk_common.skc_dport);
        // skc_dport 是网络字节序, 53=0x0035, 5353=0x14E9
        // 网络字节序 (big-endian): 53 → 0x0035, 5353 → 0x14E9
        if (src_port != __bpf_htons(53) && src_port != __bpf_htons(5353))
            goto cleanup_dns;

        // 获取保存的用户态 buffer 指针
        __u64 *buf_ptr = bpf_map_lookup_elem(&syscall_buf_map, &pid_tgid);
        if (!buf_ptr)
            goto cleanup_dns;
        __u64 user_buf = *buf_ptr;
        bpf_map_delete_elem(&syscall_buf_map, &pid_tgid);

        // 对于 recvmsg (47)，需要从 msghdr 中获取实际的 buffer 地址
        // struct msghdr { struct iovec *msg_iov; ... }
        // struct iovec { void *iov_base; size_t iov_len; }
        if (syscall_nr == 47) {
            // user_buf 指向 struct msghdr (用户态)
            // msghdr 第二个字段 msg_iov (偏移 8 字节在 x86_64)
            __u64 msg_iov = 0;
            // msg_iov 在 user_msghdr 中的偏移 (跳过 msg_name + msg_namelen)
            bpf_probe_read_user(&msg_iov, sizeof(msg_iov), (void *)(user_buf + 8));
            if (!msg_iov)
                return 0;
            // iov_base 是 iovec 的第一个字段
            __u64 iov_base = 0;
            bpf_probe_read_user(&iov_base, sizeof(iov_base), (void *)msg_iov);
            if (!iov_base)
                return 0;
            user_buf = iov_base;
        }

        // 读取 DNS 包到 per-CPU buffer 的字段，避免栈溢出
        u32 dns_key = 0;
        struct dns_event *evt = bpf_map_lookup_elem(&percpu_dns_buf, &dns_key);
        if (!evt)
            return 0;

        __builtin_memset(evt, 0, sizeof(*evt));
        evt->event_type = EVENT_TYPE_DNS;

        int read_len = retval;
        if (read_len > (int)sizeof(evt->domain))
            read_len = sizeof(evt->domain);
        if (read_len < 12)  // DNS header 至少 12 字节
            return 0;

        // 读取 DNS 包到 exe_path 字段作为临���缓冲区（256 字节，在 per-CPU map 中）
        // exe_path 稍后会被 read_full_exe_path 覆写，所以先借用
        bpf_probe_read_user(evt->exe_path, read_len & 0xFF, (void *)user_buf);

        // 提取 DNS header 中的 opcode 和 rcode
        // DNS header: ID(2) + Flags(2) + ...
        // Flags: QR(1) OPCODE(4) AA(1) TC(1) RD(1) RA(1) Z(3) RCODE(4)
        __u8 flags1 = evt->exe_path[2];
        __u8 flags2 = evt->exe_path[3];
        evt->opcode = (flags1 >> 3) & 0x0F;
        evt->rcode = flags2 & 0x0F;

        // 从 exe_path (临时DNS数据) 解析域名到 domain 字段
        __u16 qtype = 0;
        if (parse_dns_query(evt->exe_path, read_len, evt->domain, sizeof(evt->domain), &qtype) < 0)
            return 0;
        evt->query_type = qtype;

        FILL_PROCESS_INFO(task, evt);

        evt->dns_server_ip = BPF_CORE_READ(sk, __sk_common.skc_daddr);
        evt->dns_server_port = src_port;

        read_full_exe_path(task, evt->exe_path, sizeof(evt->exe_path));

        bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, evt, sizeof(*evt));
        return 0;

    cleanup_dns:
        bpf_map_delete_elem(&syscall_buf_map, &pid_tgid);
        return 0;
    }

    // 清理未匹配的 map 条目
    bpf_map_delete_elem(&syscall_fd_map, &pid_tgid);
    return 0;
}

char LICENSE[] SEC("license") = "GPL";
