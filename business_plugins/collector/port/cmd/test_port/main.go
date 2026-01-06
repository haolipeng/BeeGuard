package main

import (
	"fmt"
	"os"

	"gitlab.myinterest.top/security/agent/business_plugins/collector/port"
)

func main() {
	fmt.Println("=== Port Information Collection Test ===")
	fmt.Println("Reading port information from /proc/net/ files...")
	fmt.Println()

	// 获取所有监听端口
	ports, err := port.ListeningPorts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(ports) == 0 {
		fmt.Println("No listening ports found.")
		return
	}

	fmt.Printf("Found %d listening ports:\n\n", len(ports))

	// 打印端口信息（限制显示前 20 个，避免输出过多）
	maxPrint := 20
	if len(ports) < maxPrint {
		maxPrint = len(ports)
	}

	for i, p := range ports[:maxPrint] {
		protocol := getProtocolName(p.Protocol)
		family := getFamilyName(p.Family)

		fmt.Printf("[%d] %s %s\n", i+1, protocol, family)
		fmt.Printf("     Local:  %s:%s\n", p.Sip, p.Sport)
		fmt.Printf("     Remote: %s:%s\n", p.Dip, p.Dport)
		fmt.Printf("     State: %s", p.State)
		if p.State == "10" {
			fmt.Print(" (LISTEN)")
		} else if p.State == "7" {
			fmt.Print(" (UDP)")
		}
		fmt.Println()
		fmt.Printf("     UID: %s (%s)\n", p.Uid, p.Username)
		fmt.Printf("     Inode: %s\n", p.Inode)
		fmt.Println()
	}

	if len(ports) > maxPrint {
		fmt.Printf("... and %d more ports\n", len(ports)-maxPrint)
	}

	fmt.Println("Test completed successfully!")
	fmt.Printf("\nCompare with system command:\n")
	fmt.Println("  ss -tuln | head -20")
	fmt.Println("  netstat -tuln | head -20")
}

func getProtocolName(proto string) string {
	switch proto {
	case "6":
		return "TCP"
	case "17":
		return "UDP"
	default:
		return "UNKNOWN(" + proto + ")"
	}
}

func getFamilyName(family string) string {
	switch family {
	case "2":
		return "IPv4"
	case "10":
		return "IPv6"
	default:
		return "UNKNOWN(" + family + ")"
	}
}

