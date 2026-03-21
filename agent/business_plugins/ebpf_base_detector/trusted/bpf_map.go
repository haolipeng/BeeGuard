package trusted

import (
	"fmt"
	"unsafe"

	"github.com/cilium/ebpf"
	"ebpf_base_detector/log"
)

// PopulateTrustedExesMap 将可信任可执行文件写入 BPF map
func PopulateTrustedExesMap(trustedExesMap *ebpf.Map, config *TrustedConfig, logger *log.Logger) (int, error) {
	if !config.Enabled {
		logger.Info("Trusted executable filtering disabled by configuration")
		return 0, nil
	}

	if len(config.TrustedExecutables) == 0 {
		logger.Info("No trusted executables configured")
		return 0, nil
	}

	count := 0
	for _, path := range config.TrustedExecutables {
		// 计算哈希 (map key)
		hash := HashExePath(path)

		// 创建 ExeItem value
		item := ExeItem{
			Len:  int32(len(path)),
			Sid:  0, // 预留字段
			Hash: hash,
		}

		// 复制路径到 Name 字段
		copy(item.Name[:], []byte(path))

		// 写入 BPF map
		if err := trustedExesMap.Put(unsafe.Pointer(&hash), unsafe.Pointer(&item)); err != nil {
			logger.Warn("Failed to add trusted executable to BPF map",
				"path", path,
				"hash", fmt.Sprintf("0x%x", hash),
				"error", err)
			continue // 部分失败可接受,继续
		}

		logger.Debug("Added trusted executable to BPF map",
			"path", path,
			"hash", fmt.Sprintf("0x%x", hash))

		count++
	}

	logger.Info("Populated trusted executables map",
		"total_configured", len(config.TrustedExecutables),
		"successfully_added", count)

	return count, nil
}
