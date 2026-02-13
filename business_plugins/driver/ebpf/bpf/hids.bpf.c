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

// Per-CPU buffer for stdin/stdout path construction - 避免与 exe_path 冲突
struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __type(key, u32);
    __type(value, struct stdio_path_buf);
    __uint(max_entries, 1);
} percpu_stdio_path_buf SEC(".maps");

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

// Per-CPU Array Map for DNS raw data buffer - 避免栈溢出
struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __type(key, u32);
    __type(value, struct dns_data_buf);
    __uint(max_entries, 1);
} percpu_dns_data SEC(".maps");

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
    if (swap[4] != '/')
        rc = prepend_path(data, len, &swap[3], rc);
    else if (rc > 2)
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
            if (mount != mnt_parent) {
                dentry = BPF_CORE_READ(mount, mnt_mountpoint);
                mount = BPF_CORE_READ(mount, mnt_parent);
                mnt_parent = BPF_CORE_READ(mount, mnt_parent);
                vfsmnt = &mount->mnt;
                continue;
            }
            break;
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

    arg_start = BPF_CORE_READ(task, mm, arg_start);
    arg_end = BPF_CORE_READ(task, mm, arg_end);

    if (!arg_start || !arg_end || arg_start >= arg_end)
        return 0;

    args_len = arg_end - arg_start;

    if (args_len <= 0 || args_len > 4096)
        return 0;

    if (args_len > buf_size)
        args_len = buf_size;

    if (args_len <= 0 || args_len > buf_size)
        return 0;

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

    if (len > 256)
        len = 256;

    for (i = 0; i < 256 && i < len; i++) {
        h ^= (__u64)(s[i]);
        h *= 0x5bd1e9955bd1e995;
        h ^= h >> 47;
    }

    return h;
}

// 检查可执行文件是否可信任
static __noinline int exe_is_trusted(const char *exe_path, int path_len)
{
    __u64 hash = hash_murmur_OAAT64(exe_path, path_len);
    struct exe_item *ei;

    ei = bpf_map_lookup_elem(&trusted_exes, &hash);

    return (ei && ei->len == path_len);
}

// ========== 反弹 Shell 增强采集辅助函数 ==========

// read_fd_path: 从 task 的指定 FD 获取文件路径（通过 dentry 遍历）
// 使用 percpu_stdio_path_buf 作为工作缓冲区
// 返回路径长度（不含 \0），0 表示失败
static __noinline int read_fd_path(struct task_struct *task, int fd,
                                    char *out_buf, int out_size)
{
    u32 key = 0;
    struct stdio_path_buf *pbuf;
    struct files_struct *files;
    struct fdtable *fdt;
    struct file **fd_array;
    struct file *file_ptr;
    struct vfsmount *mnt;
    struct dentry *dentry;
    char *path_start;
    __u32 path_len;

    __builtin_memset(out_buf, 0, out_size);

    pbuf = bpf_map_lookup_elem(&percpu_stdio_path_buf, &key);
    if (!pbuf)
        return 0;

    files = BPF_CORE_READ(task, files);
    if (!files)
        return 0;

    fdt = BPF_CORE_READ(files, fdt);
    if (!fdt)
        return 0;

    if (fd >= (int)BPF_CORE_READ(fdt, max_fds))
        return 0;

    fd_array = BPF_CORE_READ(fdt, fd);
    if (!fd_array)
        return 0;

    bpf_probe_read_kernel(&file_ptr, sizeof(file_ptr), &fd_array[fd]);
    if (!file_ptr)
        return 0;

    mnt = BPF_CORE_READ(file_ptr, f_path.mnt);
    dentry = BPF_CORE_READ(file_ptr, f_path.dentry);
    if (!mnt || !dentry)
        return 0;

    path_start = build_d_path(pbuf->data, pbuf->swap, dentry, mnt, &path_len);

    if (path_len <= 1 || path_len > (__u32)out_size)
        return 0;

    bpf_probe_read_kernel(out_buf, path_len & (STDIO_PATH_LEN - 1), path_start);

    return (int)(path_len - 1);
}

