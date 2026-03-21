package port

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gitlab.myinterest.top/security/agent/business_plugins/collector/utils"
	"golang.org/x/sys/unix"
)

// parseIP 解析十六进制格式的 IP 地址
// /proc/net/ 文件中的 IP 地址是十六进制格式，需要转换为点分十进制格式
// 例如：0100007F -> 127.0.0.1
func parseIP(h string) (ret string, err error) {
	var byteIP []byte
	byteIP, err = hex.DecodeString(h)
	if err != nil {
		return
	}
	switch len(byteIP) {
	case 4:
		// IPv4 地址：4 字节，需要反转字节顺序
		// 例如：[01 00 00 7F] -> [7F 00 00 01] -> 127.0.0.1
		ret = net.IP{byteIP[3], byteIP[2], byteIP[1], byteIP[0]}.String()
		return
	case 16:
		// IPv6 地址：16 字节，需要反转每4字节的顺序
		ret = net.IP{
			byteIP[3], byteIP[2], byteIP[1], byteIP[0],
			byteIP[7], byteIP[6], byteIP[5], byteIP[4],
			byteIP[11], byteIP[10], byteIP[9], byteIP[8],
			byteIP[15], byteIP[14], byteIP[13], byteIP[12],
		}.String()
		return
	default:
		err = fmt.Errorf("unable to parse IP %s", h)
		return
	}
}

// procNet 从 /proc/net/ 文件读取端口信息
// family: unix.AF_INET (IPv4) 或 unix.AF_INET6 (IPv6)
// proto: unix.IPPROTO_TCP (TCP) 或 unix.IPPROTO_UDP (UDP)
func procNet(family, proto uint8) (ret []*Port, err error) {
	var f *os.File
	var f1, f2 string

	// 根据协议确定文件名
	switch proto {
	case unix.IPPROTO_UDP:
		f1 = "udp"
	case unix.IPPROTO_TCP:
		f1 = "tcp"
	default:
		err = fmt.Errorf("unsupported protocol %d", proto)
		return
	}

	// 根据地址族确定文件名后缀
	// 注意：Linux 系统中，IPv4 的文件名就是 tcp/udp（没有后缀），IPv6 的文件名是 tcp6/udp6
	switch family {
	case unix.AF_INET:
		f2 = "" // IPv4: /proc/net/tcp 或 /proc/net/udp（无后缀）
	case unix.AF_INET6:
		f2 = "6" // IPv6: /proc/net/tcp6 或 /proc/net/udp6
	default:
		err = fmt.Errorf("unsupported family %d", family)
		return
	}

	// 打开对应的 /proc/net/ 文件
	// 例如：/proc/net/tcp, /proc/net/tcp6, /proc/net/udp, /proc/net/udp6
	f, err = os.Open(filepath.Join("/proc/net", f1+f2))
	if err != nil {
		return
	}
	defer f.Close()

	// 限制读取大小（最大 2MB，防止文件过大）
	r := bufio.NewScanner(io.LimitReader(f, 1024*1024*2))

	// 解析文件头，确定字段位置
	hdr := map[int]string{}
	for i := 0; r.Scan(); i++ {
		if i == 0 {
			// 第一行是表头，解析字段位置
			line := r.Text()
			hdr[1] = "local_address" // 本地地址（IP:端口）
			hdr[2] = "rem_address"   // 远程地址（IP:端口）
			hdr[3] = "st"            // 状态（state）
			hdr[7] = "uid"           // 用户 ID

			// 查找 "inode" 字段的位置（在 "uid" 之后）
			uidIndex := strings.Index(line, "uid")
			if uidIndex >= 0 {
				fieldsAfterUid := strings.Fields(line[uidIndex+3:])
				for index, field := range fieldsAfterUid {
					if field == "inode" {
						hdr[8+index] = "inode"
						break
					}
				}
			}
		} else {
			// 解析数据行
			fields := strings.Fields(r.Text())
			p := &Port{}
			var parseErr error

			// 遍历字段，根据表头映射填充 Port 结构
			for i, f := range fields {
				if k, ok := hdr[i]; ok {
					switch k {
					case "local_address":
						// 解析本地地址：格式为 "IP:端口"（十六进制）
						// 例如："0100007F:0016" -> IP: 127.0.0.1, Port: 22
						addrFields := strings.Split(f, ":")
						if len(addrFields) != 2 {
							break
						}
						// 解析 IP 地址
						p.Sip, parseErr = parseIP(addrFields[0])
						if parseErr != nil {
							break
						}
						// 解析端口号（十六进制转十进制）
						var uport uint64
						uport, parseErr = strconv.ParseUint(addrFields[1], 16, 64)
						if parseErr != nil {
							break
						}
						p.Sport = strconv.FormatUint(uport, 10)

					case "rem_address":
						// 解析远程地址：格式为 "IP:端口"（十六进制）
						addrFields := strings.Split(f, ":")
						if len(addrFields) != 2 {
							break
						}
						// 解析 IP 地址
						p.Dip, parseErr = parseIP(addrFields[0])
						if parseErr != nil {
							break
						}
						// 解析端口号（十六进制转十进制）
						var uport uint64
						uport, parseErr = strconv.ParseUint(addrFields[1], 16, 64)
						if parseErr != nil {
							break
						}
						p.Dport = strconv.FormatUint(uport, 10)

					case "st":
						// 解析状态（十六进制转十进制）
						var st uint64
						st, parseErr = strconv.ParseUint(f, 16, 64)
						if parseErr != nil {
							break
						}
						p.State = strconv.FormatUint(st, 10)

					case "uid":
						// 用户 ID
						p.Uid = f
						// 根据 UID 获取用户名
						p.Username, _ = utils.GetUsername(f)

					case "inode":
						// Socket inode 号（用于后续关联进程）
						p.Inode = f
					}
				}
			}

			// 只收集监听状态的端口
			// TCP: 状态 10 = LISTEN（监听状态）
			// UDP: 状态 7 = UDP（UDP 协议本身没有状态，7 表示 UDP socket）
			if parseErr == nil {
				isListening := false
				if proto == unix.IPPROTO_UDP && p.State == "7" {
					isListening = true
				} else if proto == unix.IPPROTO_TCP && p.State == "10" {
					isListening = true
				}

				if isListening {
					ret = append(ret, p)
				}
			}
		}
	}

	return
}
