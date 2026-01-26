package host

import (
	"net"
	"os"
	"strings"
	"sync/atomic"
)

var (
	Name    atomic.Value
	IPv4    atomic.Value
	MacAddr atomic.Value // MAC地址
)

// RefreshHost 刷新主机信息（主机名、IPv4 地址、MAC地址）
func RefreshHost() {
	hostname, _ := os.Hostname()
	Name.Store(hostname)

	ipv4List := []string{}
	macAddr := ""
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, i := range interfaces {
			// 过滤虚拟接口（docker、loopback、bridge等）
			name := i.Name
			if strings.HasPrefix(name, "docker") || strings.HasPrefix(name, "lo") ||
				strings.HasPrefix(name, "veth") || strings.HasPrefix(name, "br-") ||
				strings.HasPrefix(name, "virbr") {
				continue
			}

			// 获取MAC地址（取第一个有效的物理网卡MAC）
			if macAddr == "" && len(i.HardwareAddr) > 0 {
				mac := i.HardwareAddr.String()
				// 过滤无效MAC地址
				if mac != "" && mac != "00:00:00:00:00:00" {
					macAddr = mac
				}
			}

			addrs, err := i.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				ip, _, err := net.ParseCIDR(addr.String())
				if err != nil || !ip.IsGlobalUnicast() {
					continue
				}
				// 只收集IPv4地址
				if ip4 := ip.To4(); ip4 != nil {
					ipv4List = append(ipv4List, ip4.String())
				}
			}
		}
	}
	// 限制收集数量避免过多
	if len(ipv4List) > 10 {
		ipv4List = ipv4List[:10]
	}
	// 存储IPv4地址列表
	IPv4.Store(ipv4List)
	// 存储MAC地址
	MacAddr.Store(macAddr)
}

func init() {
	RefreshHost()
}
