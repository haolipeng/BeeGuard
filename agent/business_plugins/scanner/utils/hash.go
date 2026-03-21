package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// FileHash 文件哈希结果
type FileHash struct {
	MD5    string
	SHA256 string
}

// CalcFileHash 计算文件的 MD5 和 SHA256
func CalcFileHash(path string) (*FileHash, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file %s: %w", path, err)
	}
	defer f.Close()

	md5Hash := md5.New()
	sha256Hash := sha256.New()
	writer := io.MultiWriter(md5Hash, sha256Hash)

	if _, err := io.Copy(writer, f); err != nil {
		return nil, fmt.Errorf("read file %s: %w", path, err)
	}

	return &FileHash{
		MD5:    hex.EncodeToString(md5Hash.Sum(nil)),
		SHA256: hex.EncodeToString(sha256Hash.Sum(nil)),
	}, nil
}

// CalcSHA256 计算文件的 SHA256
func CalcSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open file %s: %w", path, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("read file %s: %w", path, err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
