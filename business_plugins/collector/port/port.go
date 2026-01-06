package port

import (
	"strconv"

	"golang.org/x/sys/unix"
)

var (
	scanProto  = [2]int{unix.IPPROTO_UDP, unix.IPPROTO_TCP}
	scanFamily = [2]int{unix.AF_INET, unix.AF_INET6}
)

// Port 端口信息结构体
// 第一次移植：只包含从 /proc/net/ 文件获取的基本信息
type Port struct {
	// 网络层信息（从 inet socket 获取）
	Family   string `mapstructure:"family"`   // IP 地址族：2=IPv4, 10=IPv6
	Protocol string `mapstructure:"protocol"`  // 协议：6=TCP, 17=UDP
	State    string `mapstructure:"state"`     // 端口状态：TCP 10=LISTEN, UDP 7=UDP
	Sport    string `mapstructure:"sport"`     // 源端口（本地端口）
	Dport    string `mapstructure:"dport"`    // 目标端口（远程端口，通常为0）
	Sip      string `mapstructure:"sip"`      // 源 IP 地址（本地 IP）
	Dip      string `mapstructure:"dip"`      // 目标 IP 地址（远程 IP，通常为0.0.0.0）
	Uid      string `mapstructure:"uid"`      // 用户 ID
	Inode    string `mapstructure:"inode"`    // Socket inode 号（用于关联进程）
	Username string `mapstructure:"username"` // 用户名（从 UID 获取）

	// 进程信息（第二次移植时添加）
	// Pid     string `mapstructure:"pid"`
	// Exe     string `mapstructure:"exe"`
	// Comm    string `mapstructure:"comm"`
	// Cmdline string `mapstructure:"cmdline"`
	// Psm     string `mapstructure:"psm"`
	// PodName string `mapstructure:"pod_name"`
}

// ListeningPorts 获取所有监听端口信息
// 第一次移植：只使用 procfs 方法（读取 /proc/net/ 文件）
func ListeningPorts() (ret []*Port, err error) {
	// 遍历所有协议（TCP 和 UDP）
	for _, proto := range scanProto {
		sp := strconv.Itoa(int(proto))
		// 遍历所有地址族（IPv4 和 IPv6）
		for _, family := range scanFamily {
			// 使用 procfs 方法读取端口信息
			var ps []*Port
			ps, err = procNet(uint8(family), uint8(proto))
			if err != nil {
				// 如果读取失败，继续尝试下一个
				continue
			}
			// 将获取到的端口信息添加到结果中
			for _, p := range ps {
				// 设置协议和地址族信息
				p.Protocol = sp
				p.Family = strconv.Itoa(int(family))
				ret = append(ret, p)
			}
		}
	}
	return
}

