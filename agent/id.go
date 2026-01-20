package agent

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// fromIDFile 从文件中读取 ID 信息
func fromIDFile(file string) (id []byte, err error) {
	id, err = os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	if len(id) < 6 {
		return nil, errors.New("id too short")
	}
	id = bytes.TrimSpace(id)
	return id, nil
}

// invalidProductUUIDs 无效的产品 UUID 列表（虚拟机的默认值）
var invalidProductUUIDs = []string{
	"03000200-0400-0500-0006-000700080009",
	"02000100-0300-0400-0005-000600070008",
}

// invalidProductNames 无效的产品名称列表（虚拟机的默认值或占位符）
var invalidProductNames = [][]byte{
	[]byte("--"),
	[]byte("unknown"),
	[]byte("To be filled by O.E.M."),
	[]byte("OEM not specify"),
	[]byte("t.b.d"),
	[]byte("T.B.D"),
}

// isInvalidProductUUID 检查产品 UUID 是否为无效值
func isInvalidProductUUID(uuidStr string) bool {
	for _, invalid := range invalidProductUUIDs {
		if uuidStr == invalid {
			return true
		}
	}
	return false
}

// isInvalidProductName 检查产品名称是否为无效值
func isInvalidProductName(name []byte) bool {
	if len(name) == 0 {
		return true
	}
	nameLower := bytes.ToLower(name)
	for _, invalid := range invalidProductNames {
		if bytes.Equal(nameLower, bytes.ToLower(invalid)) {
			return true
		}
	}
	return false
}

// fromUUIDFile 从文件中读取 UUID 格式的 ID
func fromUUIDFile(file string) (id uuid.UUID, err error) {
	var idBytes []byte
	idBytes, err = os.ReadFile(file)
	if err != nil {
		return
	}
	id, err = uuid.ParseBytes(bytes.TrimSpace(idBytes))
	return
}

// GenerateIDFromDMIAndMAC 基于 DMI 信息和 MAC 地址生成 Agent ID
func GenerateIDFromDMIAndMAC() (string, bool) {
	source := []byte{}

	// 1. 读取产品 UUID（DMI）
	pdid, err := fromIDFile("/sys/class/dmi/id/product_uuid")
	if err == nil {
		pdidStr := string(pdid)
		if !isInvalidProductUUID(pdidStr) {
			source = append(source, pdid...)
		}
	}

	// 2. 读取 MAC 地址（仅尝试 eth0接口）
	emac, err := fromIDFile("/sys/class/net/eth0/address")
	if err == nil {
		source = append(source, emac...)
	}

	// 3. 检查是否有足够的信息源（至少需要 8 字节）
	if len(source) <= 8 {
		return "", false
	}

	// 4. 读取产品名称（DMI）用于验证
	pname, err := fromIDFile("/sys/class/dmi/id/product_name")
	if err != nil {
		return "", false
	}

	// 5. 验证产品名称是否有效
	if isInvalidProductName(pname) {
		return "", false
	}

	// 6. 基于多个硬件信息源生成唯一 ID（使用 SHA1）
	id := uuid.NewSHA1(uuid.NameSpaceOID, source)
	return id.String(), true
}

// GenerateIDFromMachineID 基于 machine-id 生成 Agent ID（回退方案）
func GenerateIDFromMachineID(workingDir string) string {
	// 1. 尝试读取系统 machine-id 文件（标准 UUID 格式）
	mid, err := fromUUIDFile("/etc/machine-id")
	if err == nil {
		return mid.String()
	}

	// 2. 如果格式不符合标准，基于文件内容生成 ID
	source, err2 := fromIDFile("/etc/machine-id")
	if err2 == nil {
		id := uuid.NewSHA1(uuid.NameSpaceOID, source)
		return id.String()
	}

	// 3. 尝试读取本地持久化的 machine-id 文件
	if workingDir != "" {
		localMachineID := workingDir + "/machine-id"
		mid, err = fromUUIDFile(localMachineID)
		if err == nil {
			return mid.String()
		}
	}

	// 4. 最后回退：生成全新的 UUID
	return uuid.New().String()
}

// PersistID 将 Agent ID 持久化到文件
func PersistID(workingDir, id string) error {
	if workingDir == "" {
		return errors.New("working directory cannot be empty")
	}
	if id == "" {
		return errors.New("agent ID cannot be empty")
	}

	if err := os.MkdirAll(workingDir, 0755); err != nil {
		return err
	}

	idFile := filepath.Join(workingDir, "machine-id")
	if err := os.WriteFile(idFile, []byte(id), 0600); err != nil {
		return err
	}

	return nil
}
