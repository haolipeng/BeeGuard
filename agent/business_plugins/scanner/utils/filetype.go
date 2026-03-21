package utils

import (
	"os"

	"github.com/h2non/filetype"
)

// FileTypeCategory 文件类型分类
type FileTypeCategory string

const (
	FileTypeVideo   FileTypeCategory = "video"
	FileTypeAudio   FileTypeCategory = "audio"
	FileTypeImage   FileTypeCategory = "image"
	FileTypeUnknown FileTypeCategory = "unknown"
)

// headerSize 用于魔数检测的文件头读取大小
const headerSize = 262

// DetectFileType 使用魔数检测文件类型
func DetectFileType(path string) (FileTypeCategory, error) {
	f, err := os.Open(path)
	if err != nil {
		return FileTypeUnknown, err
	}
	defer f.Close()

	header := make([]byte, headerSize)
	n, err := f.Read(header)
	if err != nil {
		return FileTypeUnknown, err
	}

	kind, _ := filetype.Match(header[:n])
	if kind == filetype.Unknown {
		return FileTypeUnknown, nil
	}

	switch {
	case filetype.IsVideo(header[:n]):
		return FileTypeVideo, nil
	case filetype.IsAudio(header[:n]):
		return FileTypeAudio, nil
	case filetype.IsImage(header[:n]):
		return FileTypeImage, nil
	default:
		return FileTypeUnknown, nil
	}
}

// ShouldSkipFileType 判断文件类型是否在跳过列表中
func ShouldSkipFileType(path string, skipTypes []string) (bool, error) {
	if len(skipTypes) == 0 {
		return false, nil
	}

	category, err := DetectFileType(path)
	if err != nil {
		return false, err
	}

	for _, t := range skipTypes {
		if FileTypeCategory(t) == category {
			return true, nil
		}
	}

	return false, nil
}
