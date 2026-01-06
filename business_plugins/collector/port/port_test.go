package port

import (
	"fmt"
	"os"
	"testing"

	"golang.org/x/sys/unix"
)

// TestListeningPorts 测试端口信息获取功能
// 这是一个简单的测试程序，用于验证 procfs 方法是否能正确读取端口信息
func TestListeningPorts(t *testing.T) {
	ports, err := ListeningPorts()
	if err != nil {
		t.Fatalf("Failed to get listening ports: %v", err)
	}

	if len(ports) == 0 {
		t.Log("No listening ports found (this might be normal)")
		return
	}

	fmt.Println("\n========== Listening Ports ==========")
	fmt.Printf("Total ports found: %d\n\n", len(ports))

	// 打印前 10 个端口信息（避免输出过多）
	maxPrint := 10
	if len(ports) < maxPrint {
		maxPrint = len(ports)
	}

	for i, p := range ports[:maxPrint] {
		fmt.Printf("[%d] %s %s:%s -> %s:%s\n",
			i+1,
			getProtocolName(p.Protocol),
			p.Sip, p.Sport,
			p.Dip, p.Dport)
		fmt.Printf("     State: %s, UID: %s (%s), Inode: %s\n",
			p.State, p.Uid, p.Username, p.Inode)
		fmt.Println()
	}

	if len(ports) > maxPrint {
		fmt.Printf("... and %d more ports\n", len(ports)-maxPrint)
	}
	fmt.Println("=====================================")
	fmt.Println()
}

// TestProcNet 测试 procNet 函数
func TestProcNet(t *testing.T) {
	// 测试 TCP IPv4
	tcp4Ports, err := procNet(unix.AF_INET, unix.IPPROTO_TCP)
	if err != nil {
		t.Logf("TCP IPv4: %v (might not be available)", err)
	} else {
		t.Logf("TCP IPv4: found %d listening ports", len(tcp4Ports))
	}

	// 测试 TCP IPv6
	tcp6Ports, err := procNet(unix.AF_INET6, unix.IPPROTO_TCP)
	if err != nil {
		t.Logf("TCP IPv6: %v (might not be available)", err)
	} else {
		t.Logf("TCP IPv6: found %d listening ports", len(tcp6Ports))
	}

	// 测试 UDP IPv4
	udp4Ports, err := procNet(unix.AF_INET, unix.IPPROTO_UDP)
	if err != nil {
		t.Logf("UDP IPv4: %v (might not be available)", err)
	} else {
		t.Logf("UDP IPv4: found %d listening ports", len(udp4Ports))
	}

	// 测试 UDP IPv6
	udp6Ports, err := procNet(unix.AF_INET6, unix.IPPROTO_UDP)
	if err != nil {
		t.Logf("UDP IPv6: %v (might not be available)", err)
	} else {
		t.Logf("UDP IPv6: found %d listening ports", len(udp6Ports))
	}
}

// TestParseIP 测试 IP 地址解析
func TestParseIP(t *testing.T) {
	testCases := []struct {
		hex      string
		expected string
	}{
		{"0100007F", "127.0.0.1"},           // IPv4 localhost
		{"00000000", "0.0.0.0"},             // IPv4 any
		{"FFFFFFFF", "255.255.255.255"},     // IPv4 broadcast
		{"00000000000000000000000000000001", "::1"}, // IPv6 localhost
	}

	for _, tc := range testCases {
		result, err := parseIP(tc.hex)
		if err != nil {
			t.Errorf("parseIP(%s) failed: %v", tc.hex, err)
			continue
		}
		if result != tc.expected {
			t.Errorf("parseIP(%s) = %s, expected %s", tc.hex, result, tc.expected)
		} else {
			t.Logf("parseIP(%s) = %s ✓", tc.hex, result)
		}
	}
}

// getProtocolName 获取协议名称（用于显示）
func getProtocolName(proto string) string {
	switch proto {
	case "6":
		return "TCP"
	case "17":
		return "UDP"
	default:
		return "UNKNOWN"
	}
}

// TestMain 主测试函数（可选）
func TestMain(m *testing.M) {
	// 检查是否有 root 权限（读取 /proc/net/ 通常不需要 root，但某些系统可能需要）
	fmt.Println("Testing port information collection...")
	fmt.Println("Note: This test reads from /proc/net/ files")
	os.Exit(m.Run())
}

