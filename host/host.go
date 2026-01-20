package host

import (
	"net"
	"os"
	"strings"
	"sync/atomic"
)

var (
	Name atomic.Value
	IPv4 atomic.Value
)

// RefreshHost 刷新主机信息（主机名和 IPv4 地址）
func RefreshHost() {
	hostname, _ := os.Hostname()
	Name.Store(hostname)

	ipv4List := []string{}
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, i := range interfaces {
			// 过滤虚拟接口（docker、loopback、bridge等）
			name := i.Name
			if strings.HasPrefix(name, "docker") || strings.HasPrefix(name, "lo") {
				continue
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
}

func init() {
	RefreshHost()
}