// find_sockfd: 扫描当前进程的 FD 0-15，寻找连接状态的 IPv4 socket
// 仅扫描当前进程（不遍历父进程链），返回第一个找到的 sock 指针
// 返回: sock 指针（NULL=未找到）
static __noinline struct sock *find_sockfd(struct task_struct *task)
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
    int max_fds;

    files = BPF_CORE_READ(task, files);
    if (!files)
        return NULL;

    fdt = BPF_CORE_READ(files, fdt);
    if (!fdt)
        return NULL;

    fd_array = BPF_CORE_READ(fdt, fd);
    if (!fd_array)
        return NULL;

    max_fds = BPF_CORE_READ(fdt, max_fds);

    for (int fd = 0; fd < SOCK_FD_LIMIT; fd++) {
        if (fd >= max_fds)
            break;

        bpf_probe_read_kernel(&file_ptr, sizeof(file_ptr), &fd_array[fd]);
        if (!file_ptr)
            continue;

        mode = BPF_CORE_READ(file_ptr, f_inode, i_mode);
        if ((mode & 0170000) != 0140000)
            continue;

        sock = (struct socket *)BPF_CORE_READ(file_ptr, private_data);
        if (!sock)
            continue;

        sk = BPF_CORE_READ(sock, sk);
        if (!sk)
            continue;

        family = BPF_CORE_READ(sk, __sk_common.skc_family);
        if (family != 2)
            continue;

        daddr = BPF_CORE_READ(sk, __sk_common.skc_daddr);
        if (daddr == 0)
            continue;

        return sk;
    }

    return NULL;
}

// ========== eBPF 程序入口 ==========

// handle_execve_event: 采集并输出 execve 事件（包含 socket 信息 enrich）
// 独立 __noinline 函数，SEC 入口极薄，重逻辑在 subprogram 中
// BPF verifier 对每个 subprogram 独立验证，各有 1M 指令预算
// 反弹 Shell 检测由用户态 Go 程序完成
static __noinline int handle_execve_event(
    struct bpf_raw_tracepoint_args *ctx)
{
    u32 key = 0;
    struct task_struct *task;
    struct task_struct *parent;

    struct execve_event *evt = bpf_map_lookup_elem(&percpu_buf, &key);
    if (!evt)
        return 0;

    evt->event_type = EVENT_TYPE_EXECVE;
    evt->fd_type = 0;
    evt->ppid = 0;
    evt->socket_pid = 0;
    evt->remote_ip = 0;
    evt->remote_port = 0;
    evt->local_port = 0;
    evt->local_ip = 0;

    task = (struct task_struct *)bpf_get_current_task();

    u64 id = bpf_get_current_pid_tgid();
    evt->pid = id;
    evt->tgid = id >> 32;

    parent = BPF_CORE_READ(task, real_parent);
    if (parent) {
        evt->ppid = BPF_CORE_READ(parent, tgid);
    }

    evt->pgid = BPF_CORE_READ(task, tgid);

    u64 uid_gid = bpf_get_current_uid_gid();
    evt->uid = uid_gid;

    bpf_get_current_comm(&evt->comm, sizeof(evt->comm));

    read_full_exe_path(task, evt->exe_path, sizeof(evt->exe_path));

    read_args(task, evt->args, sizeof(evt->args));

    read_fd_path(task, 0, evt->stdin_path, sizeof(evt->stdin_path));
    read_fd_path(task, 1, evt->stdout_path, sizeof(evt->stdout_path));

    __builtin_memset(evt->tty_name, 0, sizeof(evt->tty_name));
    {
        struct signal_struct *_sig = BPF_CORE_READ(task, signal);
        if (_sig) {
            struct tty_struct *_tty = BPF_CORE_READ(_sig, tty);
            if (_tty) {
                bpf_probe_read_kernel_str(evt->tty_name,
                    sizeof(evt->tty_name) & (TTY_NAME_LEN - 1),
                    (void *)BPF_CORE_READ(_tty, name));
            }
        }
    }

