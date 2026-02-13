package trusted

// MurmurOAAT64 实现 Murmur OAAT64 哈希算法
// 必须与 C 版本 (hids.bpf.c) 保持字节级兼容
func MurmurOAAT64(s string, length int) uint64 {
	const (
		seed       = uint64(525201411107845655)
		multiplier = uint64(0x5bd1e9955bd1e995)
		maxLen     = 256 // 与 BPF 验证器限制一致
	)

	// 限制长度以满足 BPF 验证器要求
	if length > maxLen {
		length = maxLen
	}

	h := seed
	for i := 0; i < length && i < maxLen; i++ {
		h ^= uint64(s[i])
		h *= multiplier
		h ^= h >> 47
	}

	return h
}

// HashExePath 计算可执行文件路径的哈希 (不包含 \0)
func HashExePath(path string) uint64 {
	return MurmurOAAT64(path, len(path))
}
