package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

// Filter 三层文件过滤器
type Filter struct {
	pathWhitelist []string
	skipFileTypes []string
	minFileSize   int64
	maxFileSize   int64
}

// NewFilter 创建文件过滤器
func NewFilter(pathWhitelist []string, skipFileTypes []string, minFileSize, maxFileSize int64) *Filter {
	return &Filter{
		pathWhitelist: pathWhitelist,
		skipFileTypes: skipFileTypes,
		minFileSize:   minFileSize,
		maxFileSize:   maxFileSize,
	}
}

// ShouldScan 判断文件是否需要扫描
// 返回 true 表示需要扫描，false 表示应跳过
func (f *Filter) ShouldScan(path string) (bool, string) {
	// 第一层：路径白名单检查
	if f.isWhitelisted(path) {
		return false, "whitelisted path"
	}

	// 跳过符号链接
	info, err := os.Lstat(path)
	if err != nil {
		return false, "stat failed"
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return false, "symlink"
	}

	// 不是普通文件
	if !info.Mode().IsRegular() {
		return false, "not regular file"
	}

	// 第二层：文件大小检查
	size := info.Size()
	if size < f.minFileSize {
		return false, "file too small"
	}
	if size > f.maxFileSize {
		return false, "file too large"
	}

	return true, ""
}

// ShouldScanWithMagic 完整过滤检查（包含魔数检测，I/O 开销较大）
func (f *Filter) ShouldScanWithMagic(path string) (bool, string) {
	ok, reason := f.ShouldScan(path)
	if !ok {
		return false, reason
	}

	// 第三层：魔数文件类型检测
	if len(f.skipFileTypes) > 0 {
		skip, err := shouldSkipByMagic(path, f.skipFileTypes)
		if err == nil && skip {
			return false, "skipped file type"
		}
	}

	return true, ""
}

// isWhitelisted 检查路径是否在白名单中
func (f *Filter) isWhitelisted(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	for _, w := range f.pathWhitelist {
		if absPath == w || strings.HasPrefix(absPath, w+"/") {
			return true
		}
	}
	return false
}

// IsPathWhitelisted 检查路径是否在白名单中（公开方法，供目录遍历使用）
func (f *Filter) IsPathWhitelisted(path string) bool {
	return f.isWhitelisted(path)
}

// shouldSkipByMagic 通过魔数检测判断是否需要跳过
func shouldSkipByMagic(path string, skipTypes []string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	header := make([]byte, 262)
	n, err := f.Read(header)
	if err != nil {
		return false, err
	}
	if n < 4 {
		return false, nil
	}

	// 简单的魔数检测
	category := detectCategory(header[:n])
	for _, t := range skipTypes {
		if t == category {
			return true, nil
		}
	}
	return false, nil
}

// detectCategory 通过文件头魔数判断类别
func detectCategory(header []byte) string {
	if len(header) < 4 {
		return ""
	}

	// JPEG
	if header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF {
		return "image"
	}
	// PNG
	if header[0] == 0x89 && header[1] == 0x50 && header[2] == 0x4E && header[3] == 0x47 {
		return "image"
	}
	// GIF
	if header[0] == 0x47 && header[1] == 0x49 && header[2] == 0x46 {
		return "image"
	}
	// BMP
	if header[0] == 0x42 && header[1] == 0x4D {
		return "image"
	}
	// WEBP
	if len(header) >= 12 && header[0] == 0x52 && header[1] == 0x49 && header[2] == 0x46 && header[3] == 0x46 &&
		header[8] == 0x57 && header[9] == 0x45 && header[10] == 0x42 && header[11] == 0x50 {
		return "image"
	}

	// MP4 / MOV
	if len(header) >= 8 {
		ftyp := string(header[4:8])
		if ftyp == "ftyp" {
			return "video"
		}
	}
	// AVI
	if len(header) >= 12 && header[0] == 0x52 && header[1] == 0x49 && header[2] == 0x46 && header[3] == 0x46 &&
		header[8] == 0x41 && header[9] == 0x56 && header[10] == 0x49 && header[11] == 0x20 {
		return "video"
	}
	// MKV / WebM
	if header[0] == 0x1A && header[1] == 0x45 && header[2] == 0xDF && header[3] == 0xA3 {
		return "video"
	}

	// MP3
	if (header[0] == 0xFF && (header[1]&0xE0) == 0xE0) || (header[0] == 0x49 && header[1] == 0x44 && header[2] == 0x33) {
		return "audio"
	}
	// FLAC
	if header[0] == 0x66 && header[1] == 0x4C && header[2] == 0x61 && header[3] == 0x43 {
		return "audio"
	}
	// OGG
	if header[0] == 0x4F && header[1] == 0x67 && header[2] == 0x67 && header[3] == 0x53 {
		return "audio"
	}
	// WAV
	if len(header) >= 12 && header[0] == 0x52 && header[1] == 0x49 && header[2] == 0x46 && header[3] == 0x46 &&
		header[8] == 0x57 && header[9] == 0x41 && header[10] == 0x56 && header[11] == 0x45 {
		return "audio"
	}

	return ""
}