    // TODO: 可扩展到扫描当前进程的父进程 FD 0-15的信息
    struct sock *sk = find_sockfd(task);
    if (sk) {
        evt->socket_pid = evt->tgid;
        evt->remote_ip = BPF_CORE_READ(sk, __sk_common.skc_daddr);
        evt->remote_port = BPF_CORE_READ(sk, __sk_common.skc_dport);
        evt->local_ip = BPF_CORE_READ(sk, __sk_common.skc_rcv_saddr);
        evt->local_port = BPF_CORE_READ(sk, __sk_common.skc_num);
    }


    bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU,
                          evt, sizeof(*evt));

    return 0;
}

// 监听进程执行事件
// Hook点: sched_process_exec - 在execve系统调用成功后触发
// SEC 入口极薄，所有逻辑在 handle_execve_event subprogram 中
SEC("raw_tracepoint/sched_process_exec")
int tp_proc_exec(struct bpf_raw_tracepoint_args *ctx)
{
    return handle_execve_event(ctx);
}

// 监听commit_creds调用，检测本地提权行为
// Hook点: kprobe/commit_creds - 在凭证提交前触发
// 检测条件: 原uid和euid都非0，新uid或euid为0（非root -> root）
SEC("kprobe/commit_creds")
int kp_commit_creds(struct pt_regs *ctx)
{
    u32 key = 0;
    struct task_struct *task;
    struct task_struct *parent;

    struct commit_creds_event *evt = bpf_map_lookup_elem(&percpu_creds_buf, &key);
    if (!evt)
        return 0;

    __builtin_memset(evt, 0, sizeof(*evt));
    evt->event_type = EVENT_TYPE_COMMIT_CREDS;

    task = (struct task_struct *)bpf_get_current_task();

    struct cred *new_cred = (void *)PT_REGS_PARM1_CORE(ctx);
    if (!new_cred)
        return 0;

    int old_uid = BPF_CORE_READ(task, real_cred, uid.val);
    int old_euid = BPF_CORE_READ(task, real_cred, euid.val);

    int new_uid = BPF_CORE_READ(new_cred, uid.val);
    int new_euid = BPF_CORE_READ(new_cred, euid.val);

    if ((old_uid != 0 || old_euid != 0) && (new_uid == 0 || new_euid == 0)) {
#ifdef DEBUG_PRINT
        bpf_printk("hids: PRIVILEGE ESCALATION DETECTED! Condition matched\n");
#endif

        int path_len = read_full_exe_path(task, evt->exe_path, sizeof(evt->exe_path));

        if (path_len > 0 && exe_is_trusted(evt->exe_path, path_len)) {
#ifdef DEBUG_PRINT
            bpf_printk("hids: exe_path=%s is in whitelist, skipping\n", evt->exe_path);
#endif
            return 0;
        }

#ifdef DEBUG_PRINT
        bpf_printk("hids: exe_path=%s NOT in whitelist, reporting event\n", evt->exe_path);
#endif

        u64 id = bpf_get_current_pid_tgid();
        evt->pid = id;
        evt->tgid = id >> 32;

        parent = BPF_CORE_READ(task, real_parent);
        if (parent) {
            evt->ppid = BPF_CORE_READ(parent, tgid);
        }

        evt->uid = bpf_get_current_uid_gid();

        evt->old_uid = old_uid;
        evt->old_euid = old_euid;
        evt->new_uid = new_uid;
        evt->new_euid = new_euid;

        bpf_get_current_comm(&evt->comm, sizeof(evt->comm));

#ifdef DEBUG_PRINT
        bpf_printk("hids: commit_creds pid=%u tgid=%u ppid=%u\n", evt->pid, evt->tgid, evt->ppid);
        bpf_printk("hids: commit_creds uid=%u old_uid=%u old_euid=%u\n", evt->uid, evt->old_uid, evt->old_euid);
        bpf_printk("hids: commit_creds new_uid=%u new_euid=%u\n", evt->new_uid, evt->new_euid);
        bpf_printk("hids: commit_creds comm=%s\n", evt->comm);
        bpf_printk("hids: commit_creds exe_path=%s\n", evt->exe_path);
#endif

        bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU,
                              evt, sizeof(*evt));
    }

    return 0;
}

