package agent

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGenerateIDFromDMIAndMAC 测试基于 DMI 和 MAC 地址生成 Agent ID
func TestGenerateIDFromDMIAndMAC(t *testing.T) {
	id, ok := GenerateIDFromDMIAndMAC()

	if !ok {
		t.Log("Failed to generate Agent ID from DMI and MAC")
		t.Log("This is normal if running in a VM without valid hardware info")
		return
	}

	if id == "" {
		t.Error("Generated ID should not be empty when ok is true")
		return
	}

	if len(id) != 36 {
		t.Errorf("Generated ID length is invalid: got %d, expected 36", len(id))
	}

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
	_, err := fromIDFile("/nonexistent/file")
	if err == nil {
		t.Error("Expected error when reading nonexistent file")
	}

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

// TestGenerateIDFromMachineID 测试基于 machine-id 生成 Agent ID（回退方案）
func TestGenerateIDFromMachineID(t *testing.T) {
	id := GenerateIDFromMachineID("")
	if id == "" {
		t.Error("Generated ID should not be empty")
	}

	if len(id) != 36 {
		t.Errorf("Generated ID length is invalid: got %d, expected 36", len(id))
	}

	t.Logf("Generated ID from machine-id: %s", id)
}

// TestGenerateIDFromMachineID_WithWorkingDir 测试使用本地持久化的 machine-id
func TestGenerateIDFromMachineID_WithWorkingDir(t *testing.T) {
	tmpDir := t.TempDir()

	id := GenerateIDFromMachineID(tmpDir)
	if id == "" {
		t.Error("Generated ID should not be empty")
	}

	if len(id) != 36 {
		t.Errorf("Generated ID length is invalid: got %d, expected 36", len(id))
	}

	t.Logf("Generated ID with working dir: %s", id)
}

// TestGenerateIDFromMachineID_Consistency 测试 ID 生成的一致性
func TestGenerateIDFromMachineID_Consistency(t *testing.T) {
	id1 := GenerateIDFromMachineID("")
	id2 := GenerateIDFromMachineID("")

	if id1 == "" || id2 == "" {
		t.Error("Generated IDs should not be empty")
	}

	t.Logf("First ID: %s", id1)
	t.Logf("Second ID: %s", id2)

	if _, err := os.ReadFile("/etc/machine-id"); err == nil {
		if id1 != id2 {
			t.Errorf("ID generation should be consistent when machine-id exists: first=%s, second=%s", id1, id2)
		}
	}
}

// TestFromUUIDFile 测试 UUID 文件读取函数
func TestFromUUIDFile(t *testing.T) {
	_, err := fromUUIDFile("/nonexistent/file")
	if err == nil {
		t.Error("Expected error when reading nonexistent file")
	}

	mid, err := fromUUIDFile("/etc/machine-id")
	if err == nil {
		if mid.String() == "" {
			t.Error("Expected non-empty UUID from /etc/machine-id")
		}
		t.Logf("Successfully read UUID from /etc/machine-id: %s", mid.String())
	} else {
		t.Logf("/etc/machine-id is not a valid UUID format (this is normal): %v", err)
	}
}

// TestPersistID 测试 Agent ID 持久化功能
func TestPersistID(t *testing.T) {
	tmpDir := t.TempDir()
	testID := "12345678-1234-1234-1234-123456789abc"

	err := PersistID(tmpDir, testID)
	if err != nil {
		t.Fatalf("Failed to persist ID: %v", err)
	}

	idFile := filepath.Join(tmpDir, "machine-id")
	if _, err := os.Stat(idFile); os.IsNotExist(err) {
		t.Error("ID file was not created")
	}

	content, err := os.ReadFile(idFile)
	if err != nil {
		t.Fatalf("Failed to read ID file: %v", err)
	}
	if string(content) != testID {
		t.Errorf("ID file content mismatch: got %s, expected %s", string(content), testID)
	}

	info, err := os.Stat(idFile)
	if err != nil {
		t.Fatalf("Failed to stat ID file: %v", err)
	}
	mode := info.Mode().Perm()
	expectedMode := os.FileMode(0600)
	if mode != expectedMode {
		t.Errorf("ID file permissions mismatch: got %o, expected %o", mode, expectedMode)
	}

	t.Logf("Successfully persisted ID to %s with permissions %o", idFile, mode)
}

// TestPersistID_ErrorCases 测试错误情况
func TestPersistID_ErrorCases(t *testing.T) {
	tmpDir := t.TempDir()
	testID := "12345678-1234-1234-1234-123456789abc"

	err := PersistID("", testID)
	if err == nil {
		t.Error("Expected error when working directory is empty")
	}

	err = PersistID(tmpDir, "")
	if err == nil {
		t.Error("Expected error when ID is empty")
	}
}
