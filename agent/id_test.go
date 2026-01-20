package agent

import (
	"testing"
)

// TestGenerateIDFromDMIAndMAC 测试基于 DMI 和 MAC 地址生成 Agent ID
func TestGenerateIDFromDMIAndMAC(t *testing.T) {
	id, ok := GenerateIDFromDMIAndMAC()

	if !ok {
		t.Log("Failed to generate Agent ID from DMI and MAC")
		t.Log("This is normal if running in a VM without valid hardware info")
		// 不直接失败，因为某些环境可能没有有效的硬件信息
		return
	}

	if id == "" {
		t.Error("Generated ID should not be empty when ok is true")
		return
	}

	// 验证生成的 ID 是有效的 UUID 格式
	if len(id) != 36 { // UUID 格式: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx (36 字符)
		t.Errorf("Generated ID length is invalid: got %d, expected 36", len(id))
	}

	// 验证 UUID 格式（简单检查：包含 4 个连字符）
	hyphenCount := 0
	for _, c := range id {
		if c == '-' {
			hyphenCount++
		}
	}
	if hyphenCount != 4 {
		t.Errorf("Generated ID format is invalid: expected 4 hyphens, got %d", hyphenCount)
	}

	t.Logf("Successfully generated Agent ID: %s", id)
}

// TestGenerateIDFromDMIAndMAC_Consistency 测试 ID 生成的一致性
// 在相同硬件环境下，应该生成相同的 ID
func TestGenerateIDFromDMIAndMAC_Consistency(t *testing.T) {
	id1, ok1 := GenerateIDFromDMIAndMAC()
	id2, ok2 := GenerateIDFromDMIAndMAC()

	if !ok1 || !ok2 {
		t.Skip("Skipping consistency test: hardware info not available")
		return
	}

	if id1 != id2 {
		t.Errorf("ID generation is not consistent: first=%s, second=%s", id1, id2)
	}
}

// TestFromIDFile 测试文件读取辅助函数
func TestFromIDFile(t *testing.T) {
	// 测试读取不存在的文件
	_, err := fromIDFile("/nonexistent/file")
	if err == nil {
		t.Error("Expected error when reading nonexistent file")
	}

	// 测试读取存在的文件（如果 /etc/hostname 存在）
	hostname, err := fromIDFile("/etc/hostname")
	if err == nil {
		if len(hostname) == 0 {
			t.Error("Expected non-empty content from /etc/hostname")
		}
		t.Logf("Successfully read /etc/hostname: %s", string(hostname))
	}
}

// TestIsInvalidProductUUID 测试无效 UUID 检查
func TestIsInvalidProductUUID(t *testing.T) {
	testCases := []struct {
		uuid     string
		expected bool
	}{
		{"03000200-0400-0500-0006-000700080009", true},  // 无效
		{"02000100-0300-0400-0005-000600070008", true},  // 无效
		{"12345678-1234-1234-1234-123456789abc", false}, // 有效
		{"", false}, // 空字符串
	}

	for _, tc := range testCases {
		result := isInvalidProductUUID(tc.uuid)
		if result != tc.expected {
			t.Errorf("isInvalidProductUUID(%q) = %v, expected %v", tc.uuid, result, tc.expected)
		}
	}
}

// TestIsInvalidProductName 测试无效产品名称检查
func TestIsInvalidProductName(t *testing.T) {
	testCases := []struct {
		name     []byte
		expected bool
	}{
		{[]byte("--"), true},
		{[]byte("unknown"), true},
		{[]byte("To be filled by O.E.M."), true},
		{[]byte("OEM not specify"), true},
		{[]byte("t.b.d"), true},
		{[]byte("T.B.D"), true},
		{[]byte(""), true},                         // 空字符串
		{[]byte("Dell Inc."), false},               // 有效
		{[]byte("VMware Virtual Platform"), false}, // 有效（即使是虚拟机，名称也是有效的）
	}

	for _, tc := range testCases {
		result := isInvalidProductName(tc.name)
		if result != tc.expected {
			t.Errorf("isInvalidProductName(%q) = %v, expected %v", string(tc.name), result, tc.expected)
		}
	}
}