// ========== 网络监控辅助函数 ==========
// 从 fd 获取 file 结构体
static __noinline struct file *fget_raw(struct task_struct *task, int nr)
{
    if (nr < 0 || nr >= 1024)
        return NULL;
    struct files_struct *files = BPF_CORE_READ(task, files);
    if (!files) return NULL;
    struct fdtable *fdt = BPF_CORE_READ(files, fdt);
    if (!fdt) return NULL;
    if (nr >= (int)BPF_CORE_READ(fdt, max_fds))
        return NULL;
    struct file **fds = BPF_CORE_READ(fdt, fd);
    if (!fds) return NULL;
    struct file *file = NULL;
    bpf_probe_read_kernel(&file, sizeof(file), &fds[nr]);
    return file;
}

// 从 file 获取 sock 结构体
static __noinline struct sock *sock_from_file(struct file *file)
{
    if (!file) return NULL;
    struct inode *inode = BPF_CORE_READ(file, f_inode);
    if (!inode) return NULL;
    unsigned short mode = BPF_CORE_READ(inode, i_mode);
    if ((mode & 0170000) != 0140000)
        return NULL;
    struct socket *sock = (struct socket *)BPF_CORE_READ(file, private_data);
    if (!sock) return NULL;
    return BPF_CORE_READ(sock, sk);
}

// 组合 fd -> sock
static __noinline struct sock *sockfd_lookup(struct task_struct *task, int fd)
{
    struct file *file = fget_raw(task, fd);
    if (!file) return NULL;
    return sock_from_file(file);
}

// CO-RE 位域读取 sk_protocol
static __noinline int sock_prot(struct sock *sk)
{
    unsigned long long prot = 0;
    unsigned int offset = __builtin_preserve_field_info(
        sk->sk_protocol, BPF_FIELD_BYTE_OFFSET);
    unsigned int size = __builtin_preserve_field_info(
        sk->sk_protocol, BPF_FIELD_BYTE_SIZE);
    bpf_probe_read_kernel(&prot, size & 0x0f, (void *)sk + offset);
    prot <<= __builtin_preserve_field_info(
        sk->sk_protocol, BPF_FIELD_LSHIFT_U64);
    prot >>= __builtin_preserve_field_info(
        sk->sk_protocol, BPF_FIELD_RSHIFT_U64);
    return (int)prot;
}

// 获取 socket address family
static __always_inline unsigned short sock_family(struct sock *sk)
{
    return BPF_CORE_READ(sk, __sk_common.skc_family);
}

// 从 sock 提取 IPv4 地址信息
// 包含 fallback 逻辑：skc_rcv_saddr 为 0 时尝试 inet_saddr
static __always_inline void query_ipv4(struct sock *sk,
    __u32 *src_ip, __u16 *src_port,
    __u32 *dst_ip, __u16 *dst_port)
{
    *src_ip = BPF_CORE_READ(sk, __sk_common.skc_rcv_saddr);
    *src_port = BPF_CORE_READ(sk, __sk_common.skc_num);
    *dst_ip = BPF_CORE_READ(sk, __sk_common.skc_daddr);
    *dst_port = BPF_CORE_READ(sk, __sk_common.skc_dport);

    if (*src_ip == 0) {
        struct inet_sock *inet = (struct inet_sock *)sk;
        *src_ip = BPF_CORE_READ(inet, inet_saddr);
    }
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

// DNS 域名状态机解析
// 逐字节处理 DNS 查询域名，将 3www6google3com0 转换为 .www.google.com
// 返回 1=继续, 0=结束
static __noinline int process_domain_name(char *data, char *name, int *flag, int i)
{
    char rc = *(data + 12 + i);
    int v = *flag;
    if (0 == rc) return 0;
    if (v == 0) {
        v = rc;
        name[i - 1] = 46;
    } else {
        name[i - 1] = rc;
        v = v - 1;
    }
    *flag = v;
    return 1;
}

// DNS 解析主循环
// 从 DNS 原始数据中提取域名和查询类型
// 返回: 0=成功, -1=失败
static __noinline int query_dns_record(char *data, int data_len, char *domain, int domain_size, __u16 *query_type)
{
    if (data_len < 12)
        return -1;

    int flag = 0;
    int i;

    for (i = 1; i < 64; i++) {
        if (12 + i >= data_len || i >= domain_size)
            break;
        if (!process_domain_name(data, domain, &flag, i))
            break;
    }

    if (i > 0 && i < domain_size)
        domain[i - 1] = 0;

    int qtype_offset = 12 + i + 1;
    if (qtype_offset + 2 <= data_len && qtype_offset + 2 <= DNS_RECORD_MAX) {
        *query_type = ((__u16)(__u8)data[qtype_offset] << 8) | (__u8)data[qtype_offset + 1];
    }

    return 0;
}

// ========== sys_exit: 处理系统调用返回 ==========

SEC("raw_tracepoint/sys_exit")
int tp_sys_exit(struct bpf_raw_tracepoint_args *ctx)
{
    struct pt_regs *regs = (struct pt_regs *)ctx->args[0];
    long retval = ctx->args[1];

    unsigned long syscall_nr = 0;
    bpf_probe_read_kernel(&syscall_nr, sizeof(syscall_nr), &regs->orig_ax);

    if (syscall_nr != 42 && syscall_nr != 49 && syscall_nr != 43 &&
        syscall_nr != 288 && syscall_nr != 45 && syscall_nr != 47)
        return 0;

    struct task_struct *task = (struct task_struct *)bpf_get_current_task();

    __u64 parm1 = 0, parm2 = 0;
    bpf_probe_read_kernel(&parm1, sizeof(parm1), &regs->di);
    bpf_probe_read_kernel(&parm2, sizeof(parm2), &regs->si);

    if (syscall_nr == 42) {
        if (retval != 0)
            return 0;

        int fd = (int)parm1;

        struct sock *sk = sockfd_lookup(task, fd);
        if (!sk)
            return 0;

        if (sock_family(sk) != 2)
            return 0;

        u32 key = 0;
        struct connect_event *evt = bpf_map_lookup_elem(&percpu_connect_buf, &key);
        if (!evt)
            return 0;

        __builtin_memset(evt, 0, sizeof(*evt));
        evt->event_type = EVENT_TYPE_CONNECT;

        FILL_PROCESS_INFO(task, evt);

        __u32 src_ip = 0, dst_ip = 0;
        __u16 src_port = 0, dst_port = 0;
        query_ipv4(sk, &src_ip, &src_port, &dst_ip, &dst_port);

        evt->remote_ip = dst_ip;
        evt->remote_port = dst_port;
        evt->local_ip = src_ip;
        evt->local_port = src_port;
        evt->protocol = (__u8)sock_prot(sk);
        evt->retval = (__s32)retval;

        if (evt->remote_ip == 0)
            return 0;

        read_full_exe_path(task, evt->exe_path, sizeof(evt->exe_path));

        bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, evt, sizeof(*evt));
        return 0;
    }

    if (syscall_nr == 49) {
        if (retval != 0)
            return 0;

        int fd = (int)parm1;

        struct sock *sk = sockfd_lookup(task, fd);
        if (!sk)
            return 0;

        if (sock_family(sk) != 2)
            return 0;

        u32 key = 0;
        struct bind_event *evt = bpf_map_lookup_elem(&percpu_bind_buf, &key);
        if (!evt)
            return 0;

        __builtin_memset(evt, 0, sizeof(*evt));
        evt->event_type = EVENT_TYPE_BIND;

        FILL_PROCESS_INFO(task, evt);

        __u32 src_ip = 0, dst_ip = 0;
        __u16 src_port = 0, dst_port = 0;
        query_ipv4(sk, &src_ip, &src_port, &dst_ip, &dst_port);

        evt->bind_ip = src_ip;
        evt->bind_port = bpf_htons(src_port);
        evt->protocol = (__u8)sock_prot(sk);
        evt->retval = (__s32)retval;

        read_full_exe_path(task, evt->exe_path, sizeof(evt->exe_path));

        bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, evt, sizeof(*evt));
        return 0;
    }

    if (syscall_nr == 43 || syscall_nr == 288) {
        if (retval < 0)
            return 0;

        int new_fd = (int)retval;

        struct sock *sk = sockfd_lookup(task, new_fd);
        if (!sk)
            return 0;

        if (sock_family(sk) != 2)  // AF_INET
            return 0;

        u32 key = 0;
        struct accept_event *evt = bpf_map_lookup_elem(&percpu_accept_buf, &key);
        if (!evt)
            return 0;

        __builtin_memset(evt, 0, sizeof(*evt));
        evt->event_type = EVENT_TYPE_ACCEPT;

        FILL_PROCESS_INFO(task, evt);

        __u32 src_ip = 0, dst_ip = 0;
        __u16 src_port = 0, dst_port = 0;
        query_ipv4(sk, &src_ip, &src_port, &dst_ip, &dst_port);

        evt->remote_ip = dst_ip;
        evt->remote_port = dst_port;
        evt->local_ip = src_ip;
        evt->local_port = src_port;
        evt->protocol = (__u8)sock_prot(sk);
        evt->retval = (__s32)retval;

        if (evt->remote_ip == 0)
            return 0;

        read_full_exe_path(task, evt->exe_path, sizeof(evt->exe_path));

        bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, evt, sizeof(*evt));
        return 0;
    }

    if (syscall_nr == 45 || syscall_nr == 47) {
        if (retval <= 0)
            return 0;

        int fd = (int)parm1;

        struct sock *sk = sockfd_lookup(task, fd);
        if (!sk)
            return 0;

        if (sock_family(sk) != 2)
            return 0;

        if (sock_prot(sk) != 17)
            return 0;

        __u16 peer_port = BPF_CORE_READ(sk, __sk_common.skc_dport);
        if (peer_port != __bpf_htons(53) && peer_port != __bpf_htons(5353))
            return 0;

        __u64 user_buf = parm2;

        if (syscall_nr == 47) {
            struct iovec iov;
            __u64 msg_iov_ptr = 0;

            bpf_probe_read_user(&msg_iov_ptr, sizeof(msg_iov_ptr),
                (void *)(user_buf + 16));
            if (!msg_iov_ptr)
                return 0;

            bpf_probe_read_user(&iov, sizeof(iov), (void *)msg_iov_ptr);
            if (!iov.iov_base)
                return 0;

            user_buf = (__u64)iov.iov_base;
        }

        u32 dns_data_key = 0;
        struct dns_data_buf *dns_buf = bpf_map_lookup_elem(&percpu_dns_data, &dns_data_key);
        if (!dns_buf)
            return 0;

        int read_len = (int)retval;
        if (read_len > DNS_RECORD_MAX)
            read_len = DNS_RECORD_MAX;
        if (read_len < 12)
            return 0;

        bpf_probe_read_user(dns_buf->data, read_len & DNS_RECORD_MASK, (void *)user_buf);

        if (!(dns_buf->data[2] & 0x80))
            return 0;

        u32 dns_key = 0;
        struct dns_event *evt = bpf_map_lookup_elem(&percpu_dns_buf, &dns_key);
        if (!evt)
            return 0;

        __builtin_memset(evt, 0, sizeof(*evt));
        evt->event_type = EVENT_TYPE_DNS;

        __u8 flags1 = dns_buf->data[2];
        __u8 flags2 = dns_buf->data[3];
        evt->opcode = (flags1 >> 3) & 0x0F;
        evt->rcode = flags2 & 0x0F;

        __u16 qtype = 0;
        if (query_dns_record(dns_buf->data, read_len, evt->domain, sizeof(evt->domain), &qtype) < 0)
            return 0;
        evt->query_type = qtype;

        FILL_PROCESS_INFO(task, evt);

        evt->dns_server_ip = BPF_CORE_READ(sk, __sk_common.skc_daddr);
        evt->dns_server_port = peer_port;

        read_full_exe_path(task, evt->exe_path, sizeof(evt->exe_path));

        bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, evt, sizeof(*evt));
        return 0;
    }

    return 0;
}

char LICENSE[] SEC("license") = "GPL";
